package ontology

// ReferenceType defines the semantic nature of a pointer or referring expression.
type ReferenceType string

const (
	RefPronoun             ReferenceType = "PRONOUN"
	RefDemonstrative       ReferenceType = "DEMONSTRATIVE"
	RefDefiniteDescription ReferenceType = "DEFINITE_DESCRIPTION"
)
