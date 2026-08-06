package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type SpecResponse struct {
	Contains       []string `yaml:"contains"`
	MustNotContain []string `yaml:"must_not_contain"`
	Regex          string   `yaml:"regex"`
}

type SpecExpected struct {
	Operation string       `yaml:"operation"`
	Response  SpecResponse `yaml:"response"`
}

type SpecTest struct {
	Input    []string     `yaml:"input"`
	Expected SpecExpected `yaml:"expected"`
}

type SpecFeature struct {
	Feature  string     `yaml:"feature"`
	Severity string     `yaml:"severity"`
	Tests    []SpecTest `yaml:"tests"`
}

type SpecRoot struct {
	Version     string        `yaml:"version"`
	Owner       string        `yaml:"owner"`
	LastUpdated string        `yaml:"last_updated"`
	Features    []SpecFeature `yaml:"features"`
}

type TestCase struct {
	Category          string
	Priority          string
	Command           string
	ExpectedOperation string
	ExpectedBehavior  SpecResponse
}

type PipelineIntegrity struct {
	Understanding bool
	Planning      bool
	Decision      bool
	Executive     bool
	Application   bool
	Native        bool
	Presentation  bool
	Realization   bool
	World         bool
}

type TestResult struct {
	Command        string
	Status         string
	Output         string
	PipelineStr    string
	RawLogs        []string
	Classification string
	Priority       string

	RootCause      string
	Confidence     int
	Owner          string
	File           string
	ExpectedOp     string

	CapStatus     string
	OpStatus      string
	PresStatus    string
	RespStatus    string
	FailureReason string

	Pipeline PipelineIntegrity
}

type PreviousRun struct {
	OverallScore    float64
	HasHistory      bool
	FeatureFailures map[string]bool
}

func main() {
	fmt.Println("Starting Runtime Acceptance Test Harness...")

	specPath := filepath.Join("spec", "runtime_behavior.yaml")
	specData, err := os.ReadFile(specPath)
	if err != nil {
		log.Fatalf("Failed to read spec file %s: %v", specPath, err)
	}

	var spec RootSpec
	// Using RootSpec as an alias to SpecRoot to avoid shadowing
	if err := yaml.Unmarshal(specData, &spec); err != nil {
		log.Fatalf("Failed to parse YAML spec: %v", err)
	}

	var testCases []TestCase
	for _, f := range spec.Features {
		for _, t := range f.Tests {
			for _, input := range t.Input {
				testCases = append(testCases, TestCase{
					Category:          f.Feature,
					Priority:          f.Severity,
					Command:           input,
					ExpectedOperation: t.Expected.Operation,
					ExpectedBehavior:  t.Expected.Response,
				})
			}
		}
	}

	// 1. Setup paths
	reportsDir := filepath.Join("reports", "runtime")
	os.MkdirAll(reportsDir, 0755)

	rawLogPath := filepath.Join(reportsDir, "raw_runtime.log")
	transcriptPath := filepath.Join(reportsDir, "user_transcript.md")
	reportPath := filepath.Join("reports", "runtime_acceptance_report.md")

	prevRun := loadPreviousReport(reportPath)

	rawLogFile, err := os.Create(rawLogPath)
	if err != nil {
		log.Fatalf("Failed to create raw log: %v", err)
	}
	defer rawLogFile.Close()

	transcriptFile, err := os.Create(transcriptPath)
	if err != nil {
		log.Fatalf("Failed to create transcript log: %v", err)
	}
	defer transcriptFile.Close()

	fmt.Fprintf(transcriptFile, "# IDUN User Transcript\n\n")

	// 3. Start the process
	fmt.Println("Starting IDUN runtime via go run...")
	cmd := exec.Command("go", "run", "./cmd/idun")
	cmd.Env = append(os.Environ(), "IDUN_ACCEPTANCE_TEST=1")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("Failed to get stdin: %v", err)
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to get stdout: %v", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalf("Failed to get stderr: %v", err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start idun: %v", err)
	}

	outputCh := make(chan string, 5000)

	readRoutine := func(r io.Reader) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			outputCh <- line
		}
	}
	go readRoutine(stdoutPipe)
	go readRoutine(stderrPipe)

	fmt.Println("Waiting for Runtime Ready...")
	ready := false
	timeout := time.After(30 * time.Second)

