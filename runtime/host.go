package runtime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	coretime "idun/core/time"
	"idun/core/logger"
	"idun/core/memory"
	"idun/core/scheduler"
	"idun/core/storage"
	"idun/capabilities"
	"idun/capabilities/applications"
	"idun/capabilities/applications/core"
	"idun/capabilities/native"
	"idun/intelligence/attention"
	"idun/intelligence/calibration"
	"idun/intelligence/communication"
	"idun/intelligence/constitution"
	intelcontext "idun/intelligence/context"
	decisionv3 "idun/intelligence/decision/v3"
	executivev3 "idun/intelligence/executive/v3"
	"idun/intelligence/infrastructure/inference"
	"idun/intelligence/infrastructure/registry"
	"idun/intelligence/learning"
	planningv3 "idun/intelligence/planning/v3"
	"idun/intelligence/reasoning"
	reasoningv3 "idun/intelligence/reasoning/v3"
	"idun/intelligence/reflection"

	underv3 "idun/intelligence/understanding/v3"
	underext "idun/intelligence/understanding/v3/extractors"
	undernorms "idun/intelligence/understanding/v3/normalizers"
	undercomps "idun/intelligence/understanding/v3/composers"
	underspl "idun/intelligence/understanding/v3/splitter"
	"idun/intelligence/workspace"
	"idun/kernel"
	"idun/world"
	worldtext "idun/world/adapters/text"
	"idun/world/output"
	"idun/world/output/engine"
	"idun/world/output/formatter"
	"idun/world/output/strategy"
	outtext "idun/world/plugins/text"
)

// ErrInvalidTransition indicates an illegal lifecycle operation given current runtime status.
var ErrInvalidTransition = errors.New("runtime: invalid status transition")

type runtimeDispatcher struct {
	log logger.Writer
}

func (d *runtimeDispatcher) Dispatch(target string, payload []byte) error {
	if d.log != nil {
		d.log.Info("runtime: scheduler dispatching target",
			logger.Field{Key: "target", Value: target},
			logger.Field{Key: "payload_bytes", Value: fmt.Sprintf("%d", len(payload))},
		)
	}
	return nil
}

type workspacePubAdapter struct {
	ws *workspace.Engine
}

func (a *workspacePubAdapter) Publish(ctx context.Context, env communication.Envelope) error {
	return a.ws.Publish(ctx, env)
}


type decisionWorkspaceSubAdapter struct {
	ws *workspace.Engine
}

func (a *decisionWorkspaceSubAdapter) Subscribe(topic communication.TopicID, subscriberID string, handler func(ctx context.Context, env communication.Envelope) error) (decisionv3.WorkspaceSubscription, error) {
	return a.ws.Subscribe(topic, subscriberID, handler)
}

type executiveWorkspaceSubAdapter struct {
	ws *workspace.Engine
}

func (a *executiveWorkspaceSubAdapter) Subscribe(topic communication.TopicID, subscriberID string, handler func(ctx context.Context, env communication.Envelope) error) (executivev3.WorkspaceSubscription, error) {
	return a.ws.Subscribe(topic, subscriberID, handler)
}

type contextWorkspaceSubAdapter struct {
	ws *workspace.Engine
}

func (a *contextWorkspaceSubAdapter) Subscribe(topic communication.TopicID, subscriberID string, handler func(ctx context.Context, env communication.Envelope) error) (intelcontext.WorkspaceSubscription, error) {
	return a.ws.Subscribe(topic, subscriberID, handler)
}

type dummyDialogueStateReader struct{}

func (d *dummyDialogueStateReader) GetRecentCandidates(role string, limit int) []string {
	return nil
}

func (d *dummyDialogueStateReader) GetActiveGoals() []string {
	return nil
}

func (d *dummyDialogueStateReader) GetPreviousBatch() *underv3.UnderstandingBatch {
	return nil
}

func (d *dummyDialogueStateReader) GetTemporalAnchor() time.Time {
	return time.Now()
}

type understandingWorkspaceSubAdapter struct {
	ws *workspace.Engine
}

func (a *understandingWorkspaceSubAdapter) Subscribe(topic communication.TopicID, subscriberID string, handler func(ctx context.Context, env communication.Envelope) error) (underv3.WorkspaceSubscription, error) {
	return a.ws.Subscribe(topic, subscriberID, handler)
}

type reasoningWorkspaceSubAdapter struct {
	ws *workspace.Engine
}

func (a *reasoningWorkspaceSubAdapter) Subscribe(topic communication.TopicID, subscriberID string, handler func(ctx context.Context, env communication.Envelope) error) (reasoningv3.WorkspaceSubscription, error) {
	return a.ws.Subscribe(topic, subscriberID, handler)
}

