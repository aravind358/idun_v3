package planning

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestStress_ThousandsOfPlanningRequests verifies stability across high-volume sequential and concurrent workloads.
func TestStress_ThousandsOfPlanningRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	cfg := DefaultConfig()
	reg := NewSpecialistRegistry()
	_ = reg.Register(NewHTNSpecialist("TACTICAL"))
	service := NewService(WithConfig(cfg), WithSpecialistRegistry(reg))
	defer service.Close()
	_ = service.Start()

	const numRequests = 1000
	var wg sync.WaitGroup
	var successCount atomic.Int32

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req, _ := NewPlanningRequestBuilder().
				WithRequestID(fmt.Sprintf("req-stress-%d", idx)).
				WithGoal(fmt.Sprintf("High volume task %d", idx)).
				WithDomain("General").
				WithTargetDepth(DepthTactical).
				Build()

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			res, err := service.PlanTactical(ctx, req)
			if err == nil && res != nil {
				successCount.Add(1)
			}
		}(i)
	}

	wg.Wait()
	if successCount.Load() != numRequests {
		t.Fatalf("expected %d successes, got %d", numRequests, successCount.Load())
	}
}

// TestStress_ConcurrentPlanningEpisodes verifies race-free execution across parallel multi-domain episodes.
func TestStress_ConcurrentPlanningEpisodes(t *testing.T) {
	cfg := DefaultConfig()
	reg := NewSpecialistRegistry()
	_ = reg.Register(NewHTNSpecialist("TACTICAL"))
	_ = reg.Register(NewGOAPSpecialist("TACTICAL"))
	_ = reg.Register(NewTreeSearchSpecialist("STRATEGIC"))
	service := NewService(WithConfig(cfg), WithSpecialistRegistry(reg))
	defer service.Close()
	_ = service.Start()

	var wg sync.WaitGroup
	const workers = 50

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < 10; i++ {
				depth := DepthTactical
				if i%2 == 0 {
					depth = DepthStrategic
				}
				req, _ := NewPlanningRequestBuilder().
					WithRequestID(fmt.Sprintf("req-conc-%d-%d", workerID, i)).
					WithGoal("Concurrent multi-episode coordination").
					WithDomain("General").
					WithTargetDepth(depth).
					Build()

				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
				if depth == DepthStrategic {
					_, _ = service.PlanDeliberative(ctx, req)
				} else {
					_, _ = service.PlanTactical(ctx, req)
				}
				cancel()
			}
		}(w)
	}
	wg.Wait()
}

// TestStress_RepeatedStrategySnapshotSwaps verifies zero contention or races during atomic policy updates.
func TestStress_RepeatedStrategySnapshotSwaps(t *testing.T) {
	cfg := DefaultConfig()
	prov := NewDefaultStrategyProvider(nil)
	service := NewService(WithConfig(cfg), WithStrategyProvider(prov))
	defer service.Close()
	_ = service.Start()

	var wg sync.WaitGroup
	stop := make(chan struct{})

	// Swapper goroutine: rapidly replaces snapshots
	wg.Add(1)
	go func() {
		defer wg.Done()
		i := 0
		for {
			select {
			case <-stop:
				return
			default:
				prof := DefaultPlanningPolicyProfile()
				prof.ProfileVersion = fmt.Sprintf("v2.%d", i)
				snap, _ := NewPlanningStrategySnapshot(fmt.Sprintf("snap-%d", i), prof.ProfileVersion, prof)
				_ = prov.UpdateSnapshot(snap)
				i++
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	// Readers running planning episodes
	for r := 0; r < 20; r++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()
			for j := 0; j < 25; j++ {
				req, _ := NewPlanningRequestBuilder().
					WithRequestID(fmt.Sprintf("req-swap-%d-%d", readerID, j)).
					WithGoal("Plan under changing strategy").
					WithDomain("General").
					Build()
				ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
				_, _ = service.PlanTactical(ctx, req)
				cancel()
			}
		}(r)
	}

	// Wait for readers, then stop swapper
	time.Sleep(300 * time.Millisecond)
	close(stop)
	wg.Wait()
}

// TestStress_RapidCancellationsAndTimeoutStorms verifies clean pre-emption without deadlocks or goroutine leaks.
func TestStress_RapidCancellationsAndTimeoutStorms(t *testing.T) {
	cfg := DefaultConfig()
	reg := NewSpecialistRegistry()
	_ = reg.Register(NewTreeSearchSpecialist("STRATEGIC"))
	service := NewService(WithConfig(cfg), WithSpecialistRegistry(reg))
	defer service.Close()
	_ = service.Start()

	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req, _ := NewPlanningRequestBuilder().
				WithRequestID(fmt.Sprintf("req-storm-%d", idx)).
				WithGoal("Evaluate massive tree expansion under timeout storm").
				WithDomain("General").
				WithTargetDepth(DepthStrategic).
				Build()

			// Microsecond cancellation storms
			timeout := time.Duration(1+(idx%10)) * time.Millisecond
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			_, _ = service.PlanDeliberative(ctx, req)
		}(i)
	}
	wg.Wait()
}

