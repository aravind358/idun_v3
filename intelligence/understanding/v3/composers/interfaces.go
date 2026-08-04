package composers

import (
	"idun/intelligence/understanding/v3"
)

// Runner coordinates semantic composition.
type Runner interface {
	Run(b *v3.Builder)
}