type planningWorkspaceSubAdapter struct {
	ws *workspace.Engine
}

func (a *planningWorkspaceSubAdapter) Subscribe(topic communication.TopicID, subscriberID string, handler func(ctx context.Context, env communication.Envelope) error) (planningv3.WorkspaceSubscription, error) {
	return a.ws.Subscribe(topic, subscriberID, handler)
}

type payloadStorerAdapter struct {
	store *storage.Storage
}

func (a *payloadStorerAdapter) Store(ctx context.Context, data []byte) (string, error) {
	h := sha256.Sum256(data)
	key := hex.EncodeToString(h[:])
	if err := a.store.Write(key, data); err != nil {
		return "", err
	}
	return key, nil
}

func (a *payloadStorerAdapter) StoreArtifact(ctx context.Context, artifactID string, meta storage.ArtifactIndexMeta, data []byte) (string, error) {
	h := sha256.Sum256(data)
	payloadRef := hex.EncodeToString(h[:])
	if err := a.store.StoreArtifact(artifactID, meta, payloadRef, data); err != nil {
		return "", err
	}
	return payloadRef, nil
}

func (a *payloadStorerAdapter) LookupArtifact(ctx context.Context, artifactID string) (storage.ArtifactIndexMeta, error) {
	return a.store.LookupArtifact(artifactID)
}

func (a *payloadStorerAdapter) Retrieve(ctx context.Context, key string) ([]byte, error) {
	return a.store.Read(key)
}

func (a *payloadStorerAdapter) Delete(key string) error {
	return a.store.Delete(key)
}

func (a *payloadStorerAdapter) Exists(key string) (bool, error) {
	return a.store.Exists(key)
}

func (a *payloadStorerAdapter) List(prefix string) ([]string, error) {
	return a.store.List(prefix)
}

// RuntimeHost is the master orchestrator responsible for constructing, wiring,
// registering, booting, and stopping all Layer 1 cognitive and foundational subsystems.
type RuntimeHost struct {
	mu         sync.RWMutex
	cfg        RuntimeConfiguration
	status     RuntimeStatus
	kernel     *kernel.Kernel
	manifest   *RuntimeManifest
	report     *RuntimeBootReport
	subs       []workspace.Subscription
	wrappers   map[string]*ComponentWrapper
	built      bool
	wired      bool
	registered bool

	// Core instances
	loggerSvc    logger.Writer
	storageSvc   *storage.Storage
	memorySvc    memory.Memory
	schedulerSvc *scheduler.SchedulerService
	timeSvc      coretime.TimeService

	// Foundation instances
	registrySvc   *kernel.Registry
	busSvc        *kernel.Bus
	boundarySvc   *kernel.BoundaryEngine
	permissionSvc *kernel.PermissionEngine
	constGate     *constitution.Gate
	calibSvc      *calibration.Service

	// Capability instances
	capManager capabilities.CapabilityManager
	capHandler *capabilities.ActionExecutionHandler

	// Workspace & Cognitive instances
	workspaceSvc *workspace.Engine
	attentionSvc *attention.Service
	execV3Svc    *executivev3.WorkspaceBridge
	underV3Svc   *underv3.WorkspaceBridge
	contextSvc   *intelcontext.WorkspaceBridge
	reasonV3Svc  *reasoningv3.WorkspaceBridge
	planV3Svc    *planningv3.WorkspaceBridge
	decisionV3Svc *decisionv3.WorkspaceBridge
	reflectSvc   *reflection.Service
	learningSvc  *learning.Service

	// Shared infrastructure instances
	modelRegSvc    *registry.Service
	inferenceSvc   *inference.Service

	// World boundary subsystem
	worldSvc  *world.Service
	inReader  io.Reader
	outWriter io.Writer
	doneCh    chan struct{}
}

func (h *RuntimeHost) Storage() *storage.Storage {
	return h.storageSvc
}

type HostOption func(*RuntimeHost)

// WithIOReaders injects custom I/O readers/writers for the World boundary subsystem.
func WithIOReaders(in io.Reader, out io.Writer) HostOption {
	return func(h *RuntimeHost) {
		h.inReader = in
		h.outWriter = out
	}
}

// WithRealizationModel overrides the Ollama model used by the local-realizer backend.
// Use this in tests or specialized deployments where a specific model is known to be installed.
//
//	Example: WithRealizationModel("llama3.1:8b")
func WithRealizationModel(model string) HostOption {
	return func(h *RuntimeHost) {
		if model != "" {
			h.cfg.DefaultRealizationModel = model
		}
	}
}

