package runtime

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"time"
)

// RuntimeManifest records immutable provenance metadata generated once during startup.
// It captures exact versions and fingerprints across all active subsystems to ensure
// deterministic replayability and verifiable boot records.
type RuntimeManifest struct {
	// RuntimeVersion identifies the runtime infrastructure version.
	RuntimeVersion string `json:"runtime_version"`

	// GeneratedAt records the UTC timestamp when the manifest was frozen.
	GeneratedAt time.Time `json:"generated_at"`

	// SubsystemVersions maps each subsystem canonical name to its release version.
	SubsystemVersions map[string]string `json:"subsystem_versions"`

	// PolicyFingerprints maps each subsystem canonical name to its policy profile hash.
	PolicyFingerprints map[string]string `json:"policy_fingerprints"`

	// CapabilityFingerprints maps each subsystem canonical name to its capability profile hash.
	CapabilityFingerprints map[string]string `json:"capability_fingerprints"`

	// ManifestFingerprint is the cryptographic SHA-256 hash of all above properties combined.
	ManifestFingerprint string `json:"manifest_fingerprint"`
}

// GenerateManifest constructs an immutable RuntimeManifest from known subsystem properties.
func GenerateManifest(version string, subsysVersions, policyHashes, capHashes map[string]string) *RuntimeManifest {
	m := &RuntimeManifest{
		RuntimeVersion:         version,
		GeneratedAt:            time.Now().UTC(),
		SubsystemVersions:      cloneMap(subsysVersions),
		PolicyFingerprints:     cloneMap(policyHashes),
		CapabilityFingerprints: cloneMap(capHashes),
	}
	m.ManifestFingerprint = m.computeHash()
	return m
}

func cloneMap(src map[string]string) map[string]string {
	if src == nil {
		return make(map[string]string)
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func (m *RuntimeManifest) computeHash() string {
	h := sha256.New()
	h.Write([]byte(m.RuntimeVersion))
	writeSortedMap(h, m.SubsystemVersions)
	writeSortedMap(h, m.PolicyFingerprints)
	writeSortedMap(h, m.CapabilityFingerprints)
	return hex.EncodeToString(h.Sum(nil))
}

func writeSortedMap(h interface{ Write([]byte) (int, error) }, m map[string]string) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		_, _ = h.Write([]byte(fmt.Sprintf("%s=%s;", k, m[k])))
	}
}