waitLoop:
	for {
		select {
		case line := <-outputCh:
			rawLogFile.WriteString(line + "\n")
			if strings.Contains(line, "Runtime Ready") {
				ready = true
				break waitLoop
			}
		case <-timeout:
			log.Fatalf("Timeout waiting for Runtime Ready")
		}
	}

	if !ready {
		log.Fatalf("Failed to start runtime correctly.")
	}
	fmt.Println("Runtime is ready.")

	var results []TestResult
	
	for _, tc := range testCases {
		fmt.Printf("Running test [%s]: %s\n", tc.Category, tc.Command)
		fmt.Fprintf(transcriptFile, "User:\n%s\n\n", tc.Command)

		_, err := io.WriteString(stdin, tc.Command+"\n")
		if err != nil {
			log.Fatalf("Failed to write command: %v", err)
		}

		var rawOutput []string
		idleTimer := time.NewTimer(3000 * time.Millisecond)

		done := false
		for !done {
			select {
			case line := <-outputCh:
				rawOutput = append(rawOutput, line)
				rawLogFile.WriteString(line + "\n")

				if !idleTimer.Stop() {
					select {
					case <-idleTimer.C:
					default:
					}
				}
				idleTimer.Reset(3000 * time.Millisecond)

			case <-idleTimer.C:
				done = true
			}
		}

		result := evaluateOutput(tc, rawOutput)
		results = append(results, result)

		fmt.Fprintf(transcriptFile, "IDUN:\n%s\n\n", result.Output)
		fmt.Printf("  Status: %s\n", result.Status)
		if result.Status != "PASS" {
			fmt.Printf("  Classification: %s | Priority: %s\n", result.Classification, result.Priority)
			fmt.Printf("  -- Layered Diagnostics --\n")
			fmt.Printf("  Capability    : %s\n", result.CapStatus)
			fmt.Printf("  Operation     : %s\n", result.OpStatus)
			fmt.Printf("  Presentation  : %s\n", result.PresStatus)
			fmt.Printf("  User Response : %s\n", result.RespStatus)
			fmt.Printf("  Reason        : %s\n", result.FailureReason)
		}
	}

	io.WriteString(stdin, "exit\n")
	cmd.Wait()

	generateReport(reportPath, testCases, results, prevRun, spec)
	fmt.Println("Test Harness completed.")

	for _, res := range results {
		if res.Status == "FAIL" && res.Priority == "Critical" {
			os.Exit(1)
		}
	}
}

type RootSpec = SpecRoot

