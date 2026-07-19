package decision

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"idun/intelligence/communication"
	"idun/intelligence/executive"
)

// DefaultDecisionService implements DecisionService and executive.DecisionAbility.
// It orchestrates Tier 1 Constitutional Hard Gates and Tier 2 Objective Utility Scoring
// across Reflexive and Deliberative surfaces without violating single responsibility.
type DefaultDecisionService struct {
	mu           sync.RWMutex
	started      bool
	tier1Gate    Tier1ConstitutionalGate
	tier2Scorer  Tier2ObjectiveScorer
	strategyProv StrategyProvider

	// traces holds in-memory O(1) episode traces keyed by episode ID
	traces map[string]*ReflexiveDecisionTrace

	storer     PayloadStorer
	publisher  WorkspacePublisher
	subscriber WorkspaceSubscriber
	sub        WorkspaceSubscription
}

// Option configures DefaultDecisionService dependencies.
type Option func(*DefaultDecisionService)

// WithTier1Gate overrides the default Tier 1 Constitutional Gate.
func WithTier1Gate(gate Tier1ConstitutionalGate) Option {
	return func(s *DefaultDecisionService) {
		s.tier1Gate = gate
	}
}

// WithTier2Scorer overrides the default Tier 2 Objective Scorer.
func WithTier2Scorer(scorer Tier2ObjectiveScorer) Option {
	return func(s *DefaultDecisionService) {
		s.tier2Scorer = scorer
	}
}

// WithStrategyProvider overrides the default Strategy Provider.
func WithStrategyProvider(prov StrategyProvider) Option {
	return func(s *DefaultDecisionService) {
		s.strategyProv = prov
	}
}

// WithWorkspaceBridge injects the Workspace substrate for consuming and publishing candidate plans.
func WithWorkspaceBridge(storer PayloadStorer, pub WorkspacePublisher, sub WorkspaceSubscriber) Option {
	return func(s *DefaultDecisionService) {
		s.storer = storer
		s.publisher = pub
		s.subscriber = sub
	}
}