// NewHost initializes a new RuntimeHost in STOPPED state.
func NewHost(cfg RuntimeConfiguration, opts ...HostOption) (*RuntimeHost, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("runtime host config validation failed: %w", err)
	}
	h := &RuntimeHost{
		cfg:      cfg,
		status:   StatusStopped,
		wrappers: make(map[string]*ComponentWrapper),
		doneCh:   make(chan struct{}),
	}
	for _, opt := range opts {
		opt(h)
	}
	return h, nil
}

type generativeInferenceAdapter struct {
	svc     *inference.Service
	store   *storage.Storage
	modelID string
}

func (a *generativeInferenceAdapter) Generate(ctx context.Context, prompt string) (string, error) {
	h := sha256.Sum256([]byte(prompt))
	inputRef := hex.EncodeToString(h[:])
	
	if err := a.store.Write(inputRef, []byte(prompt)); err != nil {
		return "", err
	}

	req := inference.InferenceRequest{
		ModelID:  a.modelID,
		InputRef: inputRef,
		Modality: inference.ModalityText,
		Budget:   "STANDARD",
		CallerID: "OutputEngine.GenerativeRealizer",
	}

	res, err := a.svc.Execute(ctx, req)
	if err != nil {
		return "", err
	}

	outBytes, err := a.store.Read(res.OutputRef)
	if err != nil {
		return "", err
	}

	return string(outBytes), nil
}

// Configure updates runtime configuration if the host is stopped.
func (h *RuntimeHost) Configure(cfg RuntimeConfiguration) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.status != StatusStopped {
		return fmt.Errorf("%w: cannot configure in state %s", ErrInvalidTransition, h.status)
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	h.cfg = cfg
	h.built = false
	h.wired = false
	h.registered = false
	return nil
}

// Status returns the current operational state of the runtime.
func (h *RuntimeHost) Status() RuntimeStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status
}

// Manifest returns the immutable runtime provenance manifest, or nil if not yet booted.
func (h *RuntimeHost) Manifest() *RuntimeManifest {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.manifest
}

// Report returns the diagnostic startup report, or nil if not yet booted.
func (h *RuntimeHost) Report() *RuntimeBootReport {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.report
}

// Kernel returns the active Kernel instance, or nil if not running.
func (h *RuntimeHost) Kernel() *kernel.Kernel {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.kernel
}

// Workspace returns the active Global Workspace instance, or nil if not built.
func (h *RuntimeHost) Workspace() workspace.Workspace {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.workspaceSvc == nil {
		return nil
	}
	return h.workspaceSvc
}