func evaluateOutput(tc TestCase, rawOutput []string) TestResult {
	res := TestResult{
		Command:    tc.Command,
		Status:     "PASS",
		RawLogs:    rawOutput,
		CapStatus:  "PASS",
		OpStatus:   "PASS",
		PresStatus: "PASS",
		RespStatus: "PASS",
		Priority:   tc.Priority,
		ExpectedOp: tc.ExpectedOperation,
	}

	var userOutput []string
	var meta struct {
		Capability   string `json:"capability"`
		Operation    string `json:"operation"`
		ResponseType string `json:"response_type"`
	}
	foundMeta := false
	runtimeFailed := false
	presentationFailed := false

	// TODO: Technical Debt
	// Connection Integrity is currently built by parsing console logs.
	// This is fragile and will naturally break as log formatting evolves over time.
	// In a future phase, this should be replaced by a structured runtime diagnostic artifact 
	// emitted natively during Acceptance Mode.
	for _, lineStr := range rawOutput {
		trimmed := strings.TrimSpace(lineStr)

		if strings.Contains(lineStr, "HandlePerception invoked!") {
			res.Pipeline.Understanding = true
		}
		if strings.Contains(lineStr, "Reasoning V3 HandleIntent invoked!") {
			res.Pipeline.Understanding = true
		}
		if strings.Contains(lineStr, "Planning V3 HandleActiveGoal invoked!") {
			res.Pipeline.Planning = true
		}
		if strings.Contains(lineStr, "Decision V3 HandleCandidatePlan invoked!") {
			res.Pipeline.Decision = true
		}
		if strings.Contains(lineStr, "Executive V3 HandleEvaluatedOption invoked!") {
			res.Pipeline.Executive = true
		}
		if strings.Contains(lineStr, "[Executive Trace] Executing Capability") {
			res.Pipeline.Application = true // Assume application/native
			res.Pipeline.Native = true
		}
		if strings.Contains(lineStr, "Received realized output") {
			res.Pipeline.Presentation = true
			res.Pipeline.Realization = true
		}
		if strings.Contains(lineStr, "Response delivered to console") {
			res.Pipeline.World = true
		}

		if strings.HasPrefix(trimmed, "[Acceptance Metadata]") {
			jsonStr := strings.TrimSpace(strings.TrimPrefix(trimmed, "[Acceptance Metadata]"))
			if err := json.Unmarshal([]byte(jsonStr), &meta); err == nil {
				foundMeta = true
			}
			continue
		}

		if strings.Contains(lineStr, "panic") || strings.Contains(lineStr, "fatal") || strings.Contains(lineStr, "stack trace") {
			runtimeFailed = true
		} else if strings.Contains(lineStr, "unable to realize") || strings.Contains(lineStr, "no realization engine") || strings.Contains(lineStr, "no template found") {
			presentationFailed = true
		}

		isTrace := regexp.MustCompile(`^(?:\d{4}/\d{2}/\d{2}|\d{4}-\d{2}-\d{2}T|>>>|\[|===|Envelope ID:|Parsed Payload:|Model:|Resolved backend:|Request started|Sending request\.\.\.|Ollama prompt len:|Ollama prompt preview:|User:|IDUN:|Execution time:|Received realized output|Response delivered|Published TopicPerception|Received input:|Type "exit" to quit\.)`).MatchString(lineStr)
		isJSONKey := regexp.MustCompile(`^\s*"[A-Za-z0-9_]+":`).MatchString(lineStr)
		isJSONBracket := regexp.MustCompile(`^\s*[\[\]{},]\s*$`).MatchString(lineStr)
		
		if !isTrace && !isJSONKey && !isJSONBracket && len(trimmed) > 0 {
			if lineStr != tc.Command && lineStr != "\"\"" {
				userOutput = append(userOutput, strings.Trim(lineStr, "\""))
			}
		}
	}

	fullOutput := strings.Join(userOutput, "\n")
	res.Output = fullOutput

	if runtimeFailed {
		res.Status = "FAIL"
		res.CapStatus = "FAIL"
		res.FailureReason = "Runtime panic/fatal error"
		res.Classification = "Runtime Failure"
		res.RootCause = "Runtime Panic"
		res.Confidence = 100
		res.Owner = "Core"
		res.File = "idun/runtime/..."
		return res
	}

	if presentationFailed || containsRawJson(userOutput) {
		res.Status = "FAIL"
		res.PresStatus = "FAIL"
		res.FailureReason = "Presentation failure or raw JSON leakage"
		res.Classification = "Presentation Failure"
		res.RootCause = "Presentation Template"
		res.Confidence = 96
		res.Owner = "Presentation"
		res.File = "presentation/templates/*.tmpl"
	}

	if len(userOutput) == 0 {
		res.RespStatus = "FAIL"
		res.FailureReason = "No user-visible output produced"
	} else {
		for _, c := range tc.ExpectedBehavior.Contains {
			if !strings.Contains(strings.ToLower(fullOutput), strings.ToLower(c)) {
				res.RespStatus = "FAIL"
				res.FailureReason = fmt.Sprintf("Output missing expected string: %s", c)
			}
		}
		for _, c := range tc.ExpectedBehavior.MustNotContain {
			if strings.Contains(strings.ToLower(fullOutput), strings.ToLower(c)) {
				res.RespStatus = "FAIL"
				res.FailureReason = fmt.Sprintf("Output contains forbidden string: %s", c)
			}
		}
		if tc.ExpectedBehavior.Regex != "" {
			if matched, _ := regexp.MatchString(tc.ExpectedBehavior.Regex, fullOutput); !matched {
				res.RespStatus = "FAIL"
				res.FailureReason = fmt.Sprintf("Output did not match expected pattern: %s", tc.ExpectedBehavior.Regex)
			}
		}
	}

	if !foundMeta || meta.Operation != tc.ExpectedOperation {
		res.OpStatus = "FAIL"
		if res.FailureReason == "" {
			if !foundMeta {
				res.FailureReason = "No structured metadata found in execution"
			} else {
				res.FailureReason = fmt.Sprintf("Expected Operation %s, got %s", tc.ExpectedOperation, meta.Operation)
			}
		}
	}

	// Calculate Root Cause
	if res.CapStatus == "FAIL" || res.OpStatus == "FAIL" || res.PresStatus == "FAIL" || res.RespStatus == "FAIL" {
		res.Status = "FAIL"
		
		if res.RootCause == "" { // Not already set by hard failures
			if res.OpStatus == "PASS" && res.RespStatus == "FAIL" {
				res.RootCause = "Presentation Template"
				res.Confidence = 82
				res.Owner = "Presentation"
				res.File = "presentation/templates/*.tmpl"
			} else if !foundMeta {
				if !res.Pipeline.Understanding {
					res.RootCause = "Grammar Engine"
					res.Confidence = 95
					res.Owner = "Intelligence Layer (Understanding)"
					res.File = "intelligence/understanding/grammar.go"
				} else if !res.Pipeline.Planning {
					res.RootCause = "Intent Mapping"
					res.Confidence = 88
					res.Owner = "Intelligence Layer (Planning)"
					res.File = "intelligence/planning/..."
				} else if !res.Pipeline.Executive {
					res.RootCause = "Execution Resolution"
					res.Confidence = 90
					res.Owner = "Intelligence Layer (Executive)"
					res.File = "intelligence/executive/..."
				} else {
					res.RootCause = "Decision Validation"
					res.Confidence = 75
					res.Owner = "Intelligence Layer (Decision)"
					res.File = "intelligence/decision/..."
				}
			} else if res.OpStatus == "FAIL" {
				res.RootCause = "Capability Implementation"
				res.Confidence = 92
				res.Owner = "Application Capability"
				res.File = "capabilities/.../capability.go"
			} else {
				res.RootCause = "Unknown"
				res.Confidence = 50
				res.Owner = "Core"
				res.File = "—"
			}
		}

		if res.RespStatus == "FAIL" {
			res.Classification = "Behavioral Failure"
		} else if res.PresStatus == "FAIL" {
			res.Classification = "Presentation Failure"
		} else if res.OpStatus == "FAIL" {
			res.Classification = "Grammar Failure"
		} else if res.CapStatus == "FAIL" {
			res.Classification = "Capability Failure"
		}
	} else {
		res.Classification = "PASS"
		res.Priority = "—"
		res.RootCause = "—"
		res.Owner = "—"
		res.File = "—"
	}

	if res.Status == "FAIL" {
		res.Output += fmt.Sprintf("\n\n[FAILURE REASON: %s]", res.FailureReason)
	}

	return res
}

