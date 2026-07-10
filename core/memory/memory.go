// Package memory is the Memory Service — Core Service three of three.
//
// The Memory Service provides structured knowledge persistence for IDUN
// components. It sits directly above Storage and directly below Intelligence
// in the five-pillar stack.
//
// Responsibility:
//   - Storing structured records (type, payload, metadata)
//   - Retrieving structured records by caller-assigned ID
//   - Updating record payloads
//   - Deleting records with cascade removal of all associated links
//   - Creating typed, directional links between records
//   - Removing links
//   - Forward and reverse relationship traversal
//   - Type-based indexing for bulk retrieval
//
// What this package intentionally does NOT do:
//   - Decide what to remember or forget             (Intelligence concern)
//   - Score or rank records by importance           (Intelligence concern)
//   - Summarise or infer from record content        (Intelligence concern)
//   - Understand semantic meaning of payloads       (Intelligence concern)
//   - Embeddings, vector search, or AI interaction  (Intelligence concern, V2+)
//   - Raw byte persistence                          (Storage concern)
//   - Route through the Kernel Bus                  (direct injection, same as Logger/Storage)
//   - Self-register in the Kernel Registry          (Host owns all wiring)
//
// Architecture: ADR-CS-003 (frozen 2026-07-10)
//
// Namespace: Memory owns the "memory/" key prefix in Storage. No other
// component may write to keys under this prefix. This is a convention
// enforced by the Host, not a runtime mechanism.
//
// Storage key conventions (internal — callers never see these):
//
//	memory/records/{id}                          → serialised Record (JSON)
//	memory/links/{from}/{relationship}/{to}      → serialised Link (JSON, forward index)
//	memory/rlinks/{to}/{relationship}/{from}     → empty (reverse index, presence only)
//	memory/index/type/{type}/{created_unix}/{id} → empty (type presence index)
package memory

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"idun/core/logger"
	"idun/core/storage"
)

// ErrNotFound is the exported sentinel error returned when a requested
// record or relationship does not exist in Memory. Callers check using:
//
//	errors.Is(err, memory.ErrNotFound)
var ErrNotFound = errors.New("memory: not found")

// ============================================================
// Record
// ============================================================

// Record is the fundamental unit of structured knowledge in Memory.
//
// ID is caller-assigned and must be unique within Memory. Hierarchical
// slash-separated IDs are encouraged (e.g. "person/mom", "task/call-mom").
//
// Type is a caller-defined semantic category (e.g. "person", "task", "event").
// Memory uses Type only for indexing — it does not interpret it.
//
// Payload is opaque bytes. Memory never parses or inspects the payload.
// Callers serialise their domain types before calling CreateRecord and
// deserialise after receiving bytes from GetRecord.
//
// CreatedAt is set by Memory at CreateRecord time and never changes.
// UpdatedAt is set by Memory at CreateRecord and UpdateRecord time.
// Creator is the caller-supplied identity string provided at creation time.
type Record struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Payload   []byte    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Creator   string    `json:"creator"`
}

// ============================================================
// Link
// ============================================================

// Link is a typed, directional edge between two Records.
//
// From and To are record IDs. The edge is directed: From → To.
// Relationship is a caller-defined label (e.g. "concerns", "assigned_to").
// CreatedAt is set by Memory at LinkRecords time and never changes.
// Creator is the caller-supplied identity string provided at link time.
type Link struct {
	From         string    `json:"from"`
	To           string    `json:"to"`
	Relationship string    `json:"relationship"`
	CreatedAt    time.Time `json:"created_at"`
	Creator      string    `json:"creator"`
}

// ============================================================
// Memory interface — public contract
// ============================================================