// Build constructs every subsystem using existing constructors and dependency injection.
func (h *RuntimeHost) Build() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.status != StatusStopped && h.status != StatusStarting {
		return fmt.Errorf("%w: cannot build in state %s", ErrInvalidTransition, h.status)
	}
	if h.built {
		return nil
	}

	var logWriter io.Writer = os.Stdout
	if !h.cfg.EnableLogging {
		logWriter = io.Discard
	}
	logSvc, err := logger.NewLogger(logger.Config{Output: logWriter})
	if err != nil {
		return fmt.Errorf("failed to construct logger service: %w", err)
	}
	h.loggerSvc = logSvc

	var loc *time.Location
	if h.cfg.Timezone != "" {
		parsedLoc, err := time.LoadLocation(h.cfg.Timezone)
		if err != nil {
			return fmt.Errorf("runtime: failed to load timezone %q: %w", h.cfg.Timezone, err)
		}
		loc = parsedLoc
	} else {
		loc = time.Local
	}
	h.timeSvc = coretime.NewTimeService(loc)

	storeSvc, err := storage.NewStorage(storage.Config{Path: h.cfg.StoragePath}, logSvc)
	if err != nil {
		return fmt.Errorf("failed to construct storage service: %w", err)
	}
	h.storageSvc = storeSvc

	memSvc, err := memory.NewMemoryService(memory.Config{}, storeSvc, logSvc)
	if err != nil {
		return fmt.Errorf("failed to construct memory service: %w", err)
	}
	h.memorySvc = memSvc

	disp := &runtimeDispatcher{log: logSvc}
	schedSvc, err := scheduler.NewSchedulerService(scheduler.Config{}, memSvc, logSvc, disp)
	if err != nil {
		return fmt.Errorf("failed to construct scheduler service: %w", err)
	}
	h.schedulerSvc = schedSvc

	h.registrySvc = kernel.NewRegistry()
	h.boundarySvc = kernel.NewBoundaryEngine()
	h.permissionSvc = kernel.NewPermissionEngine()

	busSvc, err := kernel.NewBus(h.registrySvc, h.boundarySvc, h.permissionSvc)
	if err != nil {
		return fmt.Errorf("failed to construct bus service: %w", err)
	}
	h.busSvc = busSvc

	h.constGate = constitution.NewGate()
	_ = h.constGate.RegisterRule(constitution.NewMaxUrgencyEscalationRule(99))
	h.calibSvc = calibration.NewService()

	// Capability framework initialization
	capRegistry := capabilities.NewRegistry()
	capResolver := capabilities.NewResolver(capRegistry)
	capLifecycle := capabilities.NewLifecycleManager(capRegistry)
	h.capManager = capabilities.NewManager(capRegistry, capResolver, capLifecycle)

	deps := native.NativeCapabilityDependencies{
		Time:      h.timeSvc,
		Memory:    h.memorySvc,
		Storage:   storeSvc,
		Scheduler: h.schedulerSvc,
		// Workspace: h.workspaceEng, // Workspace not needed yet, or initialized later
	}
	if err := native.LoadNativeCapabilities(capRegistry, deps); err != nil {
		return fmt.Errorf("failed to load native capabilities: %w", err)
	}

	appDeps := core.AppCapabilityDependencies{
		Resolver:  core.NewCapabilityManagerResolver(h.capManager),
	}
	if err := applications.LoadApplicationCapabilities(capRegistry, appDeps); err != nil {
		return fmt.Errorf("failed to load application capabilities: %w", err)
	}
	h.capHandler = capabilities.NewActionExecutionHandler(h.capManager, &payloadStorerAdapter{store: storeSvc})

	h.workspaceSvc = workspace.NewEngine()
	h.attentionSvc = attention.NewService(
		attention.WithLogger(logSvc),
		attention.WithWorkspacePublisher(&workspacePubAdapter{ws: h.workspaceSvc}, &payloadStorerAdapter{store: storeSvc}),
	)

	if h.isSubsystemEnabled("executive") {
		// Initialize Executive V3 Adapters
		v3CapReg := executivev3.NewLegacyCapabilityRegistryAdapter(h.capManager)
		v3MemAdapter := executivev3.NewLegacyMemoryAdapter(storeSvc)
		v3PlanProv := executivev3.NewStoragePlanProvider(&payloadStorerAdapter{store: storeSvc})

		// Build V3 Execution Engine
		v3ExecEngine := executivev3.NewExecutionEngine(v3PlanProv, v3CapReg, v3MemAdapter)

		// Wrap with WorkspaceBridge
		h.execV3Svc = executivev3.NewWorkspaceBridge(
			v3ExecEngine,
			&payloadStorerAdapter{store: storeSvc},
			&workspacePubAdapter{ws: h.workspaceSvc},
			&executiveWorkspaceSubAdapter{ws: h.workspaceSvc},
		)
	}

	if h.isSubsystemEnabled("planning") {
		// Initialize Capability Registry Adapter
		v3CapReg := planningv3.NewLegacyCapabilityAdapter(h.capManager)

		// Build V3 Orchestrator
		v3PlanOrch := planningv3.NewOrchestrator(v3CapReg)

		// Wrap with WorkspaceBridge
		h.planV3Svc = planningv3.NewWorkspaceBridge(
			v3PlanOrch,
			&payloadStorerAdapter{store: storeSvc},
			&workspacePubAdapter{ws: h.workspaceSvc},
			&planningWorkspaceSubAdapter{ws: h.workspaceSvc},
		)
	}
	if h.isSubsystemEnabled("decision") {
		// Initialize Validators
		v3Safety := decisionv3.NewLegacySafetyAdapter(h.constGate)
		v3Auth := decisionv3.NewDefaultAuthValidator()
		v3Policy := decisionv3.NewDefaultPolicyValidator()
		v3Budget := decisionv3.NewDefaultBudgetValidator()

		// Build V3 Orchestrator
		v3DecOrch := decisionv3.NewOrchestrator(v3Safety, v3Auth, v3Policy, v3Budget)

		// Wrap with WorkspaceBridge
		h.decisionV3Svc = decisionv3.NewWorkspaceBridge(
			v3DecOrch,
			&payloadStorerAdapter{store: storeSvc},
			&workspacePubAdapter{ws: h.workspaceSvc},
			&decisionWorkspaceSubAdapter{ws: h.workspaceSvc},
		)
	}
	if h.isSubsystemEnabled("reflection") {
		h.reflectSvc = reflection.NewService(
			reflection.WithWorkspace(h.workspaceSvc),
		)
	}
	if h.isSubsystemEnabled("learning") {
		ls, err := learning.NewService(
			learning.WithWorkspace(h.workspaceSvc),
		)
		if err != nil {
			return fmt.Errorf("failed to construct learning service: %w", err)
		}
		h.learningSvc = ls
	}

	// Shared infrastructure & Presentation initialization
	h.modelRegSvc = registry.NewService()
	h.inferenceSvc = inference.NewService(
		inference.WithResolver(h.modelRegSvc),
		inference.WithStorage(h.storageSvc),
		inference.WithLogger(h.loggerSvc),
	)
	realizationModel := h.cfg.DefaultRealizationModel
	if realizationModel == "" {
		realizationModel = "qwen2.5:1.5b"
	}
	_ = h.modelRegSvc.Register(context.Background(), "local-realizer", registry.BackendDescriptor{
		ID:             "ollama-local-01",
		DriverScheme:   "ollama",
		Endpoint:       "http://localhost:11434",
		Version:        "1.0",
		MaxConcurrency: 4,
		DriverConfig: map[string]string{
			"model": realizationModel,
		},
	})
	_ = h.modelRegSvc.Register(context.Background(), "deliberative-parser", registry.BackendDescriptor{
		ID:             "ollama-local-01",
		DriverScheme:   "ollama",
		Endpoint:       "http://localhost:11434",
		Version:        "1.0",
		MaxConcurrency: 4,
		DriverConfig: map[string]string{
			"model": realizationModel,
		},
	})
	_ = h.modelRegSvc.Register(context.Background(), "reasoning-deliberative-llm", registry.BackendDescriptor{
		ID:             "ollama-local-01",
		DriverScheme:   "ollama",
		Endpoint:       "http://localhost:11434",
		Version:        "1.0",
		MaxConcurrency: 4,
		DriverConfig: map[string]string{
			"model": realizationModel,
		},
	})


	if h.isSubsystemEnabled("understanding") {
		// Instantiate V3 dependencies
		grammar := underv3.NewDefaultGrammarSpecialist()
		neural := underv3.NewDefaultNeuralSpecialist()
		delib := underv3.NewDeliberativeWorker(h.inferenceSvc, h.workspaceSvc, 5*time.Second)
		
		// Inject Semantic Extractors
		exts := underext.NewDeterministicExtractors()

		// Inject Normalizers
		tempNorm := undernorms.NewDeterministicTemporalNormalizer(h.timeSvc)
		norms := undernorms.NewDeterministicNormalizers(tempNorm)
		// Build V3 Orchestrator
		comps := undercomps.NewDeterministicTemporalComposer()

		v3Orch := underv3.NewOrchestrator(grammar, neural, delib, exts, norms, comps, underspl.NewDeterministicSplitter(nil))
		
		// Wrap with WorkspaceBridge
		h.underV3Svc = underv3.NewWorkspaceBridge(
			v3Orch, 
			&payloadStorerAdapter{store: h.storageSvc}, 
			&workspacePubAdapter{ws: h.workspaceSvc},
			&understandingWorkspaceSubAdapter{ws: h.workspaceSvc}, 
		)
		
		// Add Context Resolver immediately after Understanding
		resolver := intelcontext.NewDefaultContextResolver()
		
		// TODO: Replace dummyDialogueStateReader with the real DialogueStateReader 
		// once the Dialogue State Manager is implemented at the global level.
		dummyState := &dummyDialogueStateReader{}
		
		h.contextSvc = intelcontext.NewWorkspaceBridge(
			resolver,
			dummyState,
			&payloadStorerAdapter{store: h.storageSvc},
			&workspacePubAdapter{ws: h.workspaceSvc},
			&contextWorkspaceSubAdapter{ws: h.workspaceSvc},
		)
	}

	if h.isSubsystemEnabled("reasoning") {
		// Initialize Memory Adapter
		v3Mem := reasoningv3.NewLegacyMemoryAdapter(memSvc)

		// Build Deliberative Specialist (S8) using the shared InferenceService.
		// Constructed here — after h.inferenceSvc is initialized (line ~505) — so the
		// InferenceService is guaranteed non-nil when the specialist is created.
		v3Delib := reasoning.NewDeliberativeSpecialist(h.inferenceSvc)

		// Build V3 Orchestrator with all dependencies.
		v3Orch := reasoningv3.NewOrchestrator(v3Mem, v3Delib)

		// Wrap with WorkspaceBridge
		h.reasonV3Svc = reasoningv3.NewWorkspaceBridge(
			v3Orch,
			&payloadStorerAdapter{store: h.storageSvc},
			&workspacePubAdapter{ws: h.workspaceSvc},
			&reasoningWorkspaceSubAdapter{ws: h.workspaceSvc},
		)
	}



	// World boundary subsystem — registered at PhaseBackground (after all cognitive subsystems).
	// Uses os.Stdin and os.Stdout for Phase 1 text I/O. Future adapters inject alternative readers/writers.
	if h.isSubsystemEnabled("world") {
		inR := h.inReader
		if inR == nil {
			inR = os.Stdin
		}
		outW := h.outWriter
		if outW == nil {
			outW = os.Stdout
		}
		inputAdapter, inputErr := worldtext.NewTextInputAdapter(inR)
		outputAdapter, outputErr := worldtext.NewTextOutputAdapter(outW)
		if inputErr != nil {
			return fmt.Errorf("failed to construct World TextInputAdapter: %w", inputErr)
		}
		if outputErr != nil {
			return fmt.Errorf("failed to construct World TextOutputAdapter: %w", outputErr)
		}
		worldSvc, worldErr := world.NewService(
			h.workspaceSvc,
			inputAdapter,
			outputAdapter,
			&payloadStorerAdapter{store: h.storageSvc},
		)
		if worldErr != nil {
			return fmt.Errorf("failed to construct World service: %w", worldErr)
		}

		// Initialize Output Pipeline (O-Series)
		pluginRegistry := output.NewDefaultPluginRegistry()
		aggregator := output.NewDefaultAggregator()

		tmplPath := filepath.Join(h.cfg.StoragePath, "templates")
		if _, err := os.Stat(tmplPath); os.IsNotExist(err) {
			tmplPath = filepath.Join("..", "data", "runtime", "templates")
		}
		tmplEngine := formatter.NewEngine(tmplPath)

		detRealizer := engine.NewDeterministicRealizer(tmplEngine)
		genRealizer := engine.NewGenerativeRealizer(&generativeInferenceAdapter{
			svc:     h.inferenceSvc,
			store:   h.storageSvc,
			modelID: h.cfg.DefaultRealizationModel,
		})

		strat := strategy.NewDefaultStrategy(output.Descriptor{
			ResponseType: "fallback",
			Realizer:     genRealizer,
			TemplateID:   "",
		})
		
		for _, rt := range []string{"greeting", "time", "weather", "communicative", "date", "system", "notes", "calculator", "files", "reminder"} {
			strat.Register(output.ResponseType(rt), output.Descriptor{
				ResponseType: output.ResponseType(rt),
				Realizer:     detRealizer,
				TemplateID:   rt,
			})
		}

		orchestrator := output.NewOrchestratingEngine(strat)
		
		txtFmt := outtext.NewDefaultTextFormatter()
		txtPlugin := outtext.NewPlugin(txtFmt, outtext.NewTextWriterAdapter(outW))
		_ = pluginRegistry.Register(txtPlugin)

		outputManager := output.NewDefaultOutputManager(
			pluginRegistry, 
			aggregator, 
			orchestrator,
			&payloadStorerAdapter{store: h.storageSvc},
		)
		worldSvc.SetOutputDispatcher(outputManager.Dispatch)

		h.worldSvc = worldSvc
	}

	h.built = true
	return nil
}