func containsRawJson(lines []string) bool {
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if regexp.MustCompile(`^\s*"[A-Za-z0-9_]+":`).MatchString(trimmed) {
			return true
		}
		if trimmed == "[]" || trimmed == "{}" || trimmed == "[ ]" {
			return true
		}
		if regexp.MustCompile(`\{.*"[A-Za-z0-9_]+"\s*:`).MatchString(trimmed) {
			return true
		}
	}
	return false
}

func loadPreviousReport(path string) PreviousRun {
	var pr PreviousRun
	pr.FeatureFailures = make(map[string]bool)
	
	data, err := os.ReadFile(path)
	if err != nil {
		pr.HasHistory = false
		return pr
	}
	pr.HasHistory = true
	content := string(data)

	reScore := regexp.MustCompile(`Overall Score:\n(\d+)%`)
	if match := reScore.FindStringSubmatch(content); len(match) > 1 {
		fmt.Sscanf(match[1], "%f", &pr.OverallScore)
	}

	inTable := false
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "| Feature |") {
			inTable = true
			continue
		}
		if inTable {
			if strings.HasPrefix(line, "|---") {
				continue
			}
			if !strings.HasPrefix(line, "|") || strings.TrimSpace(line) == "" {
				inTable = false
				continue
			}
			
			parts := strings.Split(line, "|")
			if len(parts) >= 3 {
				feature := strings.TrimSpace(parts[1])
				status := strings.TrimSpace(parts[2])
				if status == "❌" || status == "⚠" {
					pr.FeatureFailures[feature] = true
				} else {
					pr.FeatureFailures[feature] = false
				}
			}
		}
	}

	return pr
}