// Memory is the capability interface that callers depend on.
//
// Callers receive a Memory, never a *MemoryService. This is the same
// architectural pattern used by storage.Store and logger.Writer: callers
// depend on capabilities, not on concrete types.
type Memory interface {
	// CreateRecord creates a new record. Returns an error if a record with
	// the same ID already exists. Callers must call UpdateRecord to modify
	// an existing record. The explicit separation forces callers to express
	// intent — this is a policy decision that belongs in the caller, not
	// in Memory.
	CreateRecord(record Record) error

	// GetRecord returns the record stored under id.
	// Returns ErrNotFound if no record with that ID exists.
	// Returns an owned copy; mutating the result does not affect Memory.
	GetRecord(id string) (Record, error)

	// UpdateRecord replaces the payload of an existing record in full.
	// Updates UpdatedAt. Does not change CreatedAt or Creator.
	// Returns ErrNotFound if the record does not exist.
	UpdateRecord(record Record) error

	// DeleteRecord permanently removes the record and cascades: all forward
	// and reverse links where this record appears as From or To are removed.
	// Returns ErrNotFound if the record does not exist.
	DeleteRecord(id string) error

	// ListRecordsByType returns all records of the given type, in creation
	// order (oldest first). Returns an empty (non-nil) slice if no records
	// of that type exist.
	ListRecordsByType(recordType string) ([]Record, error)

	// LinkRecords creates a directional, typed edge from → to.
	// Writes both a forward index entry and a reverse index entry.
	// Returns an error if either record does not exist.
	// Returns an error if the identical (from, to, relationship) edge exists.
	LinkRecords(from, to, relationship, creator string) error

	// UnlinkRecords removes a specific directional edge and its reverse
	// index entry. Returns ErrNotFound if the edge does not exist.
	UnlinkRecords(from, to, relationship string) error

	// GetLinkedRecords is forward traversal. Returns all records reachable
	// from `from` via edges of the given relationship type.
	// Returns an empty (non-nil) slice if no such edges exist.
	GetLinkedRecords(from, relationship string) ([]Record, error)

	// GetLinkedRecordsReverse is reverse traversal. Returns all records that
	// have a `relationship` edge pointing to `to`.
	// Returns an empty (non-nil) slice if no such edges exist.
	GetLinkedRecordsReverse(to, relationship string) ([]Record, error)
}

// ============================================================
// Config
// ============================================================

// Config configures the Memory Service at construction time.
type Config struct {
	// Namespace is the Storage key prefix used by this Memory Service.
	// If empty, NewMemoryService applies the default "memory/".
	// Override in tests to isolate multiple Memory instances sharing Storage.
	Namespace string
}

// ============================================================
// MemoryService — concrete implementation
// ============================================================

// MemoryService is the concrete implementation of the Memory Service.
//
// It implements:
//   - memory.Memory   (all record and relationship operations)
//   - kernel.Component (Name() string) — for Host registration in the Registry
//
// Do not construct MemoryService with a struct literal. Always use
// NewMemoryService so that dependencies are validated, defaults are applied,
// and the type index is rebuilt from Storage.
//
// MemoryService is safe for concurrent use. Multiple goroutines may call
// any method simultaneously. A single sync.RWMutex serialises writes and
// allows concurrent reads.
type MemoryService struct {
	mu        sync.RWMutex
	ns        string         // Storage key namespace prefix, e.g. "memory/"
	store     storage.Store  // raw persistence backend
	log       logger.Writer  // structured logger
	closed    bool           // true after Close()
	typeIndex map[string][]typeEntry // type → ordered slice of {createdUnix, id}
}

// typeEntry records the creation timestamp and ID of a record in the type index.
// The slice is kept in insertion order (creation order), matching the ADR requirement
// that ListRecordsByType returns records oldest-first.
type typeEntry struct {
	CreatedUnix int64
	ID          string
}

// ============================================================
// Constructor
// ============================================================

// NewMemoryService constructs a Memory Service backed by the given Store and Logger.
//
// cfg.Namespace defaults to "memory/" if empty.
// store must not be nil.
// log must not be nil.
//
// Construction scans the Storage type index to rebuild the in-memory type map.
// Returns an error if dependencies are invalid or if the index scan fails.
func NewMemoryService(cfg Config, store storage.Store, log logger.Writer) (*MemoryService, error) {
	if store == nil {
		return nil, errors.New("memory: store is required")
	}
	if log == nil {
		return nil, errors.New("memory: logger is required")
	}

	ns := cfg.Namespace
	if ns == "" {
		ns = "memory/"
	}

	m := &MemoryService{
		ns:        ns,
		store:     store,
		log:       log,
		typeIndex: make(map[string][]typeEntry),
	}

	if err := m.rebuildTypeIndex(); err != nil {
		return nil, fmt.Errorf("memory: rebuild type index: %w", err)
	}

	m.log.Info("MemoryService initialized",
		logger.Field{Key: "namespace", Value: ns},
	)

	return m, nil
}