func (h *RuntimeHost) isSubsystemEnabled(name string) bool {
	if h.cfg.EnabledSubsystems == nil {
		return true
	}
	enabled, ok := h.cfg.EnabledSubsystems[name]
	if !ok {
		return true
	}
	return enabled
}

// Wire connects subscriptions across Global Workspace channels.
func (h *RuntimeHost) Wire() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.status != StatusStopped && h.status != StatusStarting {
		return fmt.Errorf("%w: cannot wire in state %s", ErrInvalidTransition, h.status)
	}
	if !h.built {
		return errors.New("runtime: cannot wire before build")
	}
	if h.wired {
		return nil
	}

	h.subs = nil

	// Register default baseline monitoring subscriptions on Workspace if Executive is available
	if h.workspaceSvc != nil && h.execV3Svc != nil {
		sub, err := h.workspaceSvc.Subscribe(communication.TopicPerception, "runtime-perception-bridge", func(ctx context.Context, env communication.Envelope) error {
			return nil
		})
		if err == nil && sub != nil {
			h.subs = append(h.subs, sub)
		}
	}

	if h.workspaceSvc != nil && h.capHandler != nil {
		sub, err := h.workspaceSvc.Subscribe(communication.TopicActionExecution, "capability-execution-handler", h.capHandler.HandleActionExecution)
		if err == nil && sub != nil {
			h.subs = append(h.subs, sub)
		}
	}

	if h.workspaceSvc != nil && h.attentionSvc != nil {
		// Subscribe Attention to TopicCandidatePlans
		subPlans, err := h.workspaceSvc.Subscribe(communication.TopicCandidatePlans, "attention-candidate-plans", func(ctx context.Context, env communication.Envelope) error {
			stim := attention.Stimulus{
				ID:            env.ID,
				Source:        env.Source,
				PayloadRef:    env.PayloadRef,
				SalienceScore: int(env.RawConfidence * 100),
			}
			_, evalErr := h.attentionSvc.EvaluateTrace(ctx, stim)
			return evalErr
		})
		if err == nil && subPlans != nil {
			h.subs = append(h.subs, subPlans)
		}

		// Subscribe Attention to TopicEvaluatedOptions
		subOpts, err := h.workspaceSvc.Subscribe(communication.TopicEvaluatedOptions, "attention-evaluated-options", func(ctx context.Context, env communication.Envelope) error {
			stim := attention.Stimulus{
				ID:            env.ID,
				Source:        env.Source,
				PayloadRef:    env.PayloadRef,
				SalienceScore: int(env.RawConfidence * 100),
			}
			_, evalErr := h.attentionSvc.EvaluateTrace(ctx, stim)
			return evalErr
		})
		if err == nil && subOpts != nil {
			h.subs = append(h.subs, subOpts)
		}
	}

	h.wired = true
	return nil
}

