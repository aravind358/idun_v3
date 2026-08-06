package v3

import (
	"context"
	"encoding/json"
	"fmt"
	"idun/intelligence/planning/v3"
	"os"
	"sync"
	"time"
)

// DAGExecutor traverses and executes a planning.ExecutionPlan DAG concurrently.
type DAGExecutor struct {
	registry CapabilityRegistry
	memory   MemoryProvider
}

// NewDAGExecutor creates a new DAGExecutor.
func NewDAGExecutor(registry CapabilityRegistry, memory MemoryProvider) *DAGExecutor {
	return &DAGExecutor{
		registry: registry,
		memory:   memory,
	}
}

// Execute traverses the plan, invoking capabilities concurrently while respecting dependencies.
// It implements Graceful Degradation: if a node fails, it stops scheduling new nodes
// but allows currently executing sibling nodes to finish cleanly.
func (e *DAGExecutor) Execute(ctx context.Context, plan *v3.ExecutionPlan) (map[string]NodeResult, ExecutionStatus, error) {
	nodes := plan.Nodes()
	edges := plan.Edges()

	inDegree := make(map[string]int)
	adjList := make(map[string][]string)
	nodeMap := make(map[string]v3.PlanNode)

	for _, n := range nodes {
		inDegree[n.NodeID()] = 0
		nodeMap[n.NodeID()] = n
	}

	for _, edge := range edges {
		inDegree[edge.ToNodeID()]++
		adjList[edge.FromNodeID()] = append(adjList[edge.FromNodeID()], edge.ToNodeID())
	}

	results := make(map[string]NodeResult)
	var mu sync.Mutex

	// Channels for orchestrating execution
	readyQueue := make(chan v3.PlanNode, len(nodes))
	doneQueue := make(chan string, len(nodes))
	errorChan := make(chan error, len(nodes))

	// Queue initial root nodes
	for id, deg := range inDegree {
		if deg == 0 {
			readyQueue <- nodeMap[id]
		}
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Inject PlanIntent for placeholder capabilities that need original context
	ctx = context.WithValue(ctx, "planIntent", plan.PlanIntent())

	completedCount := 0
	hasFailed := false
	var impasseErr error

	// Worker dispatcher
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case node := <-readyQueue:
				// If the DAG has failed globally, skip scheduling new nodes (Graceful Degradation)
				mu.Lock()
				failed := hasFailed
				mu.Unlock()

				if failed {
					mu.Lock()
					results[node.NodeID()] = NodeResult{
						NodeID: node.NodeID(),
						Status: NodeSkipped,
					}
					mu.Unlock()
					doneQueue <- node.NodeID()
					continue
				}

				wg.Add(1)
				go e.executeNode(ctx, plan, node, &wg, doneQueue, errorChan, results, &mu)
			}
		}
	}()

	// Wait loop for completion
	for completedCount < len(nodes) {
		select {
		case <-ctx.Done():
			// Context timed out or was cancelled by the caller (Executive)
			// Wait for running nodes to abort cleanly
			wg.Wait()
			status := StatusCancelled
			if ctx.Err() == context.DeadlineExceeded {
				status = StatusTimedOut
			}
			return results, status, ctx.Err()

		case err := <-errorChan:
			// A node failed or encountered an impasse.
			mu.Lock()
			hasFailed = true
			if err.Error() == "impasse: executor not found" {
				impasseErr = err
			}
			mu.Unlock()
			// We DO NOT cancel the context here. We let active siblings finish.

		case doneNodeID := <-doneQueue:
			completedCount++
			
			// Decrement in-degree for dependents if DAG hasn't failed
			mu.Lock()
			failed := hasFailed
			mu.Unlock()
			
			if !failed {
				for _, neighbor := range adjList[doneNodeID] {
					inDegree[neighbor]--
					if inDegree[neighbor] == 0 {
						readyQueue <- nodeMap[neighbor]
					}
				}
			} else {
				// If failed, just mark dependents as skipped and push to doneQueue to clear them
				for _, neighbor := range adjList[doneNodeID] {
					if inDegree[neighbor] > 0 { // prevent double queuing
						inDegree[neighbor] = -1 // mark as processed
						mu.Lock()
						results[neighbor] = NodeResult{
							NodeID: neighbor,
							Status: NodeSkipped,
						}
						mu.Unlock()
						doneQueue <- neighbor
					}
				}
			}
		}
	}

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if impasseErr != nil {
		return results, StatusImpasse, impasseErr
	}
	if hasFailed {
		return results, StatusFailed, fmt.Errorf("DAG execution failed")
	}

	return results, StatusCompleted, nil
}