// ============================================================
// kernel.Component interface
// ============================================================

// Name returns "MemoryService", satisfying the kernel.Component interface.
// This enables the Host to register Memory in the Service Registry.
// No V1 caller discovers Memory through the Registry at runtime;
// registration is a Host option for diagnostic visibility.
func (m *MemoryService) Name() string {
	return "MemoryService"
}

// ============================================================
// Lifecycle
// ============================================================

// Close marks the MemoryService as closed. After Close, all public methods
// return an error. Close is idempotent: calling it more than once returns nil.
// Close does not close Storage — Memory does not own Storage's lifecycle.
func (m *MemoryService) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil
	}
	m.closed = true

	m.log.Info("MemoryService closed",
		logger.Field{Key: "namespace", Value: m.ns},
	)
	return nil
}

// ============================================================
// Record operations
// ============================================================

// CreateRecord creates a new record. Returns an error if the ID already exists.
func (m *MemoryService) CreateRecord(record Record) error {
	if err := validateRecordInput(record); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return errors.New("memory: closed")
	}

	// Check for duplicate ID.
	exists, err := m.store.Exists(m.recordKey(record.ID))
	if err != nil {
		return fmt.Errorf("memory: CreateRecord %q: check existence: %w", record.ID, err)
	}
	if exists {
		return fmt.Errorf("memory: CreateRecord %q: record already exists", record.ID)
	}

	now := time.Now().UTC()
	record.CreatedAt = now
	record.UpdatedAt = now
	if record.Payload == nil {
		record.Payload = []byte{}
	}

	if err := m.writeRecord(record); err != nil {
		return fmt.Errorf("memory: CreateRecord %q: %w", record.ID, err)
	}

	// Update type index.
	m.typeIndex[record.Type] = append(m.typeIndex[record.Type], typeEntry{
		CreatedUnix: now.Unix(),
		ID:          record.ID,
	})

	m.log.Debug("record created",
		logger.Field{Key: "id", Value: record.ID},
		logger.Field{Key: "type", Value: record.Type},
	)
	return nil
}

// GetRecord returns the record stored under id. Returns ErrNotFound if absent.
func (m *MemoryService) GetRecord(id string) (Record, error) {
	if id == "" {
		return Record{}, errors.New("memory: GetRecord: id must not be empty")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return Record{}, errors.New("memory: closed")
	}

	return m.readRecord(id)
}