// Register wraps constructed instances into phased ComponentWrappers and registers them with Kernel Registry.
func (h *RuntimeHost) Register() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.status != StatusStopped && h.status != StatusStarting {
		return fmt.Errorf("%w: cannot register in state %s", ErrInvalidTransition, h.status)
	}
	if !h.built || !h.wired {
		return errors.New("runtime: cannot register before build and wire")
	}
	if h.registered {
		return nil
	}

	h.wrappers = make(map[string]*ComponentWrapper)

	registerWrap := func(name string, phase kernel.Phase, inst interface{}) {
		if inst == nil {
			return
		}
		w := NewWrapper(name, phase, inst, nil, nil)
		h.wrappers[name] = w
		_ = h.registrySvc.Register(w)
	}

	// Phase 1: Core
	registerWrap("Core.Storage", kernel.PhaseCore, h.storageSvc)
	registerWrap("Core.Memory", kernel.PhaseCore, h.memorySvc)
	registerWrap("Core.Scheduler", kernel.PhaseCore, h.schedulerSvc)

	// Phase 2: Foundation
	registerWrap("Foundation.Registry", kernel.PhaseInfrastructure, h.registrySvc)
	registerWrap("Foundation.Bus", kernel.PhaseInfrastructure, h.busSvc)
	registerWrap("Foundation.Boundary", kernel.PhaseInfrastructure, h.boundarySvc)
	registerWrap("Foundation.Permission", kernel.PhaseInfrastructure, h.permissionSvc)
	registerWrap("Foundation.Constitution", kernel.PhaseInfrastructure, h.constGate)
	registerWrap("Foundation.Calibration", kernel.PhaseInfrastructure, h.calibSvc)
	if h.modelRegSvc != nil {
		registerWrap("Infrastructure.Registry", kernel.PhaseInfrastructure, h.modelRegSvc)
	}
	if h.inferenceSvc != nil {
		registerWrap("Infrastructure.Inference", kernel.PhaseInfrastructure, h.inferenceSvc)
	}

	// Phase 3: Workspace
	registerWrap("Intelligence.Workspace", kernel.PhaseWorkspace, h.workspaceSvc)

	// Phase 4: Executive & Attention
	registerWrap("Intelligence.Attention", kernel.PhaseExecutive, h.attentionSvc)
	if h.execV3Svc != nil {
		registerWrap("Intelligence.Executive", kernel.PhaseExecutive, h.execV3Svc)
	}

	// Phase 5: Cognitive
	if h.underV3Svc != nil {
		registerWrap("Intelligence.Understanding", kernel.PhaseCognitive, h.underV3Svc)
	}
	if h.contextSvc != nil {
		registerWrap("Intelligence.ContextResolver", kernel.PhaseCognitive, h.contextSvc)
	}
	if h.reasonV3Svc != nil {
		registerWrap("Intelligence.Reasoning", kernel.PhaseCognitive, h.reasonV3Svc)
	}
	if h.planV3Svc != nil {
		registerWrap("Intelligence.Planning", kernel.PhaseCognitive, h.planV3Svc)
	}
	if h.decisionV3Svc != nil {
		registerWrap("Intelligence.Decision", kernel.PhaseCognitive, h.decisionV3Svc)
	}

	// Phase 6: Background
	if h.reflectSvc != nil {
		registerWrap("Intelligence.Reflection", kernel.PhaseBackground, h.reflectSvc)
	}
	if h.learningSvc != nil {
		registerWrap("Intelligence.Learning", kernel.PhaseBackground, h.learningSvc)
	}
	if h.worldSvc != nil {
		registerWrap("World.Service", kernel.PhaseBackground, h.worldSvc)
	}

	h.registered = true
	return nil
}