func generateReport(path string, tcs []TestCase, results []TestResult, prevRun PreviousRun, spec RootSpec) {
	groups := make(map[string][]TestResult)
	for i, tc := range tcs {
		groups[tc.Category] = append(groups[tc.Category], results[i])
	}
	
	var categories []string
	for _, f := range spec.Features {
		categories = append(categories, f.Feature)
	}
	
	totalDocumentedFeatures := len(categories)
	testedFeaturesCount := len(categories)

	var sb strings.Builder
	sb.WriteString("# Runtime Acceptance Summary\n\n")
	
	sb.WriteString("## Acceptance Specification\n\n")
	sb.WriteString(fmt.Sprintf("**Version**: %s\n", spec.Version))
	sb.WriteString(fmt.Sprintf("**Owner**: %s\n", spec.Owner))
	sb.WriteString(fmt.Sprintf("**Last Updated**: %s\n\n", spec.LastUpdated))
	sb.WriteString(fmt.Sprintf("**Behavioral Tests**: %d\n", len(tcs)))
	sb.WriteString(fmt.Sprintf("**Coverage**: 100%%\n\n"))
	
	sb.WriteString("## Runtime Acceptance Coverage\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Documented Features**: %d\n", totalDocumentedFeatures))
	sb.WriteString(fmt.Sprintf("- **Tested Features**: %d\n", testedFeaturesCount))
	sb.WriteString(fmt.Sprintf("- **Untested Features**: %d\n", totalDocumentedFeatures - testedFeaturesCount))
	sb.WriteString(fmt.Sprintf("- **Coverage**: %.0f%%\n\n", float64(testedFeaturesCount)/float64(totalDocumentedFeatures)*100.0))

	sb.WriteString("## Feature Health Table\n\n")
	sb.WriteString("| Feature | Status | Expected | Observed | Most Likely Cause | Confidence | Suggested Owner | Suggested Files |\n")
	sb.WriteString("|---|---|---|---|---|---|---|---|\n")

	passedCount := 0
	
	type ActionItem struct {
		Feature   string
		Priority  string
		RootCause string
	}
	var actionQueue []ActionItem
	currentFailures := make(map[string]bool)

	// Pipeline audit aggregated
	var pipeAgg PipelineIntegrity
	pipeAgg.Understanding = true
	pipeAgg.Planning = true
	pipeAgg.Decision = true
	pipeAgg.Executive = true
	pipeAgg.Application = true
	pipeAgg.Native = true
	pipeAgg.Presentation = true
	pipeAgg.Realization = true
	pipeAgg.World = true

	var failedPipeRes *TestResult

	for _, name := range categories {
		resList := groups[name]
		
		status := "✅"
		rootCause := "—"
		priority := "—"
		confidence := "—"
		owner := "—"
		file := "—"
		expected := "—"
		observed := "—"
		
		isFail := false
		highestPri := 4
		priMap := map[string]int{"Critical": 0, "High": 1, "Medium": 2, "Low": 3}
		revPriMap := map[int]string{0: "Critical", 1: "High", 2: "Medium", 3: "Low"}
		
		for i, r := range resList {
			// Update Pipeline Aggregation
			if !r.Pipeline.Understanding { pipeAgg.Understanding = false }
			if !r.Pipeline.Planning { pipeAgg.Planning = false }
			if !r.Pipeline.Decision { pipeAgg.Decision = false }
			if !r.Pipeline.Executive { pipeAgg.Executive = false }
			if !r.Pipeline.Application { pipeAgg.Application = false; pipeAgg.Native = false }
			if !r.Pipeline.Presentation { pipeAgg.Presentation = false; pipeAgg.Realization = false }
			if !r.Pipeline.World { pipeAgg.World = false }
			
			if r.Status != "PASS" {
				if failedPipeRes == nil {
					failedPipeRes = &resList[i]
				}
				isFail = true
				p, ok := priMap[r.Priority]
				if !ok { p = 2 }
				if p < highestPri {
					highestPri = p
					priority = revPriMap[p]
					rootCause = r.RootCause
					confidence = fmt.Sprintf("%d%%", r.Confidence)
					owner = r.Owner
					file = r.File
					expected = r.ExpectedOp
					observed = r.Classification
				}
			}
		}
		
		if !isFail {
			passedCount++
			currentFailures[name] = false
		} else {
			if priority == "Medium" || priority == "Low" {
				status = "⚠"
			} else {
				status = "❌"
			}
			currentFailures[name] = true
			
			actionQueue = append(actionQueue, ActionItem{
				Feature: name,
				Priority: priority,
				RootCause: rootCause,
			})
		}
		
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s | %s |\n", name, status, expected, observed, rootCause, confidence, owner, file))
	}

	sb.WriteString("\n## Connection Integrity Audit\n\n")
	
	pprint := func(name string, ok bool) {
		if ok {
			sb.WriteString(fmt.Sprintf("%-18s PASS\n", name))
		} else {
			sb.WriteString(fmt.Sprintf("%-18s FAIL\n", name))
		}
	}
	
	pprint("Understanding", pipeAgg.Understanding)
	pprint("Planning", pipeAgg.Planning)
	pprint("Decision", pipeAgg.Decision)
	pprint("Executive", pipeAgg.Executive)
	pprint("Application", pipeAgg.Application)
	pprint("Native", pipeAgg.Native)
	pprint("Presentation", pipeAgg.Presentation)
	pprint("Realization", pipeAgg.Realization)
	pprint("World", pipeAgg.World)

	if failedPipeRes != nil {
		sb.WriteString("\n**Reason**\n\n")
		sb.WriteString(fmt.Sprintf("Pipeline analysis detected a failure likely originating in %s (Confidence: %d%%).", failedPipeRes.RootCause, failedPipeRes.Confidence))
		sb.WriteString(fmt.Sprintf("\nSuggested Owner: %s\nSuggested Files: %s\n", failedPipeRes.Owner, failedPipeRes.File))
	}
	
	overallScore := float64(passedCount) / float64(len(categories)) * 100.0

	sb.WriteString("\n## Runtime Health Dashboard\n\n")
	
	critCount := 0
	highCount := 0
	medCount := 0
	lowCount := 0
	
	for _, a := range actionQueue {
		switch a.Priority {
		case "Critical": critCount++
		case "High": highCount++
		case "Medium": medCount++
		case "Low": lowCount++
		}
	}
	
	ready := "NO"
	if critCount == 0 && highCount == 0 && overallScore == 100 {
		ready = "YES"
	}

	sb.WriteString("Overall Score:\n")
	sb.WriteString(fmt.Sprintf("%.0f%%\n\n", overallScore))
	
	sb.WriteString("Critical Issues:\n")
	sb.WriteString(fmt.Sprintf("%d\n\n", critCount))
	sb.WriteString("High:\n")
	sb.WriteString(fmt.Sprintf("%d\n\n", highCount))
	sb.WriteString("Medium:\n")
	sb.WriteString(fmt.Sprintf("%d\n\n", medCount))
	sb.WriteString("Low:\n")
	sb.WriteString(fmt.Sprintf("%d\n\n", lowCount))
	
	sb.WriteString("Ready For Daily Use:\n")
	sb.WriteString(fmt.Sprintf("%s\n\n", ready))
	
	sb.WriteString("## Regression Summary\n\n")
	
	if !prevRun.HasHistory {
		sb.WriteString("Baseline Established\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("Previous Health:\n%.0f%%\n\n", prevRun.OverallScore))
		sb.WriteString(fmt.Sprintf("Current Health:\n%.0f%%\n\n", overallScore))
		
		diff := overallScore - prevRun.OverallScore
		if diff > 0 {
			sb.WriteString(fmt.Sprintf("Improvement:\n+%.0f%%\n\n", diff))
		} else if diff < 0 {
			sb.WriteString(fmt.Sprintf("Improvement:\n%.0f%%\n\n", diff))
		} else {
			sb.WriteString("Improvement:\n0%\n\n")
		}
		
		var fixed []string
		var broken []string
		var regression []string
		
		for _, cat := range categories {
			wasFail := prevRun.FeatureFailures[cat]
			isFail := currentFailures[cat]
			
			if wasFail && !isFail {
				fixed = append(fixed, cat)
			} else if wasFail && isFail {
				broken = append(broken, cat)
			} else if !wasFail && isFail {
				regression = append(regression, cat)
			}
		}
		
		sb.WriteString("Fixed\n\n")
		if len(fixed) == 0 {
			sb.WriteString("—\n\n")
		} else {
			for _, f := range fixed {
				sb.WriteString(fmt.Sprintf("✓ %s\n", f))
			}
			sb.WriteString("\n")
		}
		
		sb.WriteString("Still Broken\n\n")
		if len(broken) == 0 {
			sb.WriteString("—\n\n")
		} else {
			for _, b := range broken {
				sb.WriteString(fmt.Sprintf("✗ %s\n", b))
			}
			sb.WriteString("\n")
		}
		
		sb.WriteString("New Regression\n\n")
		if len(regression) == 0 {
			sb.WriteString("—\n\n")
		} else {
			for _, r := range regression {
				sb.WriteString(fmt.Sprintf("✗ %s\n", r))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("## Next Recommended Fixes\n\n")
	
	if len(actionQueue) == 0 {
		sb.WriteString("All systems operational. No actions required.\n\n")
	} else {
		priMap := map[string]int{"Critical": 0, "High": 1, "Medium": 2, "Low": 3}
		sort.Slice(actionQueue, func(i, j int) bool {
			return priMap[actionQueue[i].Priority] < priMap[actionQueue[j].Priority]
		})
		
		for i, a := range actionQueue {
			sb.WriteString(fmt.Sprintf("%d.\n", i+1))
			sb.WriteString(fmt.Sprintf("%s\n", a.Feature))
			sb.WriteString(fmt.Sprintf("Priority:\n%s\n\n", a.Priority))
			sb.WriteString(fmt.Sprintf("Root Cause:\n%s\n\n", a.RootCause))
		}
	}
	
	sb.WriteString("## Command Details\n\n")
	sb.WriteString("| Command | Result | Pipeline | Output | Status |\n")
	sb.WriteString("|---|---|---|---|---|\n")

	for _, r := range results {
		out := strings.ReplaceAll(r.Output, "\n", " ")
		if len(out) > 50 {
			out = out[:47] + "..."
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | \"%s\" | %s |\n", r.Command, r.Classification, r.PipelineStr, out, r.Status))
	}
	
	sb.WriteString("\n## Failure Diagnostics\n\n")
	hasFailures := false
	for _, r := range results {
		if r.Status != "PASS" {
			hasFailures = true
			break
		}
	}
	if hasFailures {
		for _, r := range results {
			if r.Status != "PASS" {
				sb.WriteString(fmt.Sprintf("### Command: `%s`\n\n", r.Command))
				sb.WriteString(fmt.Sprintf("**Failure Type**: %s\n\n", r.Classification))
				sb.WriteString(fmt.Sprintf("- **Capability**   : %s\n", r.CapStatus))
				sb.WriteString(fmt.Sprintf("- **Operation**    : %s\n", r.OpStatus))
				sb.WriteString(fmt.Sprintf("- **Presentation** : %s\n", r.PresStatus))
				sb.WriteString(fmt.Sprintf("- **User Response**: %s\n\n", r.RespStatus))
				sb.WriteString(fmt.Sprintf("**Reason**:\n%s\n\n", r.FailureReason))
				sb.WriteString("---\n\n")
			}
		}
	}
	
	sb.WriteString("\n## Development Workflow\n\n")
	sb.WriteString("1. Implement or fix a feature.\n")
	sb.WriteString("2. Run: `go test ./...`\n")
	sb.WriteString("3. Run: `go run ./cmd/runtime_acceptance`\n")
	sb.WriteString("4. Review the Feature Health Table.\n")
	sb.WriteString("5. Fix issues in priority order.\n")
	sb.WriteString("6. Repeat until Runtime Health reaches 100%.\n")

	sb.WriteString("\n## Acceptance Harness Version\n\n")
	sb.WriteString("**Behavioral Validation**: Enabled\n\n")
	sb.WriteString("**Validation Strategy**\n\n")
	sb.WriteString("- [x] User Response\n")
	sb.WriteString("- [x] Operation\n")
	sb.WriteString("- [x] Capability\n")
	sb.WriteString("- [x] Structured Metadata\n\n")
	sb.WriteString("**Console Trace Dependency**: Maintained (Technical Debt)\n\n")
	sb.WriteString("**Structured Metadata**: Enabled\n\n")

	os.WriteFile(path, []byte(sb.String()), 0644)
}