// UpdateRecord replaces the payload of an existing record. Returns ErrNotFound
// if the record does not exist. Does not change CreatedAt or Creator.
func (m *MemoryService) UpdateRecord(record Record) error {
	if record.ID == "" {
		return errors.New("memory: UpdateRecord: id must not be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return errors.New("memory: closed")
	}

	// Read existing to preserve immutable fields.
	existing, err := m.readRecord(record.ID)
	if err != nil {
		return fmt.Errorf("memory: UpdateRecord %q: %w", record.ID, err)
	}

	existing.Payload = record.Payload
	if existing.Payload == nil {
		existing.Payload = []byte{}
	}
	existing.UpdatedAt = time.Now().UTC()

	if err := m.writeRecord(existing); err != nil {
		return fmt.Errorf("memory: UpdateRecord %q: %w", record.ID, err)
	}

	m.log.Debug("record updated",
		logger.Field{Key: "id", Value: record.ID},
	)
	return nil
}

// DeleteRecord permanently removes the record and cascades deletion of all
// forward and reverse links where this record appears as From or To.
// Returns ErrNotFound if the record does not exist.
func (m *MemoryService) DeleteRecord(id string) error {
	if id == "" {
		return errors.New("memory: DeleteRecord: id must not be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return errors.New("memory: closed")
	}

	// Confirm the record exists and get its type for index cleanup.
	existing, err := m.readRecord(id)
	if err != nil {
		return fmt.Errorf("memory: DeleteRecord %q: %w", id, err)
	}

	// Cascade: remove all outgoing forward links and their reverse mirrors.
	if err := m.cascadeDeleteOutgoingLinks(id); err != nil {
		return fmt.Errorf("memory: DeleteRecord %q: cascade outgoing: %w", id, err)
	}

	// Cascade: remove all incoming reverse links and their forward mirrors.
	if err := m.cascadeDeleteIncomingLinks(id); err != nil {
		return fmt.Errorf("memory: DeleteRecord %q: cascade incoming: %w", id, err)
	}

	// Delete the record itself.
	if err := m.store.Delete(m.recordKey(id)); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("memory: DeleteRecord %q: %w", id, err)
	}

	// Remove from type index.
	m.removeFromTypeIndex(existing.Type, id)

	m.log.Debug("record deleted",
		logger.Field{Key: "id", Value: id},
	)
	return nil
}

// ListRecordsByType returns all records of the given type in creation order.
// Returns an empty (non-nil) slice if no records of that type exist.
func (m *MemoryService) ListRecordsByType(recordType string) ([]Record, error) {
	if recordType == "" {
		return nil, errors.New("memory: ListRecordsByType: type must not be empty")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return nil, errors.New("memory: closed")
	}

	entries := m.typeIndex[recordType]
	results := make([]Record, 0, len(entries))
	for _, e := range entries {
		rec, err := m.readRecord(e.ID)
		if err != nil {
			// Record may have been deleted between index rebuild and now.
			// Skip it rather than returning a partial error.
			continue
		}
		results = append(results, rec)
	}
	return results, nil
}

// ============================================================
// Relationship operations
// ============================================================

// LinkRecords creates a directional, typed edge from → to.
// Writes both a forward index entry and a reverse index entry.
func (m *MemoryService) LinkRecords(from, to, relationship, creator string) error {
	switch {
	case from == "":
		return errors.New("memory: LinkRecords: from must not be empty")
	case to == "":
		return errors.New("memory: LinkRecords: to must not be empty")
	case relationship == "":
		return errors.New("memory: LinkRecords: relationship must not be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return errors.New("memory: closed")
	}

	// Both records must exist.
	if _, err := m.readRecord(from); err != nil {
		return fmt.Errorf("memory: LinkRecords: from %q: %w", from, err)
	}
	if _, err := m.readRecord(to); err != nil {
		return fmt.Errorf("memory: LinkRecords: to %q: %w", to, err)
	}

	// Check for duplicate edge.
	fwdKey := m.linkKey(from, relationship, to)
	exists, err := m.store.Exists(fwdKey)
	if err != nil {
		return fmt.Errorf("memory: LinkRecords: check duplicate: %w", err)
	}
	if exists {
		return fmt.Errorf("memory: LinkRecords: link %q -[%s]-> %q already exists", from, relationship, to)
	}

	now := time.Now().UTC()
	link := Link{
		From:         from,
		To:           to,
		Relationship: relationship,
		CreatedAt:    now,
		Creator:      creator,
	}

	data, err := json.Marshal(link)
	if err != nil {
		return fmt.Errorf("memory: LinkRecords: serialise: %w", err)
	}

	// Write forward index (carries the full Link payload).
	if err := m.store.Write(fwdKey, data); err != nil {
		return fmt.Errorf("memory: LinkRecords: write forward index: %w", err)
	}

	// Write reverse index (presence only — empty value).
	rlinkKey := m.rlinkKey(to, relationship, from)
	if err := m.store.Write(rlinkKey, []byte{}); err != nil {
		return fmt.Errorf("memory: LinkRecords: write reverse index: %w", err)
	}

	m.log.Debug("link created",
		logger.Field{Key: "from", Value: from},
		logger.Field{Key: "to", Value: to},
		logger.Field{Key: "relationship", Value: relationship},
	)
	return nil
}

// UnlinkRecords removes a specific directional edge and its reverse index entry.
// Returns ErrNotFound if the edge does not exist.
func (m *MemoryService) UnlinkRecords(from, to, relationship string) error {
	switch {
	case from == "":
		return errors.New("memory: UnlinkRecords: from must not be empty")
	case to == "":
		return errors.New("memory: UnlinkRecords: to must not be empty")
	case relationship == "":
		return errors.New("memory: UnlinkRecords: relationship must not be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return errors.New("memory: closed")
	}

	return m.unlinkLocked(from, to, relationship)
}

// GetLinkedRecords performs forward traversal. Returns all records reachable
// from `from` via edges of the given relationship type.
func (m *MemoryService) GetLinkedRecords(from, relationship string) ([]Record, error) {
	switch {
	case from == "":
		return nil, errors.New("memory: GetLinkedRecords: from must not be empty")
	case relationship == "":
		return nil, errors.New("memory: GetLinkedRecords: relationship must not be empty")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return nil, errors.New("memory: closed")
	}

	// Prefix scan the forward index to find all `to` IDs.
	prefix := m.linkPrefix(from, relationship)
	keys, err := m.store.List(prefix)
	if err != nil {
		return nil, fmt.Errorf("memory: GetLinkedRecords: list forward index: %w", err)
	}

	return m.resolveLinkedRecords(keys, prefix)
}

// GetLinkedRecordsReverse performs reverse traversal. Returns all records that
// have a `relationship` edge pointing to `to`.
func (m *MemoryService) GetLinkedRecordsReverse(to, relationship string) ([]Record, error) {
	switch {
	case to == "":
		return nil, errors.New("memory: GetLinkedRecordsReverse: to must not be empty")
	case relationship == "":
		return nil, errors.New("memory: GetLinkedRecordsReverse: relationship must not be empty")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return nil, errors.New("memory: closed")
	}

	// Prefix scan the reverse index to find all `from` IDs.
	prefix := m.rlinkPrefix(to, relationship)
	keys, err := m.store.List(prefix)
	if err != nil {
		return nil, fmt.Errorf("memory: GetLinkedRecordsReverse: list reverse index: %w", err)
	}

	return m.resolveLinkedRecords(keys, prefix)
}

// ============================================================
// Internal helpers — key construction
// ============================================================

func (m *MemoryService) recordKey(id string) string {
	return m.ns + "records/" + id
}

// linkKey builds the forward index Storage key for an edge.
// Key: {ns}links/{from}/{relationship}/{to}
func (m *MemoryService) linkKey(from, relationship, to string) string {
	return m.ns + "links/" + from + "/" + relationship + "/" + to
}

// linkPrefix returns the prefix for scanning all forward edges from `from`
// with the given relationship.
// Prefix: {ns}links/{from}/{relationship}/
func (m *MemoryService) linkPrefix(from, relationship string) string {
	return m.ns + "links/" + from + "/" + relationship + "/"
}

// rlinkKey builds the reverse index Storage key for an edge.
// Key: {ns}rlinks/{to}/{relationship}/{from}
func (m *MemoryService) rlinkKey(to, relationship, from string) string {
	return m.ns + "rlinks/" + to + "/" + relationship + "/" + from
}

// rlinkPrefix returns the prefix for scanning all reverse edges pointing to
// `to` with the given relationship.
// Prefix: {ns}rlinks/{to}/{relationship}/
func (m *MemoryService) rlinkPrefix(to, relationship string) string {
	return m.ns + "rlinks/" + to + "/" + relationship + "/"
}

// typeIndexKey builds a Storage key for the type presence index.
// Key: {ns}index/type/{type}/{createdUnix}/{id}
// The unix timestamp as a zero-padded decimal ensures lexicographic order
// matches creation order when Storage.List returns sorted keys.
func (m *MemoryService) typeIndexKey(recordType string, createdUnix int64, id string) string {
	return fmt.Sprintf("%sindex/type/%s/%020d/%s", m.ns, recordType, createdUnix, id)
}

// typeIndexPrefix returns the prefix for scanning all index entries of a type.
func (m *MemoryService) typeIndexPrefix(recordType string) string {
	return m.ns + "index/type/" + recordType + "/"
}

// ============================================================
// Internal helpers — record I/O
// ============================================================

// readRecord deserialises the record stored under id.
// Translates storage.ErrNotFound to memory.ErrNotFound.
// Must be called while holding at least a read lock.
func (m *MemoryService) readRecord(id string) (Record, error) {
	data, err := m.store.Read(m.recordKey(id))
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return Record{}, ErrNotFound
		}
		return Record{}, fmt.Errorf("read record %q: %w", id, err)
	}

	var rec Record
	if err := json.Unmarshal(data, &rec); err != nil {
		return Record{}, fmt.Errorf("deserialise record %q: %w", id, err)
	}
	return rec, nil
}

// writeRecord serialises and stores a record, and writes its type index entry.
// Must be called while holding the write lock.
func (m *MemoryService) writeRecord(record Record) error {
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("serialise record %q: %w", record.ID, err)
	}

	if err := m.store.Write(m.recordKey(record.ID), data); err != nil {
		return fmt.Errorf("write record %q: %w", record.ID, err)
	}

	// Write type index entry (presence only — empty value).
	idxKey := m.typeIndexKey(record.Type, record.CreatedAt.Unix(), record.ID)
	if err := m.store.Write(idxKey, []byte{}); err != nil {
		return fmt.Errorf("write type index for %q: %w", record.ID, err)
	}

	return nil
}