// Start executes Build, Wire, and Register if needed, boots the Kernel across Phase 1-6,
// and produces the immutable RuntimeManifest and RuntimeBootReport.
func (h *RuntimeHost) Start(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	h.mu.Lock()
	if h.status != StatusStopped {
		status := h.status
		h.mu.Unlock()
		return fmt.Errorf("%w: cannot start in state %s", ErrInvalidTransition, status)
	}
	h.status = StatusStarting
	h.mu.Unlock()

	startTime := time.Now()

	if err := h.Build(); err != nil {
		return h.failBoot(startTime, fmt.Errorf("build stage failed: %w", err))
	}
	if err := h.Wire(); err != nil {
		return h.failBoot(startTime, fmt.Errorf("wire stage failed: %w", err))
	}
	if err := h.Register(); err != nil {
		return h.failBoot(startTime, fmt.Errorf("register stage failed: %w", err))
	}

	h.mu.Lock()
	cfg := kernel.Config{
		Registry:   h.registrySvc,
		Bus:        h.busSvc,
		Boundary:   h.boundarySvc,
		Permission: h.permissionSvc,
	}
	h.mu.Unlock()

	k, err := kernel.Boot(cfg)
	if err != nil {
		return h.failBoot(startTime, fmt.Errorf("kernel boot failed: %w", err))
	}

	h.mu.Lock()
	h.kernel = k
	h.status = StatusRunning

	subsysVersions := make(map[string]string)
	policyHashes := make(map[string]string)
	capHashes := make(map[string]string)
	startedNames := make([]string, 0, len(h.wrappers))

	for name := range h.wrappers {
		startedNames = append(startedNames, name)
		subsysVersions[name] = "2.0.0-FROZEN"
		policyHashes[name] = "policy-v2-default"
		capHashes[name] = "caps-v2-default"
	}

	var skipped []string
	for _, optional := range []string{"understanding", "reasoning", "planning", "decision", "reflection", "learning"} {
		if !h.isSubsystemEnabled(optional) {
			skipped = append(skipped, "Intelligence."+optional)
		}
	}
	if !h.isSubsystemEnabled("world") {
		skipped = append(skipped, "World.Service")
	}
	if !h.isSubsystemEnabled("realization") {
		skipped = append(skipped, "Presentation.LanguageRealization")
	}

	h.manifest = GenerateManifest(h.cfg.RuntimeVersion, subsysVersions, policyHashes, capHashes)
	h.report = &RuntimeBootReport{
		BootDuration:      time.Since(startTime),
		StartedComponents: startedNames,
		SkippedComponents: skipped,
		Manifest:          h.manifest,
		Warnings:          []string{},
		Success:           true,
	}
	h.mu.Unlock()

	devLog("Runtime", "RuntimeHost started.")

	if h.loggerSvc != nil {
		h.loggerSvc.Info("runtime: boot sequence completed successfully",
			logger.Field{Key: "duration_ms", Value: fmt.Sprintf("%d", h.report.BootDuration.Milliseconds())},
			logger.Field{Key: "manifest_hash", Value: h.manifest.ManifestFingerprint},
		)
	}

	if h.worldSvc != nil {
		worldDone := h.worldSvc.Done()
		go func() {
			select {
			case <-worldDone:
				_ = h.Stop()
			case <-h.doneCh:
			}
		}()
	}

	return nil
}

