// Package scheduler implements IDUN's Core Scheduler Service (Version 1).
//
// As specified in frozen ADR-CS-004 v1.1, Scheduler is a deterministic mechanical
// engine responsible exclusively for absolute point-in-time countdown and dispatch.
//
// It exposes exactly two methods: ScheduleOnce and Cancel.
// All scheduled jobs are persisted in the Memory Service under Record.Type = "scheduled_job".
package scheduler

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"idun/core/logger"
	"idun/core/memory"
)

const (
	// JobRecordType is the memory.Record.Type used to persist scheduled jobs.
	JobRecordType = "scheduled_job"
)

var (
	// ErrJobNotFound is returned when a requested job ID does not exist.
	ErrJobNotFound = errors.New("scheduler: job not found")

	// ErrDuplicateID is returned when scheduling a job with an ID that already exists.
	ErrDuplicateID = errors.New("scheduler: job ID already exists")

	// ErrClosed is returned when operating on a closed SchedulerService.
	ErrClosed = errors.New("scheduler: service closed")
)

// Job represents a scheduled absolute point-in-time execution.
type Job struct {
	ID        string        `json:"id"`
	Target    string        `json:"target"`
	Payload   []byte        `json:"payload"`
	ExecuteAt time.Time     `json:"execute_at"`
	Interval  time.Duration `json:"interval,omitempty"`
	Retries   int           `json:"retries,omitempty"`
	MaxRetry  int           `json:"max_retry,omitempty"`
}

// Dispatcher defines the capability required to deliver a due job to its target.
type Dispatcher interface {
	Dispatch(target string, payload []byte) error
}

// Scheduler defines the capability interface for time and periodic scheduling.
type Scheduler interface {
	ScheduleOnce(id, target string, payload []byte, executeAt time.Time) error
	SchedulePeriodic(id, target string, payload []byte, interval time.Duration) error
	Cancel(id string) error
}

// Config configures the SchedulerService.
type Config struct{}

// SchedulerService implements Scheduler and kernel.Component.
type SchedulerService struct {
	mu     sync.Mutex
	mem    memory.Memory
	log    logger.Writer
	disp   Dispatcher
	queue  []Job
	closed bool

	startOnce sync.Once
	closeOnce sync.Once
	wg        sync.WaitGroup
	wakeup    chan struct{}
}

// NewSchedulerService constructs a new SchedulerService, recovers any persisted jobs
// from Memory, immediately dispatches overdue jobs, and prepares the time queue.
func NewSchedulerService(cfg Config, mem memory.Memory, log logger.Writer, disp Dispatcher) (*SchedulerService, error) {
	if mem == nil {
		return nil, errors.New("scheduler: nil memory service")
	}
	if log == nil {
		return nil, errors.New("scheduler: nil logger")
	}
	if disp == nil {
		return nil, errors.New("scheduler: nil dispatcher")
	}

	s := &SchedulerService{
		mem:    mem,
		log:    log,
		disp:   disp,
		wakeup: make(chan struct{}, 1),
	}

	if err := s.recoverJobs(); err != nil {
		return nil, err
	}

	return s, nil
}

// Name returns the canonical kernel component name.
func (s *SchedulerService) Name() string {
	return "SchedulerService"
}

// recoverJobs loads scheduled jobs from Memory. Overdue jobs are dispatched immediately
// and deleted or rescheduled from Memory. Future jobs are enqueued in chronological order.
func (s *SchedulerService) recoverJobs() error {
	records, err := s.mem.ListRecordsByType(JobRecordType)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	for _, rec := range records {
		var job Job
		if err := json.Unmarshal(rec.Payload, &job); err != nil {
			s.log.Warn("scheduler: ignoring corrupt job record in memory",
				logger.Field{Key: "id", Value: rec.ID},
				logger.Field{Key: "error", Value: err.Error()},
			)
			continue
		}

		// Ensure payload is normalized to non-nil slice.
		if job.Payload == nil {
			job.Payload = []byte{}
		}

		if !job.ExecuteAt.After(now) {
			// Overdue job: dispatch immediately.
			s.log.Info("scheduler: dispatching overdue job on boot",
				logger.Field{Key: "id", Value: job.ID},
				logger.Field{Key: "target", Value: job.Target},
			)
			if err := s.disp.Dispatch(job.Target, job.Payload); err != nil {
				s.log.Error("scheduler: overdue job dispatch failed",
					logger.Field{Key: "id", Value: job.ID},
					logger.Field{Key: "error", Value: err.Error()},
				)
			}
			if job.Interval > 0 {
				job.ExecuteAt = now.Add(job.Interval)
				job.Retries = 0
				data, _ := json.Marshal(job)
				_ = s.mem.UpdateRecord(memory.Record{
					ID:      job.ID,
					Type:    JobRecordType,
					Payload: data,
					Creator: "SchedulerService",
				})
				s.queue = append(s.queue, job)
			} else {
				_ = s.mem.DeleteRecord(job.ID)
			}
			continue
		}

		s.queue = append(s.queue, job)
	}

	s.sortQueue()
	return nil
}