// ============================================================
// Internal helpers — relationship operations
// ============================================================

// unlinkLocked removes a forward link and its reverse index entry.
// Must be called while holding the write lock.
func (m *MemoryService) unlinkLocked(from, to, relationship string) error {
	fwdKey := m.linkKey(from, relationship, to)
	if err := m.store.Delete(fwdKey); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("delete forward link: %w", err)
	}

	// Remove reverse index entry. If it is already gone, that is acceptable
	// (e.g. partial write followed by recovery). Do not treat as an error.
	rlinkKey := m.rlinkKey(to, relationship, from)
	if err := m.store.Delete(rlinkKey); err != nil && !errors.Is(err, storage.ErrNotFound) {
		return fmt.Errorf("delete reverse link: %w", err)
	}

	m.log.Debug("link removed",
		logger.Field{Key: "from", Value: from},
		logger.Field{Key: "to", Value: to},
		logger.Field{Key: "relationship", Value: relationship},
	)
	return nil
}

// cascadeDeleteOutgoingLinks removes all forward links where `id` is the source
// (i.e. all edges id -[rel]-> x), and their corresponding reverse index entries.
// Must be called while holding the write lock.
func (m *MemoryService) cascadeDeleteOutgoingLinks(id string) error {
	// Forward index prefix for all relationships from this record.
	// Pattern: {ns}links/{id}/
	prefix := m.ns + "links/" + id + "/"
	keys, err := m.store.List(prefix)
	if err != nil {
		return fmt.Errorf("list outgoing links for %q: %w", id, err)
	}

	for _, key := range keys {
		// Key format: {ns}links/{from}/{relationship}/{to}
		// Strip the namespace prefix, then split.
		rel := strings.TrimPrefix(key, m.ns+"links/")
		parts := strings.SplitN(rel, "/", 3)
		if len(parts) != 3 {
			continue // malformed key, skip
		}
		// parts[0]=from, parts[1]=relationship, parts[2]=to
		// but the from here is `id`, relationship is parts[1], to is parts[2]
		// However from could contain slashes. We need to strip `{id}/` prefix.
		fromAndRest := strings.TrimPrefix(rel, id+"/")
		relParts := strings.SplitN(fromAndRest, "/", 2)
		if len(relParts) != 2 {
			continue
		}
		relationship := relParts[0]
		to := relParts[1]

		if err := m.store.Delete(key); err != nil && !errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("delete outgoing link %q: %w", key, err)
		}
		rlinkKey := m.rlinkKey(to, relationship, id)
		if err := m.store.Delete(rlinkKey); err != nil && !errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("delete reverse entry for outgoing link %q: %w", key, err)
		}
	}
	return nil
}