func (h *RuntimeHost) failBoot(start time.Time, err error) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.status = StatusFailed
	h.report = &RuntimeBootReport{
		BootDuration:      time.Since(start),
		StartedComponents: []string{},
		SkippedComponents: []string{},
		Warnings:          []string{err.Error()},
		Success:           false,
	}
	if h.loggerSvc != nil {
		h.loggerSvc.Error("runtime: boot sequence failed",
			logger.Field{Key: "error", Value: err.Error()},
		)
	}
	return err
}

// Stop initiates reverse-topological shutdown and transitions to STOPPED.
func (h *RuntimeHost) Stop() error {
	h.mu.Lock()
	if h.status != StatusRunning {
		h.mu.Unlock()
		return nil
	}
	h.status = StatusStopping
	subs := h.subs
	k := h.kernel
	h.mu.Unlock()

	for _, sub := range subs {
		_ = sub.Cancel()
	}

	if k != nil {
		k.Shutdown()
	}

	h.mu.Lock()
	h.status = StatusStopped
	h.kernel = nil
	if h.doneCh != nil {
		select {
		case <-h.doneCh:
		default:
			close(h.doneCh)
		}
	}
	h.mu.Unlock()

	if h.loggerSvc != nil {
		h.loggerSvc.Info("runtime: host stopped cleanly")
	}
	return nil
}

// Done returns a read-only channel that is closed when the RuntimeHost stops.
func (h *RuntimeHost) Done() <-chan struct{} {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.doneCh
}
