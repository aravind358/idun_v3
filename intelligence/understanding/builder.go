package understanding

// SemanticFrameBuilder provides a safe, fluent builder for constructing validated,
// immutable SemanticFrame values matching Architecture Version 2.0.
type SemanticFrameBuilder struct {
	frame SemanticFrame
	err   error
}

// NewSemanticFrameBuilder creates a new SemanticFrame builder initialized with default
// version ("2.0"), non-nil AmbiguitySet, and the specified incoming envelope ID.
func NewSemanticFrameBuilder(envelopeID string) *SemanticFrameBuilder {
	b := &SemanticFrameBuilder{
		frame: SemanticFrame{
			FrameVersion: FrameVersion,
			EnvelopeID:   envelopeID,
			Status:       StatusUnambiguous,
			AmbiguitySet: []Hypothesis{},
		},
	}
	if envelopeID == "" {
		b.err = ErrMissingEnvelopeID
	}
	return b
}

// WithStatus sets the InterpretationStatus on the frame.
func (b *SemanticFrameBuilder) WithStatus(status InterpretationStatus) *SemanticFrameBuilder {
	if b.err != nil {
		return b
	}
	b.frame.Status = status
	return b
}

// WithPrimaryHypothesis sets the highest calibrated interpretation hypothesis.
func (b *SemanticFrameBuilder) WithPrimaryHypothesis(intent string, conf float64, layer SourceLayer, slots ...Slot) *SemanticFrameBuilder {
	if b.err != nil {
		return b
	}
	slotCopy := make([]Slot, len(slots))
	copy(slotCopy, slots)

	h := Hypothesis{
		Intent:               intent,
		CalibratedConfidence: conf,
		SourceLayer:          layer,
		Slots:                slotCopy,
	}
	if err := h.Validate(); err != nil {
		b.err = err
		return b
	}
	b.frame.PrimaryHypothesis = h
	return b
}

// AddAmbiguousHypothesis appends a runner-up hypothesis to the bounded AmbiguitySet.
// Returns ErrBeamOverflow if total hypotheses exceed MaxBeamWidth (3).
func (b *SemanticFrameBuilder) AddAmbiguousHypothesis(intent string, conf float64, layer SourceLayer, delta float64, slots ...Slot) *SemanticFrameBuilder {
	if b.err != nil {
		return b
	}
	if len(b.frame.AmbiguitySet)+1 >= MaxBeamWidth {
		b.err = ErrBeamOverflow
		return b
	}

	slotCopy := make([]Slot, len(slots))
	copy(slotCopy, slots)

	h := Hypothesis{
		Intent:               intent,
		CalibratedConfidence: conf,
		SourceLayer:          layer,
		DeltaFromPrimary:     delta,
		Slots:                slotCopy,
	}
	if err := h.Validate(); err != nil {
		b.err = err
		return b
	}

	b.frame.AmbiguitySet = append(b.frame.AmbiguitySet, h)
	if len(b.frame.AmbiguitySet) > 0 && b.frame.Status == StatusUnambiguous {
		b.frame.Status = StatusAmbiguousBeam
	}
	return b
}

// WithTopDownPrior records any top-down dialogue expectation or active goal prior applied.
func (b *SemanticFrameBuilder) WithTopDownPrior(prior string) *SemanticFrameBuilder {
	if b.err != nil {
		return b
	}
	b.frame.TopDownPriorApplied = prior
	return b
}

// WithProcessedDuration records total processing latency in milliseconds.
func (b *SemanticFrameBuilder) WithProcessedDuration(ms float64) *SemanticFrameBuilder {
	if b.err != nil {
		return b
	}
	b.frame.ProcessedDurationMs = ms
	return b
}

// Build validates all invariants and returns a deep-cloned immutable SemanticFrame.
func (b *SemanticFrameBuilder) Build() (SemanticFrame, error) {
	if b.err != nil {
		return SemanticFrame{}, b.err
	}
	if err := b.frame.Validate(); err != nil {
		return SemanticFrame{}, err
	}
	return b.frame.Clone(), nil
}
