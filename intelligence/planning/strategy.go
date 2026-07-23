package planning

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync/atomic"
)

// ============================================================================
// StrategyProvider Implementation
// ============================================================================

// DefaultStrategyProvider implements StrategyProvider using lock-free atomic pointers.
//
// Architectural Invariant: Planning reads passively via ActiveSnapshot().
// Planning never mutates policy weights or switches profiles. Learning remains sole publisher.
type DefaultStrategyProvider struct {
	snapshot atomic.Pointer[PlanningStrategySnapshot]
}

// NewDefaultStrategyProvider initializes a strategy provider. If nil is passed,
// it initializes a baseline default snapshot (`PROFILE_GENERAL_BASE`).
func NewDefaultStrategyProvider(initial *PlanningStrategySnapshot) *DefaultStrategyProvider {
	p := &DefaultStrategyProvider{}
	if initial == nil {
		defProfile := DefaultPlanningPolicyProfile()
		snap, _ := NewPlanningStrategySnapshot("snap-default-boot", defProfile.ProfileVersion, defProfile)
		p.snapshot.Store(snap)
	} else {
		p.snapshot.Store(initial)
	}
	return p
}

// ActiveSnapshot returns the currently active strategy snapshot via lock-free load.
func (p *DefaultStrategyProvider) ActiveSnapshot() *PlanningStrategySnapshot {
	return p.snapshot.Load()
}

// UpdateSnapshot atomically replaces the active snapshot when published by Learning.
func (p *DefaultStrategyProvider) UpdateSnapshot(newSnapshot *PlanningStrategySnapshot) error {
	if newSnapshot == nil {
		return errors.New("cannot update to nil PlanningStrategySnapshot")
	}
	if newSnapshot.ActiveProfile() == nil {
		return errors.New("cannot update to snapshot containing nil profile")
	}
	p.snapshot.Store(newSnapshot)
	return nil
}

// ============================================================================
// PlanFingerprinter Implementation
// ============================================================================

// DefaultPlanFingerprinter computes deterministic SHA-256 hashes over structural content ONLY.
//
// Architectural Invariant: To guarantee reproducibility and deduplication accuracy over decades,
// the fingerprint explicitly includes only structural graph elements (`Goal`, `Domain`, `Subgoals`,
// `Dependencies`, `Preconditions`, `Postconditions`, `RequiredResources`), and completely ignores
// variable estimates (`EstimatedCost`, `EstimatedDuration`), timestamps, IDs, and transient statuses.
type DefaultPlanFingerprinter struct{}

// NewDefaultPlanFingerprinter constructs a canonical fingerprint generator.
func NewDefaultPlanFingerprinter() *DefaultPlanFingerprinter {
	return &DefaultPlanFingerprinter{}
}

// structuralPayload captures canonical sorted fields for hashing.
type structuralPayload struct {
	SchemaVersion     string                `json:"schema_version"`
	Domain            string                `json:"domain"`
	PlannerID         string                `json:"planner_id,omitempty"`
	PlannerType       string                `json:"planner_type,omitempty"`
	Goal              string                `json:"goal"`
	Subgoals          []Subgoal             `json:"subgoals"`
	Dependencies      []DependencyEdge      `json:"dependencies"`
	Preconditions     []string              `json:"preconditions"`
	Postconditions    []string              `json:"postconditions"`
	RequiredResources []ResourceRequirement `json:"required_resources"`
}

// ComputeFingerprint returns the deterministic SHA-256 hex digest for the CandidatePlan.
func (f *DefaultPlanFingerprinter) ComputeFingerprint(plan *CandidatePlan) (string, error) {
	if plan == nil {
		return "", errors.New("cannot compute fingerprint for nil CandidatePlan")
	}

	// Make copies of slices and sort them canonically by ID/string to ensure ordering invariant
	subgoals := make([]Subgoal, len(plan.Subgoals))
	copy(subgoals, plan.Subgoals)
	sort.Slice(subgoals, func(i, j int) bool {
		return subgoals[i].SubgoalID < subgoals[j].SubgoalID
	})

	dependencies := make([]DependencyEdge, len(plan.Dependencies))
	copy(dependencies, plan.Dependencies)
	sort.Slice(dependencies, func(i, j int) bool {
		return dependencies[i].EdgeID < dependencies[j].EdgeID
	})

	preconditions := make([]string, len(plan.Preconditions))
	copy(preconditions, plan.Preconditions)
	sort.Strings(preconditions)

	postconditions := make([]string, len(plan.Postconditions))
	copy(postconditions, plan.Postconditions)
	sort.Strings(postconditions)

	resources := make([]ResourceRequirement, len(plan.RequiredResources))
	copy(resources, plan.RequiredResources)
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].ResourceID < resources[j].ResourceID
	})

	payload := structuralPayload{
		SchemaVersion:     plan.SchemaVersion,
		Domain:            plan.Domain,
		PlannerID:         plan.PlannerID,
		PlannerType:       plan.PlannerType,
		Goal:              plan.Goal,
		Subgoals:          subgoals,
		Dependencies:      dependencies,
		Preconditions:     preconditions,
		Postconditions:    postconditions,
		RequiredResources: resources,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal structural payload for fingerprinting: %w", err)
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