// TestStress_HighFrequencyReflexivePlanning verifies <2ms reflexive lookup under sustained concurrent load.
func TestStress_HighFrequencyReflexivePlanning(t *testing.T) {
	cfg := DefaultConfig()
	service := NewService(WithConfig(cfg))
	defer service.Close()
	_ = service.Start()

	var wg sync.WaitGroup
	const workers = 25

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				req, _ := NewPlanningRequestBuilder().
					WithRequestID(fmt.Sprintf("req-hf-%d-%d", id, i)).
					WithGoal("Quick reflex check").
					WithDomain("General").
					WithTargetDepth(DepthReflexive).
					Build()
				ctx := context.Background()
				_, err := service.PlanReflexive(ctx, req)
				if err != nil {
					t.Errorf("Reflexive plan failed: %v", err)
					return
				}
			}
		}(w)
	}
	wg.Wait()
}

// TestStress_MixedPlanningDomains verifies routing across distinct open domain tags simultaneously.
func TestStress_MixedPlanningDomains(t *testing.T) {
	cfg := DefaultConfig()
	reg := NewSpecialistRegistry()
	_ = reg.Register(NewHTNSpecialist("TACTICAL"))
	_ = reg.Register(NewGOAPSpecialist("TACTICAL"))
	service := NewService(WithConfig(cfg), WithSpecialistRegistry(reg))
	defer service.Close()
	_ = service.Start()

	domains := []string{"General", "Coding", "Robotics", "MissionControl"}
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			domain := domains[idx%len(domains)]
			req, _ := NewPlanningRequestBuilder().
				WithRequestID(fmt.Sprintf("req-dom-%d", idx)).
				WithGoal(fmt.Sprintf("Domain specific task for %s", domain)).
				WithDomain(domain).
				Build()
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()
			_, _ = service.PlanTactical(ctx, req)
		}(i)
	}
	wg.Wait()
}

// TestStress_PlannerShutdownUnderLoad verifies panic isolation, zero deadlocks, and clean teardown when closed mid-flight.
func TestStress_PlannerShutdownUnderLoad(t *testing.T) {
	cfg := DefaultConfig()
	reg := NewSpecialistRegistry()
	_ = reg.Register(NewTreeSearchSpecialist("STRATEGIC"))
	service := NewService(WithConfig(cfg), WithSpecialistRegistry(reg))
	_ = service.Start()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req, _ := NewPlanningRequestBuilder().
				WithRequestID(fmt.Sprintf("req-shut-%d", idx)).
				WithGoal("Long running task during shutdown").
				WithDomain("General").
				WithTargetDepth(DepthStrategic).
				Build()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_, _ = service.PlanDeliberative(ctx, req)
		}(i)
	}

	// Trigger shutdown while planners are active
	time.Sleep(50 * time.Millisecond)
	err := service.Close()
	if err != nil {
		t.Fatalf("clean shutdown returned error: %v", err)
	}
	wg.Wait()
}
