package foundation

import (
	"regexp"
	"testing"
	"time"
)

func TestNewUUID(t *testing.T) {
	uuid1, err := NewUUID()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	uuid2, err := NewUUID()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if uuid1 == uuid2 {
		t.Errorf("expected unique UUIDs, got duplicates: %s", uuid1)
	}

	// Basic UUID v4 validation regex
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	if !uuidRegex.MatchString(uuid1) {
		t.Errorf("UUID %q does not match v4 format", uuid1)
	}
}

func TestTimestampString(t *testing.T) {
	now := time.Now()
	ts := Timestamp(now)
	str := ts.String()

	if str == "" {
		t.Errorf("Timestamp.String() returned empty string")
	}

	_, err := time.Parse(time.RFC3339Nano, str)
	if err != nil {
		t.Errorf("Timestamp.String() is not a valid RFC3339Nano string: %v", err)
	}
}
