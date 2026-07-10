// Package storage is the Storage Service — Core Service two of three.
//
// The Storage Service provides raw, key-addressed physical persistence for IDUN
// components. It is purely mechanical: it stores bytes under hierarchical keys
// and retrieves them faithfully.
//
// Responsibility:
//   - Physical persistence of raw bytes under string keys
//   - Distinguishing "key not found" from I/O failure via ErrNotFound
//   - Goroutine-safe concurrent access
//
// What this package intentionally does NOT do:
//   - Understand semantic meaning or relationships (Memory concern)
//   - Enforce schema or data validation (caller concern)
//   - Provide query or search capabilities (Intelligence/Memory concern)
//   - Route operations through the Kernel Bus (direct dependency injection)
//   - Register itself in the Kernel Registry (Host owns all wiring)
//
// Architecture: ADR-CS-002 (frozen 2026-07-09)
package storage

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"idun/core/logger"
)

// ErrNotFound is the exported sentinel error returned when a requested key
// does not exist in storage. Callers should check failures using:
//
//	errors.Is(err, storage.ErrNotFound)
var ErrNotFound = errors.New("storage: key not found")

// Store defines the capability contract exposed to callers requiring raw
// persistence (such as the Memory Service).
//
// Callers depend on Store rather than *Storage.
type Store interface {
	Write(key string, value []byte) error
	Read(key string) ([]byte, error)
	Delete(key string) error
	Exists(key string) (bool, error)
	List(prefix string) ([]string, error)
}

// Config configures the Storage Service at construction time.
type Config struct {
	// Path is the root directory where the filesystem backend stores data.
	// Required; NewStorage returns an error if Path is empty.
	Path string
}

// Storage is the concrete implementation of the Storage Service.
// It implements Store, kernel.Component (Name() string), and Close() error.
type Storage struct {
	mu     sync.Mutex
	path   string
	log    logger.Writer
	closed bool
}

// NewStorage constructs a new Storage Service using the provided configuration
// and injected logger.
//
// Returns an error if cfg.Path is empty or log is nil.
func NewStorage(cfg Config, log logger.Writer) (*Storage, error) {
	if cfg.Path == "" {
		return nil, errors.New("storage: path is required")
	}
	if log == nil {
		return nil, errors.New("storage: logger is required")
	}

	absPath, err := filepath.Abs(cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("storage: resolve absolute path: %w", err)
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil, fmt.Errorf("storage: create root directory: %w", err)
	}

	s := &Storage{
		path: absPath,
		log:  log,
	}

	s.log.Info("StorageService initialized",
		logger.Field{Key: "path", Value: absPath},
	)

	return s, nil
}

// Name returns "StorageService", satisfying the kernel.Component interface
// for Service Registry registration.
func (s *Storage) Name() string {
	return "StorageService"
}

// Close flushes any pending resources and marks the storage closed.
// Close is idempotent: calling Close on an already-closed Storage returns nil.
// After Close is called, all Store operations return an error.
func (s *Storage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true

	s.log.Info("StorageService closed",
		logger.Field{Key: "path", Value: s.path},
	)
	return nil
}

// Write stores value bytes under key. Overwrites if the key already exists.
// Automatically creates any necessary parent directories.
func (s *Storage) Write(key string, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fullPath, err := s.resolveKeyLocked(key)
	if err != nil {
		return err
	}

	parentDir := filepath.Dir(fullPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("storage: write %q: create directory: %w", key, err)
	}

	if value == nil {
		value = []byte{}
	}

	if err := os.WriteFile(fullPath, value, 0644); err != nil {
		return fmt.Errorf("storage: write %q: %w", key, err)
	}

	return nil
}

// Read retrieves the bytes stored under key.
// Returns (nil, ErrNotFound) if the key does not exist.
// Returns an owned copy of the stored bytes.
func (s *Storage) Read(key string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fullPath, err := s.resolveKeyLocked(key)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("storage: read %q: %w", key, err)
	}

	out := make([]byte, len(data))
	copy(out, data)
	return out, nil
}

// Delete permanently removes the record stored under key.
// Returns ErrNotFound if the key does not exist.
func (s *Storage) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fullPath, err := s.resolveKeyLocked(key)
	if err != nil {
		return err
	}

	err = os.Remove(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrNotFound
		}
		return fmt.Errorf("storage: delete %q: %w", key, err)
	}

	return nil
}

// Exists reports whether a key has a stored value.
// Returns (false, nil) if the key does not exist (not-found is not an error).
func (s *Storage) Exists(key string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fullPath, err := s.resolveKeyLocked(key)
	if err != nil {
		return false, err
	}

	info, err := os.Stat(fullPath)
	if err == nil {
		if info.IsDir() {
			return false, nil
		}
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("storage: exists %q: %w", key, err)
}

// List returns all keys matching the string prefix, sorted alphabetically.
// Passing an empty prefix "" returns all stored keys.
func (s *Storage) List(prefix string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, errors.New("storage: closed")
	}

	keys := []string{}

	err := filepath.WalkDir(s.path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, os.ErrNotExist) && path == s.path {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(s.path, path)
		if err != nil {
			return err
		}

		key := filepath.ToSlash(rel)
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("storage: list: %w", err)
	}

	sort.Strings(keys)
	return keys, nil
}

// resolveKeyLocked validates key constraints and converts a logical key to an absolute filepath.
// Must be called while holding s.mu.
func (s *Storage) resolveKeyLocked(key string) (string, error) {
	if s.closed {
		return "", errors.New("storage: closed")
	}
	if key == "" {
		return "", errors.New("storage: empty key")
	}
	if strings.ContainsRune(key, '\x00') {
		return "", errors.New("storage: invalid key characters")
	}

	cleanKey := filepath.Clean(filepath.FromSlash(key))
	fullPath := filepath.Join(s.path, cleanKey)

	// Prevent directory traversal outside s.path
	rel, err := filepath.Rel(s.path, fullPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", errors.New("storage: key path traversal")
	}

	return fullPath, nil
}
