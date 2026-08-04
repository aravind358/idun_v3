package v3

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"idun/capabilities"
	"idun/core/foundation"
	"idun/core/storage"
	planningv3 "idun/intelligence/planning/v3"
)

// LegacyCapabilityExecutor wraps a core capability.
type LegacyCapabilityExecutor struct {
	cap capabilities.Capability
}

func (e *LegacyCapabilityExecutor) Execute(ctx context.Context, params map[string]any) ([]byte, error) {
	strParams := make(map[string]string)
	for k, v := range params {
		if k == "operand1" { k = "a" }
		if k == "operand2" { k = "b" }
		if k == "operator" { 
			k = "operation" 
			switch fmt.Sprintf("%v", v) {
			case "+": v = "add"
			case "-": v = "subtract"
			case "*": v = "multiply"
			case "/": v = "divide"
			case "%": v = "modulo"
			}
		}
		if k == "task" && string(e.cap.ID()) == "app-rem-1" { k = "message" }
		if k == "operation" {
			val := strings.ToLower(fmt.Sprintf("%v", v))
			if string(e.cap.ID()) == "app-rem-1" && (val == "remind" || val == "create") {
				v = "set"
			}
			
			if string(e.cap.ID()) == "app-notes-1" {
				switch val {
				case "take", "save", "set", "note":
					v = "create"
				case "open":
					v = "read"
				case "show":
					if _, hasTitle := params["title"]; hasTitle {
						v = "read"
					} else {
						v = "list"
					}
				case "remove":
					v = "delete"
				case "what":
					v = "list"
				}
			}
		}
		strParams[k] = fmt.Sprintf("%v", v)
	}

	// Default Note Naming Rule Implementation
	if string(e.cap.ID()) == "app-notes-1" {
		op := strings.ToLower(strParams["operation"])
		if op == "create" {
			if _, hasTitle := strParams["title"]; !hasTitle || strParams["title"] == "" {
				// Deterministic collision-safe policy for MVP: Use timestamp
				strParams["title"] = fmt.Sprintf("Quick Note %d", time.Now().Unix())
			}
		}
	}

	req := capabilities.CapabilityRequest{
		RequirementID: string(e.cap.ID()),
		Parameters:    strParams,
	}

	res, err := e.cap.Execute(ctx, req)
	if err != nil {
		return nil, err
	}
	// Return the entire CapabilityResult JSON so RealizationStrategy is preserved
	return json.Marshal(res)
}

// LegacyCapabilityRegistryAdapter wraps the core capability manager.
type LegacyCapabilityRegistryAdapter struct {
	manager capabilities.CapabilityManager
}

func NewLegacyCapabilityRegistryAdapter(manager capabilities.CapabilityManager) CapabilityRegistry {
	return &LegacyCapabilityRegistryAdapter{manager: manager}
}

// MockCommunicativeExecutor handles communicative intents in the mock V3 pipeline
type MockCommunicativeExecutor struct {
	intent string
}

func (e *MockCommunicativeExecutor) Execute(ctx context.Context, params map[string]any) ([]byte, error) {
	intent := e.intent
	if pi, ok := ctx.Value("planIntent").(string); ok && pi != "" {
		intent = pi
	}

	res := capabilities.CapabilityResult{
		Realization:  capabilities.Generative,
		ResponseType: "communicative",
		Data: map[string]interface{}{
			"intent": intent,
			"status": "mocked",
		},
	}
	return json.Marshal(res)
}

func (a *LegacyCapabilityRegistryAdapter) Resolve(capabilityID string) (CapabilityExecutor, error) {
	if capabilityID == "mock.capability" {
		return &MockCommunicativeExecutor{intent: "communicative"}, nil
	}

	if a.manager == nil || a.manager.Registry() == nil {
		return nil, fmt.Errorf("capability manager not wired")
	}

	cap, ok := a.manager.Registry().Get(capabilityID)
	if !ok {
		return nil, fmt.Errorf("capability not found: %s", capabilityID)
	}

	return &LegacyCapabilityExecutor{cap: cap}, nil
}


// LegacyMemoryAdapter wraps core storage for Executive V3 MemoryProvider.
type LegacyMemoryAdapter struct {
	store *storage.Storage
}

func NewLegacyMemoryAdapter(store *storage.Storage) MemoryProvider {
	return &LegacyMemoryAdapter{store: store}
}

func (a *LegacyMemoryAdapter) StorePayload(ctx context.Context, payload []byte) (string, error) {
	if a.store == nil {
		return "", fmt.Errorf("storage not wired")
	}
	hash := sha256.Sum256(payload)
	key := fmt.Sprintf("%x", hash)
	if err := a.store.Write(key, payload); err != nil {
		return "", err
	}
	return key, nil
}

// StoragePlanProvider retrieves plans from the Artifact Index.
type StoragePlanProvider struct {
	storer interface {
		LookupArtifact(ctx context.Context, artifactID string) (storage.ArtifactIndexMeta, error)
		Retrieve(ctx context.Context, key string) ([]byte, error)
	}
}

func NewStoragePlanProvider(storer interface {
	LookupArtifact(ctx context.Context, artifactID string) (storage.ArtifactIndexMeta, error)
	Retrieve(ctx context.Context, key string) ([]byte, error)
}) *StoragePlanProvider {
	return &StoragePlanProvider{storer: storer}
}

func (s *StoragePlanProvider) GetPlan(ctx context.Context, planID foundation.ArtifactID) (*planningv3.ExecutionPlan, error) {
	meta, err := s.storer.LookupArtifact(ctx, string(planID))
	if err != nil {
		return nil, fmt.Errorf("executivev3: failed to lookup plan artifact %s: %w", planID, err)
	}

	data, err := s.storer.Retrieve(ctx, meta.PayloadRef)
	if err != nil {
		return nil, fmt.Errorf("executivev3: failed to retrieve plan payload %s: %w", meta.PayloadRef, err)
	}

	var plan planningv3.ExecutionPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("executivev3: failed to unmarshal plan %s: %w", planID, err)
	}

	return &plan, nil
}