func (s *SchedulerService) sortQueue() {
	sort.Slice(s.queue, func(i, j int) bool {
		return s.queue[i].ExecuteAt.Before(s.queue[j].ExecuteAt)
	})
}

// Start launches the background timer loop.
func (s *SchedulerService) Start() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return ErrClosed
	}
	s.mu.Unlock()

	s.startOnce.Do(func() {
		s.wg.Add(1)
		go s.runLoop()
	})
	return nil
}

// Close gracefully stops the background timer loop.
func (s *SchedulerService) Close() error {
	s.closeOnce.Do(func() {
		s.mu.Lock()
		s.closed = true
		s.mu.Unlock()

		s.signalWakeup()
		s.wg.Wait()
	})
	return nil
}

func (s *SchedulerService) signalWakeup() {
	select {
	case s.wakeup <- struct{}{}:
	default:
	}
}

// ScheduleOnce schedules a job to execute at absolute timestamp executeAt.
// If executeAt is in the past or now, it is dispatched immediately.
func (s *SchedulerService) ScheduleOnce(id, target string, payload []byte, executeAt time.Time) error {
	if id == "" {
		return errors.New("scheduler: empty job ID")
	}
	if target == "" {
		return errors.New("scheduler: empty target")
	}
	if executeAt.IsZero() {
		return errors.New("scheduler: zero executeAt")
	}
	if payload == nil {
		payload = []byte{}
	}

	s.mu.Lock()

	if s.closed {
		s.mu.Unlock()
		return ErrClosed
	}

	// Check for duplicate in in-memory queue.
	for _, j := range s.queue {
		if j.ID == id {
			s.mu.Unlock()
			return ErrDuplicateID
		}
	}
	// Check Memory for duplicate record ID.
	if _, err := s.mem.GetRecord(id); err == nil {
		s.mu.Unlock()
		return ErrDuplicateID
	}

	now := time.Now().UTC()
	if !executeAt.After(now) {
		// Past or immediate execution: unlock mutex before dispatching.
		s.mu.Unlock()
		return s.disp.Dispatch(target, payload)
	}

	job := Job{
		ID:        id,
		Target:    target,
		Payload:   payload,
		ExecuteAt: executeAt.UTC(),
	}

	data, err := json.Marshal(job)
	if err != nil {
		s.mu.Unlock()
		return err
	}

	rec := memory.Record{
		ID:      id,
		Type:    JobRecordType,
		Payload: data,
		Creator: "SchedulerService",
	}
	if err := s.mem.CreateRecord(rec); err != nil {
		s.mu.Unlock()
		return err
	}

	wasHead := len(s.queue) == 0 || executeAt.Before(s.queue[0].ExecuteAt)
	s.queue = append(s.queue, job)
	s.sortQueue()
	s.mu.Unlock()

	if wasHead {
		s.signalWakeup()
	}

	return nil
}

// SchedulePeriodic schedules a recurring job that dispatches every interval.
func (s *SchedulerService) SchedulePeriodic(id, target string, payload []byte, interval time.Duration) error {
	if id == "" {
		return errors.New("scheduler: empty job ID")
	}
	if target == "" {
		return errors.New("scheduler: empty target")
	}
	if interval <= 0 {
		return errors.New("scheduler: non-positive interval")
	}
	if payload == nil {
		payload = []byte{}
	}

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return ErrClosed
	}

	for _, j := range s.queue {
		if j.ID == id {
			s.mu.Unlock()
			return ErrDuplicateID
		}
	}
	if _, err := s.mem.GetRecord(id); err == nil {
		s.mu.Unlock()
		return ErrDuplicateID
	}

	executeAt := time.Now().UTC().Add(interval)
	job := Job{
		ID:        id,
		Target:    target,
		Payload:   payload,
		ExecuteAt: executeAt,
		Interval:  interval,
		MaxRetry:  3,
	}

	data, err := json.Marshal(job)
	if err != nil {
		s.mu.Unlock()
		return err
	}

	rec := memory.Record{
		ID:      id,
		Type:    JobRecordType,
		Payload: data,
		Creator: "SchedulerService",
	}
	if err := s.mem.CreateRecord(rec); err != nil {
		s.mu.Unlock()
		return err
	}

	wasHead := len(s.queue) == 0 || executeAt.Before(s.queue[0].ExecuteAt)
	s.queue = append(s.queue, job)
	s.sortQueue()
	s.mu.Unlock()

	if wasHead {
		s.signalWakeup()
	}

	return nil
}

