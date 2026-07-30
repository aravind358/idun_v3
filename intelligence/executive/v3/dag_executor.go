package v3

import (
	"context"
	"fmt"
	"idun/intelligence/planning/v3"
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
				go e.executeNode(ctx, node, &wg, doneQueue, errorChan, results, &mu)
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

func (e *DAGExecutor) executeNode(ctx context.Context, node v3.PlanNode, wg *sync.WaitGroup, done chan<- string, errChan chan<- error, results map[string]NodeResult, mu *sync.Mutex) {
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

	// 2. Execute Physical Capability
	payload, err := executor.Execute(ctx, node.BoundParams())
	duration := time.Since(start)

	var outputRef string
	var execErr error

	if err != nil {
		execErr = err
	} else if len(payload) > 0 {
		// 3. Centralized CAS Storage
		// Wait, if MemoryProvider fails?
		ref, memErr := e.memory.StorePayload(ctx, payload)
		if memErr != nil {
			execErr = fmt.Errorf("failed to store payload in CAS: %w", memErr)
		} else {
			outputRef = ref
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