// NewService constructs a DefaultDecisionService with sensible defaults.
func NewService(opts ...Option) *DefaultDecisionService {
	s := &DefaultDecisionService{
		tier1Gate:    NewDefaultConstitutionalGate(),
		tier2Scorer:  NewDefaultObjectiveScorer(),
		strategyProv: NewDefaultStrategyProvider(nil),
		traces:       make(map[string]*ReflexiveDecisionTrace),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Ability returns executive.AbilityDecision.
func (s *DefaultDecisionService) Ability() executive.CognitiveAbility {
	return executive.AbilityDecision
}

// Start boots the Decision service lifecycle.
func (s *DefaultDecisionService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.started {
		return nil
	}

	if s.subscriber != nil && s.sub == nil {
		sub, err := s.subscriber.Subscribe(communication.TopicCandidatePlans, "Intelligence.Decision", s.handleCandidatePlans)
		if err != nil {
			return fmt.Errorf("decision: failed to subscribe to TopicCandidatePlans: %w", err)
		}
		s.sub = sub
	}

	s.started = true
	return nil
}

// Close gracefully shuts down the Decision service.
func (s *DefaultDecisionService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.started = false
	if s.sub != nil {
		_ = s.sub.Cancel()
		s.sub = nil
	}
	return nil
}

// getOrCreateTrace returns or initializes the O(1) ReflexiveDecisionTrace for an episode.
func (s *DefaultDecisionService) getOrCreateTrace(episodeID, strategyVer string) *ReflexiveDecisionTrace {
	s.mu.Lock()
	defer s.mu.Unlock()
	trace, ok := s.traces[episodeID]
	if !ok {
		trace = NewReflexiveDecisionTrace(episodeID, strategyVer)
		s.traces[episodeID] = trace
	}
	return trace
}

// GetEpisodeTrace retrieves the O(1) memory-bounded trace accumulator for an episode.
func (s *DefaultDecisionService) GetEpisodeTrace(episodeID string) (*ReflexiveDecisionTrace, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	trace, ok := s.traces[episodeID]
	return trace, ok
}

type planPayload struct {
	PlanID            string  `json:"plan_id"`
	Goal              string  `json:"goal"`
	Domain            string  `json:"domain"`
	PlannerID         string  `json:"planner_id,omitempty"`
	PlannerType       string  `json:"planner_type,omitempty"`
	EstimatedCost     float64 `json:"estimated_cost"`
	ConfidenceProfile struct {
		OverallConfidence float64 `json:"overall_confidence"`
	} `json:"confidence_profile"`
	SourceTier string `json:"source_tier"`
}

type planningResultPayload struct {
	ResultID string         `json:"result_id"`
	Plans    []*planPayload `json:"plans"`
}

func (s *DefaultDecisionService) handleCandidatePlans(ctx context.Context, env communication.Envelope) error {
	s.mu.RLock()
	if !s.started {
		s.mu.RUnlock()
		return nil
	}
	s.mu.RUnlock()

	if env.ID == "" || env.Topic != communication.TopicCandidatePlans || env.PayloadRef == "" {
		return errors.New("decision: invalid candidate plans envelope")
	}

	devLog("Decision", "Received TopicCandidatePlans")

	if s.storer == nil || s.publisher == nil {
		return errors.New("decision: workspace bridge not fully configured")
	}

	data, err := s.storer.Retrieve(ctx, env.PayloadRef)
	if err != nil {
		return fmt.Errorf("decision: failed to retrieve candidate plans payload: %w", err)
	}

	var resPayload planningResultPayload
	var plans []*planPayload

	if err := json.Unmarshal(data, &resPayload); err == nil && len(resPayload.Plans) > 0 {
		plans = resPayload.Plans
	} else {
		var singlePlan planPayload
		if err := json.Unmarshal(data, &singlePlan); err != nil {
			return fmt.Errorf("decision: failed to parse payload as PlanningResult or Plan: %w", err)
		}
		plans = []*planPayload{&singlePlan}
	}

	cs := CandidateSet{
		EpisodeID:  env.ID,
		Candidates: make([]Candidate, 0, len(plans)),
	}

	for _, p := range plans {
		if p == nil {
			continue
		}
		meta := make(map[string]string)
		if p.PlannerID != "" {
			meta["planner_id"] = p.PlannerID
		}
		if p.PlannerType != "" {
			meta["planner_type"] = p.PlannerType
		}
		if p.SourceTier != "" {
			meta["source_tier"] = p.SourceTier
		}
		cs.Candidates = append(cs.Candidates, Candidate{
			ID:            p.PlanID,
			Description:   p.Goal,
			SourceAbility: "Planning",
			Attributes: map[string]float64{
				"confidence": p.ConfidenceProfile.OverallConfidence,
				"cost":       p.EstimatedCost,
			},
			Metadata: meta,
		})
	}

	if len(cs.Candidates) == 0 {
		return errors.New("decision: parsed envelope yielded no valid candidates")
	}

	// Always trigger Deliberative mode for planning artifact evaluation.
	rec, err := s.EvaluateDeliberative(ctx, cs)
	if err != nil {
		return fmt.Errorf("decision: failed to evaluate candidate plans: %w", err)
	}

	var selectedDesc string
	for _, cand := range cs.Candidates {
		if cand.ID == rec.SelectedCandidateID {
			selectedDesc = cand.Description
			break
		}
	}
	if selectedDesc == "" && len(cs.Candidates) > 0 {
		selectedDesc = cs.Candidates[0].Description
	}

	devLog("Decision", "Candidate selected")

	parentID := env.ParentRef
	if parentID == "" {
		parentID = env.ID
	}
	_, err = PublishDeliberativeDecision(ctx, rec, selectedDesc, s.storer, s.publisher, parentID)
	if err != nil {
		return err
	}
	devLog("Decision", "Published TopicEvaluatedOptions")
	return nil
}

// EvaluateReflexive executes fast-path linear utility scoring (<2ms budget).
// If uncertainty, ambiguity margin, or risk thresholds are breached, it returns
// OutcomeEscalateToDeliberative without automatically invoking Deliberative mode.
func (s *DefaultDecisionService) EvaluateReflexive(ctx context.Context, cs CandidateSet) (*DecisionRecord, error) {
	start := time.Now()
	if err := cs.Validate(); err != nil {
		return nil, err
	}

	snapshot, err := s.strategyProv.ActiveSnapshot()
	if err != nil {
		return nil, err
	}

	trace := s.getOrCreateTrace(cs.EpisodeID, snapshot.StrategyVersion)

	// 1. Tier 1 Hard Constitutional Gate
	surviving, rejected, err := s.tier1Gate.Filter(ctx, cs)
	if err != nil {
		return nil, err
	}

	rec := &DecisionRecord{
		DecisionID:         fmt.Sprintf("dec-%d", time.Now().UnixNano()),
		EpisodeID:          cs.EpisodeID,
		SchemaVersion:      "2.0.0-FROZEN",
		Timestamp:          time.Now(),
		StrategyVersion:    snapshot.StrategyVersion,
		DeliberationDepth:  DepthReflexive,
		ReplaySeed:         uint64(time.Now().UnixNano()),
		RejectedCandidates: rejected,
	}

	// If all candidates rejected by Tier 1 Hard Gate
	if len(surviving) == 0 {
		rec.SelectedOutcome = OutcomeAbstain
		rec.Confidence = 1.0
		rec.Rationale = "all candidate alternatives disqualified by Tier 1 Constitutional Hard Gate"

		latencyUs := uint32(time.Since(start).Microseconds())
		trace.RecordDecision(rec, latencyUs, 0.0, false, nil)
		if err := rec.Validate(); err != nil {
			return nil, err
		}
		return rec, nil
	}

	// 2. Tier 2 Objective Utility Scorer
	scores, err := s.tier2Scorer.ScoreReflexive(surviving, snapshot)
	if err != nil {
		return nil, err
	}

	topScore := scores[0]

	// Populate score deltas for rejected candidates
	for i := range rec.RejectedCandidates {
		rec.RejectedCandidates[i].ScoreDelta = topScore.Score
	}

	// 3. Evaluate Escalation Triggers (Multi-Dimensional Escalation Vector)
	var triggeredDimensions []string
	var topTwoMargin float64 = 1.0
	isNearTie := false

	if len(scores) >= 2 {
		topTwoMargin = math.Abs(scores[0].Score - scores[1].Score)
		if topTwoMargin < snapshot.EscalationAmbiguityMargin {
			triggeredDimensions = append(triggeredDimensions, "AMBIGUITY_MARGIN")
			isNearTie = true
		}
	}

	if topScore.Confidence < snapshot.EscalationConfidenceFloor {
		triggeredDimensions = append(triggeredDimensions, "CONFIDENCE_DROP")
	}

	// Check if top candidate has high tail risk complexity
	for _, cand := range surviving {
		if cand.ID == topScore.CandidateID && cand.EstimatedCost > cand.EstimatedBenefit*2.0 && cand.EstimatedCost > 0 {
			triggeredDimensions = append(triggeredDimensions, "TAIL_RISK")
			break
		}
	}

	var anomaly *MicroDecisionAnomaly

	// If escalation triggered, emit EscalationRecommendation (NEVER automatically invoke Deliberative)
	if len(triggeredDimensions) > 0 {
		rec.SelectedOutcome = OutcomeEscalateToDeliberative
		rec.Confidence = topScore.Confidence
		rec.Rationale = fmt.Sprintf("reflexive evaluation triggered escalation dimensions: %v", triggeredDimensions)
		rec.EscalationRecommendation = &EscalationRecommendation{
			TriggeredDimensions: triggeredDimensions,
			ConfidenceDelta:     snapshot.EscalationConfidenceFloor - topScore.Confidence,
			UtilityScoreMargin:  topTwoMargin,
			Reason:              rec.Rationale,
		}

		anomaly = &MicroDecisionAnomaly{
			DecisionID:      rec.DecisionID,
			Timestamp:       rec.Timestamp,
			AnomalyType:     "ESCALATED",
			TopCandidateID:  topScore.CandidateID,
			ConfidenceScore: topScore.Confidence,
		}
	} else {
		rec.SelectedOutcome = OutcomeCommit
		rec.SelectedCandidateID = topScore.CandidateID
		rec.Confidence = topScore.Confidence
		rec.Rationale = topScore.Rationale
	}

	latencyUs := uint32(time.Since(start).Microseconds())
	trace.RecordDecision(rec, latencyUs, topTwoMargin, isNearTie, anomaly)

	if err := rec.Validate(); err != nil {
		return nil, err
	}

	return rec, nil
}

// EvaluateDeliberative executes rigorous Multi-Criteria Decision Analysis (MCDA) (50-500ms budget).
func (s *DefaultDecisionService) EvaluateDeliberative(ctx context.Context, cs CandidateSet) (*DecisionRecord, error) {
	if err := cs.Validate(); err != nil {
		return nil, err
	}

	snapshot, err := s.strategyProv.ActiveSnapshot()
	if err != nil {
		return nil, err
	}

	surviving, rejected, err := s.tier1Gate.Filter(ctx, cs)
	if err != nil {
		return nil, err
	}

	rec := &DecisionRecord{
		DecisionID:         fmt.Sprintf("dec-delib-%d", time.Now().UnixNano()),
		EpisodeID:          cs.EpisodeID,
		SchemaVersion:      "2.0.0-FROZEN",
		Timestamp:          time.Now(),
		StrategyVersion:    snapshot.StrategyVersion,
		DeliberationDepth:  DepthDeliberative,
		ReplaySeed:         uint64(time.Now().UnixNano()),
		RejectedCandidates: rejected,
	}

	if len(surviving) == 0 {
		rec.SelectedOutcome = OutcomeAbstain
		rec.Confidence = 1.0
		rec.Rationale = "all candidate alternatives disqualified by Tier 1 Constitutional Hard Gate"
		if err := rec.Validate(); err != nil {
			return nil, err
		}
		return rec, nil
	}

	scores, tradeoffMatrix, err := s.tier2Scorer.ScoreDeliberative(ctx, surviving, snapshot)
	if err != nil {
		return nil, err
	}

	topScore := scores[0]
	rec.SelectedOutcome = OutcomeCommit
	rec.SelectedCandidateID = topScore.CandidateID
	rec.Confidence = topScore.Confidence
	rec.Rationale = fmt.Sprintf("deliberative MCDA selected top candidate %s with score %.3f", topScore.CandidateID, topScore.Score)
	rec.TradeoffMatrix = tradeoffMatrix

	if err := rec.Validate(); err != nil {
		return nil, err
	}

	return rec, nil
}

// ExecuteTask satisfies executive.AbilityDriver.
func (s *DefaultDecisionService) ExecuteTask(ctx context.Context, workflowID, nodeID string, budget executive.BudgetTier, payloadRef string) (executive.EpistemicStatus, string, error) {
	return executive.StatusConfident, payloadRef, nil
}

// SelectAction satisfies executive.DecisionAbility.
func (s *DefaultDecisionService) SelectAction(ctx context.Context, optionsRef string) (string, error) {
	return optionsRef, nil
}