// Cancel removes a scheduled job by ID from the queue and deletes it from Memory.
func (s *SchedulerService) Cancel(id string) error {
	if id == "" {
		return errors.New("scheduler: empty job ID")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrClosed
	}

	foundIdx := -1
	for i, j := range s.queue {
		if j.ID == id {
			foundIdx = i
			break
		}
	}

	if foundIdx == -1 {
		// Verify if it exists in Memory at all.
		_, err := s.mem.GetRecord(id)
		if errors.Is(err, memory.ErrNotFound) {
			return ErrJobNotFound
		}
		// If found in Memory but not queue, delete it.
		_ = s.mem.DeleteRecord(id)
		return ErrJobNotFound
	}

	wasHead := foundIdx == 0
	s.queue = append(s.queue[:foundIdx], s.queue[foundIdx+1:]...)

	if err := s.mem.DeleteRecord(id); err != nil && !errors.Is(err, memory.ErrNotFound) {
		return err
	}

	if wasHead {
		s.signalWakeup()
	}

	return nil
}

// runLoop manages the background countdown timer and dispatches due jobs.
func (s *SchedulerService) runLoop() {
	defer s.wg.Done()

	var timer *time.Timer
	var timerCh <-chan time.Time

	stopTimer := func() {
		if timer != nil {
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer = nil
			timerCh = nil
		}
	}

	for {
		s.mu.Lock()
		if s.closed {
			s.mu.Unlock()
			stopTimer()
			return
		}

		if len(s.queue) == 0 {
			stopTimer()
			s.mu.Unlock()
			<-s.wakeup
			continue
		}

		headJob := s.queue[0]
		now := time.Now().UTC()
		delay := headJob.ExecuteAt.Sub(now)

		if delay <= 0 {
			// Job is due!
			s.queue = s.queue[1:]
			s.mu.Unlock()

			err := s.disp.Dispatch(headJob.Target, headJob.Payload)
			if err != nil {
				s.log.Error("scheduler: job dispatch failed",
					logger.Field{Key: "id", Value: headJob.ID},
					logger.Field{Key: "target", Value: headJob.Target},
					logger.Field{Key: "error", Value: err.Error()},
				)
				maxRetries := headJob.MaxRetry
				if maxRetries == 0 && headJob.Interval > 0 {
					maxRetries = 3
				}
				if maxRetries > 0 && headJob.Retries < maxRetries {
					headJob.Retries++
					backoff := time.Duration(1<<headJob.Retries) * 100 * time.Millisecond
					headJob.ExecuteAt = time.Now().UTC().Add(backoff)
					s.log.Warn("scheduler: scheduling retry backoff for failed job",
						logger.Field{Key: "id", Value: headJob.ID},
						logger.Field{Key: "retry", Value: fmt.Sprintf("%d", headJob.Retries)},
					)
					data, _ := json.Marshal(headJob)
					s.mu.Lock()
					if !s.closed {
						if _, recErr := s.mem.GetRecord(headJob.ID); recErr == nil {
							_ = s.mem.UpdateRecord(memory.Record{
								ID:      headJob.ID,
								Type:    JobRecordType,
								Payload: data,
								Creator: "SchedulerService",
							})
							s.queue = append(s.queue, headJob)
							s.sortQueue()
							s.signalWakeup()
						}
					}
					s.mu.Unlock()
					continue
				}
				if maxRetries > 0 {
					s.log.Error("scheduler: job dispatch failed after max retries",
						logger.Field{Key: "id", Value: headJob.ID},
					)
				}
			}

			if headJob.Interval > 0 {
				headJob.ExecuteAt = time.Now().UTC().Add(headJob.Interval)
				headJob.Retries = 0
				data, _ := json.Marshal(headJob)
				s.mu.Lock()
				if !s.closed {
					if _, recErr := s.mem.GetRecord(headJob.ID); recErr == nil {
						_ = s.mem.UpdateRecord(memory.Record{
							ID:      headJob.ID,
							Type:    JobRecordType,
							Payload: data,
							Creator: "SchedulerService",
						})
						s.queue = append(s.queue, headJob)
						s.sortQueue()
						s.signalWakeup()
					}
				}
				s.mu.Unlock()
			} else {
				_ = s.mem.DeleteRecord(headJob.ID)
			}
			continue
		}

		// Job is in the future: arm timer.
		stopTimer()
		timer = time.NewTimer(delay)
		timerCh = timer.C
		s.mu.Unlock()

		select {
		case <-timerCh:
			// Timer fired; loop around to pop and dispatch headJob.
		case <-s.wakeup:
			// Queue changed or closed; loop around to re-evaluate.
		}
	}
}
