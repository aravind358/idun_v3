package understanding

import (
	"fmt"
	"testing"
)

func TestScratch(t *testing.T) {
	norm := NewDefaultNormalizer()
	n := norm.Normalize("what about battery")
	fmt.Printf("Cleaned: '%s'\n", n.Cleaned)
	
	n2 := norm.Normalize("and tomorrow")
	fmt.Printf("Cleaned: '%s'\n", n2.Cleaned)
}	

