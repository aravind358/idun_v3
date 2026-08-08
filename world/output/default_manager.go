package output

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"idun/capabilities"
	"idun/intelligence/executive/v3"
	"idun/world"
)

// PayloadRetriever retrieves raw bytes from the content-addressed store.
// OutputManager uses this to resolve capability payloads referenced by NodeResult.OutputRef.
type PayloadRetriever interface {
	Retrieve(ctx context.Context, key string) ([]byte, error)
}

// DefaultOutputManager implements OutputManager.
type DefaultOutputManager struct {
	registry   PluginRegistry
	aggregator Aggregator
	engine     OutputEngine
	retriever  PayloadRetriever
}

// NewDefaultOutputManager creates a new DefaultOutputManager.
// retriever may be nil; when nil, OutputRef payloads will not be resolved and
// the realized content will fall back to an execution summary.
func NewDefaultOutputManager(registry PluginRegistry, aggregator Aggregator, engine OutputEngine, retriever PayloadRetriever) *DefaultOutputManager {
	return &DefaultOutputManager{
		registry:   registry,
		aggregator: aggregator,
		engine:     engine,
		retriever:  retriever,
	}
}

// Dispatch begins the asynchronous output process for a given ExecutionResult.
// It is guaranteed not to block the caller — all work executes in a background goroutine.
func (m *DefaultOutputManager) Dispatch(ctx context.Context, interaction *world.Interaction, execResult *v3.ExecutionResult) error {
	if m.registry == nil {
		return fmt.Errorf("world/output: plugin registry is not initialized")
	}

	// The Workspace event loop must never wait for output realization.
	// We dispatch the work asynchronously.
	go func() {
		// Use a background context so the pipeline continues even if the original
		// request context times out.
		bgCtx := context.Background()
		m.process(bgCtx, interaction, execResult)
	}()

	return nil
}

// resolveCapabilityText extracts human-readable text from a raw capability result payload.
// It follows the same decoding contract that Presentation.Router uses, maintaining
// behavioral parity during the O4 cutover.
func resolveCapabilityText(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}

	// Attempt to decode as a CapabilityResult (the canonical capability payload format).
	var capResult capabilities.CapabilityResult
	if err := json.Unmarshal(raw, &capResult); err == nil {
		if !capResult.Success && capResult.Error != nil {
			return fmt.Sprintf("I wasn't able to complete that: %s", capResult.Error.Message)
		}
		// Extract display text from Data map (convention: key "response" or "text").
		if capResult.Data != nil {
			for _, key := range []string{"response", "text", "message", "result", "output"} {
				if v, ok := capResult.Data[key]; ok {
					if s, ok := v.(string); ok && s != "" {
						return s
					}
				}
			}
			// Fallback: build a brief summary from the Data map.
			var parts []string
			for k, v := range capResult.Data {
				parts = append(parts, fmt.Sprintf("%s: %v", k, v))
			}
			if len(parts) > 0 {
				return strings.Join(parts, ", ")
			}
		}
		if capResult.Operation != "" {
			return fmt.Sprintf("Completed: %s", capResult.Operation)
		}
	}

	// Attempt to decode as a RealizedOutput (in case an LLM realizer already ran).
	var realized struct {
		RealizedText string `json:"realized_text"`
	}
	if err := json.Unmarshal(raw, &realized); err == nil && realized.RealizedText != "" {
		return realized.RealizedText
	}

	// Raw text fallback (e.g. plain string payload from a test fixture or legacy node).
	return string(raw)
}

// process is the synchronous core of the output pipeline, run inside a goroutine by Dispatch.
// Responsibility split (per O4 architectural rules):
//   1. OutputManager  — retrieves CAS payloads, builds ResolvedContent
//   2. OutputEngine   — transforms CompositeResponse → OutputDocument (no CAS access)
//   3. OutputPlugin   — formats and writes the OutputDocument (modality-specific I/O)
func (m *DefaultOutputManager) process(ctx context.Context, interaction *world.Interaction, execResult *v3.ExecutionResult) {
	// 1. Aggregate — sort NodeResults by GoalIndex for deterministic semantic ordering.
	compositeResp, err := m.aggregator.Aggregate(ctx, execResult)
	if err != nil {
		fmt.Printf("[OutputManager] Error aggregating node results: %v\n", err)
		return
	}

	// 2. Resolve CAS payloads — OutputManager is the only layer that touches CAS.
	//    We iterate in semantic goal order (OrderedNodeResults, already sorted by GoalIndex).
	if m.retriever != nil && len(compositeResp.OrderedNodeResults) > 0 {
		var segments []string
		for _, nr := range compositeResp.OrderedNodeResults {
			switch nr.Status {
			case v3.NodeCompleted:
				if nr.OutputRef == "" {
					continue
				}
				raw, retrieveErr := m.retriever.Retrieve(ctx, nr.OutputRef)
				if retrieveErr != nil {
					fmt.Printf("[OutputManager] Warning: could not retrieve payload for node %s: %v\n", nr.NodeID, retrieveErr)
					continue
				}
				if text := resolveCapabilityText(raw); text != "" {
					segments = append(segments, text)
				}
			case v3.NodeFailed:
				if nr.Error != "" {
					segments = append(segments, fmt.Sprintf("One action could not be completed: %s", nr.Error))
				}
			case v3.NodeSkipped:
				// Skipped nodes produce no output.
			}
		}
		compositeResp.ResolvedContent = strings.Join(segments, "\n\n")
	}

	// Fallback when there are no segments (e.g. no retriever, or all nodes skipped).
	if compositeResp.ResolvedContent == "" {
		compositeResp.ResolvedContent = fmt.Sprintf("[Execution completed at %s]", time.Now().Format(time.RFC3339))
	}

	// 3. Realize — OutputEngine transforms CompositeResponse → OutputDocument.
	doc, err := m.engine.Realize(ctx, compositeResp)
	if err != nil {
		fmt.Printf("[OutputManager] Error realizing composite response: %v\n", err)
		return
	}

	// Propagate interaction session ID if the engine did not set one.
	if doc.SessionID == "" && interaction != nil {
		doc.SessionID = interaction.SessionID
	}

	// 4. Resolve plugin for the interaction's modality.
	modality := world.ModalityText
	if interaction != nil {
		modality = interaction.Modality
	}

	plugin, err := m.registry.Get(modality)
	if err != nil {
		fmt.Printf("[OutputManager] Error getting plugin for modality '%s': %v\n", modality, err)
		return
	}

	// 5. Format — modality-specific formatting (e.g. terminal colours, Markdown rendering).
	if err := plugin.Format(ctx, doc); err != nil {
		fmt.Printf("[OutputManager] Error formatting document: %v\n", err)
		return
	}

	// 6. Write — physical I/O delivery.
	if err := plugin.Write(ctx, doc); err != nil {
		fmt.Printf("[OutputManager] Error writing document: %v\n", err)
		return
	}
}
