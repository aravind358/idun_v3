package executive

import (
	"fmt"
)

// ExecutiveResultStatus records what occurred at the service output level during coordination,
// cleanly separated from why the coordination terminated (ExecutiveTerminationReason).
type ExecutiveResultStatus string

const (
	// StatusSuccess indicates coordination completed successfully across all required transitions.
	StatusSuccess ExecutiveResultStatus = "SUCCESS"

	// StatusPartial indicates partial execution where some branches executed before interruption or limits.
	StatusPartial ExecutiveResultStatus = "PARTIAL"

	// StatusWaiting indicates coordination suspended cooperatively waiting for external input or time event.
	StatusWaiting ExecutiveResultStatus = "WAITING"

	// StatusFailed indicates terminal coordination failure due to dependency errors or contradictions.
	StatusFailed ExecutiveResultStatus = "FAILED"

	// StatusAborted indicates explicit executive cancellation or preemption.
	StatusAborted ExecutiveResultStatus = "ABORTED"

	// StatusInterrupted indicates immediate preemption by Band 0 Critical Safety or Band 1 RealTime task.
	StatusInterrupted ExecutiveResultStatus = "INTERRUPTED"

	// StatusNoAction indicates coordination abstained or required zero actions.
	StatusNoAction ExecutiveResultStatus = "NO_ACTION"
)

// String returns the canonical string representation of ExecutiveResultStatus.
func (s ExecutiveResultStatus) String() string {
	return string(s)
}

// Validate checks whether s is a recognized ExecutiveResultStatus enum value.
func (s ExecutiveResultStatus) Validate() error {
	switch s {
	case StatusSuccess, StatusPartial, StatusWaiting, StatusFailed, StatusAborted, StatusInterrupted, StatusNoAction:
		return nil
	default:
		return fmt.Errorf("%w: unknown result status %q", ErrInvalidResult, s)
	}
}

// ExecutiveTerminationReason records the factual reason why coordination execution stopped.
type ExecutiveTerminationReason string

const (
	// ReasonSuccess indicates natural completion of the coordination workflow.
	ReasonSuccess ExecutiveTerminationReason = "SUCCESS"

	// ReasonUserCancelled indicates explicit cancellation triggered by user request.
	ReasonUserCancelled ExecutiveTerminationReason = "USER_CANCELLED"

	// ReasonTimeBudgetExceeded indicates execution halted due to time or execution fuel budget exhaustion.
	ReasonTimeBudgetExceeded ExecutiveTerminationReason = "TIME_BUDGET_EXCEEDED"

	// ReasonDependencyFailure indicates failure in an underlying cognitive ability driver.
	ReasonDependencyFailure ExecutiveTerminationReason = "DEPENDENCY_FAILURE"

	// ReasonConstitutionBlock indicates immediate veto by the Constitutional Action Gate.
	ReasonConstitutionBlock ExecutiveTerminationReason = "CONSTITUTION_BLOCK"

	// ReasonResourceExhausted indicates starvation of cycle budget units or concurrency slots.
	ReasonResourceExhausted ExecutiveTerminationReason = "RESOURCE_EXHAUSTED"

	// ReasonSystemShutdown indicates shutdown of the underlying Executive or Kernel engine.
	ReasonSystemShutdown ExecutiveTerminationReason = "SYSTEM_SHUTDOWN"

	// ReasonInterrupted indicates preemption by a higher-priority task band.
	ReasonInterrupted ExecutiveTerminationReason = "INTERRUPTED"

	// ReasonExecutiveAbort indicates an internal executive safety abort or unresolvable contradiction.
	ReasonExecutiveAbort ExecutiveTerminationReason = "EXECUTIVE_ABORT"
)

// String returns the canonical string representation of ExecutiveTerminationReason.
func (r ExecutiveTerminationReason) String() string {
	return string(r)
}

// Validate checks whether r is a recognized ExecutiveTerminationReason enum value.
func (r ExecutiveTerminationReason) Validate() error {
	switch r {
	case ReasonSuccess, ReasonUserCancelled, ReasonTimeBudgetExceeded, ReasonDependencyFailure,
		ReasonConstitutionBlock, ReasonResourceExhausted, ReasonSystemShutdown, ReasonInterrupted, ReasonExecutiveAbort:
		return nil
	default:
		return fmt.Errorf("%w: unknown termination reason %q", ErrInvalidResult, r)
	}
}
