package network

import (
	"sync"
	"time"
)

// CapabilityMetrics tracks execution diagnostics safely.
type CapabilityMetrics struct {
	mu                     sync.RWMutex
	ExecutionCount         int64
	SuccessCount           int64
	FailureCount           int64
	TotalExecutionDuration time.Duration
	LastSuccessfulTime     time.Time
	LastFailedTime         time.Time
}

func NewCapabilityMetrics() *CapabilityMetrics {
	return &CapabilityMetrics{}
}

func (m *CapabilityMetrics) RecordSuccess(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ExecutionCount++
	m.SuccessCount++
	m.TotalExecutionDuration += duration
	m.LastSuccessfulTime = time.Now()
}

func (m *CapabilityMetrics) RecordFailure(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ExecutionCount++
	m.FailureCount++
	m.TotalExecutionDuration += duration
	m.LastFailedTime = time.Now()
}

func (m *CapabilityMetrics) AverageDuration() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.ExecutionCount == 0 {
		return 0
	}
	return m.TotalExecutionDuration / time.Duration(m.ExecutionCount)
}