// cascadeDeleteIncomingLinks removes all reverse index entries where `id` is the
// target (i.e. all edges x -[rel]-> id), and their corresponding forward entries.
// Must be called while holding the write lock.
func (m *MemoryService) cascadeDeleteIncomingLinks(id string) error {
	// Reverse index prefix for all relationships pointing to this record.
	// Pattern: {ns}rlinks/{id}/
	prefix := m.ns + "rlinks/" + id + "/"
	keys, err := m.store.List(prefix)
	if err != nil {
		return fmt.Errorf("list incoming links for %q: %w", id, err)
	}

	for _, key := range keys {
		// Key format: {ns}rlinks/{to}/{relationship}/{from}
		// Strip the namespace prefix.
		rel := strings.TrimPrefix(key, m.ns+"rlinks/")
		// Strip `{id}/` to get `{relationship}/{from}`
		toAndRest := strings.TrimPrefix(rel, id+"/")
		relParts := strings.SplitN(toAndRest, "/", 2)
		if len(relParts) != 2 {
			continue
		}
		relationship := relParts[0]
		from := relParts[1]

		if err := m.store.Delete(key); err != nil && !errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("delete reverse entry for incoming link %q: %w", key, err)
		}
		fwdKey := m.linkKey(from, relationship, id)
		if err := m.store.Delete(fwdKey); err != nil && !errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("delete forward entry for incoming link %q: %w", key, err)
		}
	}
	return nil
}