func (e *DAGExecutor) executeNode(ctx context.Context, plan *v3.ExecutionPlan, node v3.PlanNode, wg *sync.WaitGroup, done chan<- string, errChan chan<- error, results map[string]NodeResult, mu *sync.Mutex) {
	defer wg.Done()

	start := time.Now()
	capID := string(node.Capability())

	// 1. Resolve Capability
	executor, err := e.registry.Resolve(capID)
	if err != nil {
		mu.Lock()
		results[node.NodeID()] = NodeResult{
			NodeID:   node.NodeID(),
			Status:   NodeFailed,
			Duration: time.Since(start),
			Error:    fmt.Sprintf("failed to resolve capability: %v", err),
		}
		mu.Unlock()
		errChan <- fmt.Errorf("impasse: executor not found")
		done <- node.NodeID()
		return
	}

	// Inject PlanIntent if missing
	params := node.BoundParams()
	if _, ok := params["intent"]; !ok {
		if plan != nil && plan.PlanIntent() != "" {
			params["intent"] = plan.PlanIntent()
		}
	}

	// 2. Execute Physical Capability
	fmt.Printf("\n[Executive Trace] Executing Capability %q with parameters:\n", capID)
	for k, v := range params {
		fmt.Printf("    %s = %v\n", k, v)
	}
	fmt.Println()
	payload, err := executor.Execute(ctx, params)
	if err != nil {
		fmt.Printf("[Executive Trace] Capability %q failed: %v\n", capID, err)
	} else {
		fmt.Printf("[Executive Trace] Capability %q succeeded. Result: %s\n", capID, payload)
	}
	duration := time.Since(start)

	var outputRef string
	var execErr error

	if err != nil {
		execErr = err
	} else if len(payload) > 0 {
		// 3. Centralized CAS Storage
		ref, memErr := e.memory.StorePayload(ctx, payload)
		if memErr != nil {
			execErr = fmt.Errorf("failed to store payload in CAS: %w", memErr)
		} else {
			outputRef = ref
		}
		
		// TODO: Architectural Debt
		// This is a temporary testing hook for the Runtime Acceptance Harness.
		// The Executive should not have knowledge of specific testing environments.
		// In a future phase, this should be replaced by an Acceptance Observer that 
		// subscribes to ExecutionCompleted events, completely removing this logic 
		// from the execution pipeline.
		if os.Getenv("IDUN_ACCEPTANCE_TEST") == "1" {
			type capRes struct {
				Operation    string `json:"operation"`
				ResponseType string `json:"response_type"`
				Data         struct {
					Intent string `json:"intent"`
				} `json:"data"`
			}
			var cr capRes
			if json.Unmarshal(payload, &cr) == nil {
				op := cr.Operation
				if op == "" && cr.Data.Intent != "" {
					op = cr.Data.Intent
				}

				type execMeta struct {
					Capability   string `json:"capability"`
					Operation    string `json:"operation"`
					ResponseType string `json:"response_type"`
				}
				metaBytes, _ := json.Marshal(execMeta{
					Capability:   capID,
					Operation:    op,
					ResponseType: cr.ResponseType,
				})
				fmt.Printf("\n[Acceptance Metadata] %s\n", metaBytes)
			}
		}
	}

	// 4. Record Result
	mu.Lock()
	res := NodeResult{
		NodeID:    node.NodeID(),
		Duration:  duration,
		OutputRef: outputRef,
	}
	if execErr != nil {
		res.Status = NodeFailed
		res.Error = execErr.Error()
	} else {
		res.Status = NodeCompleted
	}
	results[node.NodeID()] = res
	mu.Unlock()

	if execErr != nil {
		errChan <- execErr
	}
	done <- node.NodeID()
}