// resolveLinkedRecords fetches the records identified by the IDs encoded in the
// given Storage keys. keys is a sorted list of Storage keys returned by List.
// prefix is stripped from each key to extract the terminal ID component.
// Must be called while holding at least a read lock.
func (m *MemoryService) resolveLinkedRecords(keys []string, prefix string) ([]Record, error) {
	results := make([]Record, 0, len(keys))
	for _, key := range keys {
		// The terminal component after the prefix is the linked record ID.
		linkedID := strings.TrimPrefix(key, prefix)
		if linkedID == "" {
			continue
		}
		rec, err := m.readRecord(linkedID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				// Dangling reference — record was deleted after the link was created.
				// Skip rather than returning a partial error.
				continue
			}
			return nil, fmt.Errorf("fetch linked record %q: %w", linkedID, err)
		}
		results = append(results, rec)
	}
	return results, nil
}

// ============================================================
// Internal helpers — type index
// ============================================================

// rebuildTypeIndex scans the Storage type index prefix and reconstructs the
// in-memory typeIndex map. Called once at construction time.
// Must NOT be called while holding the mutex (called before construction completes).
func (m *MemoryService) rebuildTypeIndex() error {
	prefix := m.ns + "index/type/"
	keys, err := m.store.List(prefix)
	if err != nil {
		return fmt.Errorf("scan type index: %w", err)
	}

	for _, key := range keys {
		// Key format: {ns}index/type/{type}/{createdUnix}/{id}
		rel := strings.TrimPrefix(key, prefix)
		// rel = "{type}/{createdUnix}/{id}"
		// Split into exactly 3 parts: type, createdUnix, id.
		// Note: type and id may themselves contain slashes (e.g. "person/mom").
		// We cannot use a simple split. Instead we split on the first two "/" separators
		// working from the structure: the createdUnix is always 20 digits.
		// Format is: {type}/{20-digit-unix}/{id}
		// We find the createdUnix by locating the 20-digit segment.
		parts := strings.SplitN(rel, "/", -1)
		if len(parts) < 3 {
			continue // malformed, skip
		}

		// Find the 20-digit unix timestamp segment. It is always exactly 20 decimal digits.
		// The type occupies parts[0..n-2] and id is parts[n+1..end] where n is the
		// index of the unix segment.
		unixIdx := -1
		for i, p := range parts {
			if len(p) == 20 && isAllDigits(p) {
				unixIdx = i
				break
			}
		}
		if unixIdx < 0 || unixIdx == 0 || unixIdx >= len(parts)-1 {
			continue // malformed
		}

		recordType := strings.Join(parts[:unixIdx], "/")
		unixStr := parts[unixIdx]
		recordID := strings.Join(parts[unixIdx+1:], "/")

		var createdUnix int64
		if _, err := fmt.Sscanf(unixStr, "%d", &createdUnix); err != nil {
			continue
		}

		m.typeIndex[recordType] = append(m.typeIndex[recordType], typeEntry{
			CreatedUnix: createdUnix,
			ID:          recordID,
		})
	}
	return nil
}

// removeFromTypeIndex removes the entry for `id` from the type index slice
// for `recordType`. Also removes the corresponding Storage key.
// Must be called while holding the write lock.
func (m *MemoryService) removeFromTypeIndex(recordType, id string) {
	entries := m.typeIndex[recordType]
	for i, e := range entries {
		if e.ID == id {
			// Remove type index Storage key.
			idxKey := m.typeIndexKey(recordType, e.CreatedUnix, id)
			_ = m.store.Delete(idxKey) // best-effort; record is already gone

			// Remove from in-memory slice preserving order.
			m.typeIndex[recordType] = append(entries[:i], entries[i+1:]...)
			if len(m.typeIndex[recordType]) == 0 {
				delete(m.typeIndex, recordType)
			}
			return
		}
	}
}

// ============================================================
// Input validation
// ============================================================

// validateRecordInput checks the minimum required fields for CreateRecord.
func validateRecordInput(record Record) error {
	if record.ID == "" {
		return errors.New("memory: record ID must not be empty")
	}
	if record.Type == "" {
		return errors.New("memory: record Type must not be empty")
	}
	return nil
}

// isAllDigits reports whether s consists entirely of ASCII decimal digits.
func isAllDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}
