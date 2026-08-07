package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	coretime "idun/core/time"
	"idun/boundary/perception"
	"idun/core/foundation"
	underv3 "idun/intelligence/understanding/v3"
	underext "idun/intelligence/understanding/v3/extractors"
	undernorms "idun/intelligence/understanding/v3/normalizers"
	undercomps "idun/intelligence/understanding/v3/composers"
	underspl "idun/intelligence/understanding/v3/splitter"
)

func main() {
	v3Grammar := underv3.NewDefaultGrammarSpecialist()
	
	exts := underext.NewDeterministicExtractors()
	timeSvc := coretime.NewTimeService(time.Local)
	tempNorm := undernorms.NewDeterministicTemporalNormalizer(timeSvc)
	norms := undernorms.NewDeterministicNormalizers(tempNorm)
	comps := undercomps.NewDeterministicTemporalComposer()
	
	splitter := underspl.NewDeterministicSplitter(nil)
	orch := underv3.NewOrchestrator(v3Grammar, nil, nil, exts, norms, comps, splitter)

	corpus := []string{
		"Take a note to buy milk.",
		"Take a note saying buy milk.",
		"Note that I owe John ₹500.",
		"Create a note titled Shopping.",
		"Create a note titled Shopping saying buy milk.",
		"Read my Shopping note.",
		"Open my Shopping note.",
		"Show my note called Shopping.",
		"List my notes.",
		"Delete my Shopping note.",
		"Remove my Shopping note.",
		"Delete the note called Shopping.",
		// Negative cases
		"Take a note.",
		"Read my note.",
		"Delete my note.",
		"Create a note titled.",
		"Create a note saying.",
		// File Positive Tests
		"Create directory Test.",
		"Make a folder called Photos.",
		"List files.",
		"Show files in Documents.",
		"Open report.pdf.",
		"Delete report.pdf.",
		"Move report.pdf to Documents.",
		"Copy report.pdf to Backup.",
		"Rename report.pdf to report-old.pdf.",
		"Open docs/report.pdf",
		"Move images/logo.png to archive",
		"Open docs/../report.pdf",
		"Open ./notes.txt",
		"Open images/./logo.png",
		// File Negative Tests
		"Delete C:\\Windows\\System32",
		"Delete ../../config",
		"Move report.pdf to C:\\Windows",
		"Delete everything",
		"Delete ../../../secret.txt",
		"Open ..\\..\\Windows\\System32",
		"Copy ../../config.json",
		"Open C:\\Windows\\explorer.exe",
		"Delete C:\\Users",
		"List D:\\",
		"Delete all files",
		"Remove all folders",
		"Format drive",
		"Erase disk",
	}

	for _, text := range corpus {
		fmt.Printf("Input: %q\n", text)
		
		envID, _ := foundation.NewUUID()
		artID, _ := foundation.NewUUID()
		env, _ := perception.NewBuilder().
			ArtifactID(artID).
			EnvelopeID(envID).
			RawInput(text).
			Version("3.0").
			Timestamp(time.Now()).
			Build()
			
		batch, err := orch.Analyze(context.Background(), env)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			continue
		}
		
		interps := batch.Interpretations()
		for _, interp := range interps {
			fmt.Printf("  Intent: %s\n", interp.PrimaryIntent())
			
			// Show TemporalAnchors
			fmt.Printf("  TemporalAnchors:\n")
			for i, a := range interp.TemporalAnchors() {
				fmt.Printf("    [%d]\n", i)
				fmt.Printf("    Surface: %s\n", a.Surface())
				fmt.Printf("    Normalized: %s\n", a.Normalized())
			}
			
			// Show ComposedTimestamps
			fmt.Printf("  ComposedTimestamps:\n")
			for _, c := range interp.ComposedTimestamps() {
				fmt.Printf("    - %s\n", c)
			}
			
			// Emulate Reasoning Enrichment Logic
			fmt.Printf("  Reasoning Enriched Slots (Simulated):\n")
			for _, slot := range interp.PrimaryHypothesis().Slots() {
				enrichedVal := slot.Value()
				if len(interp.ComposedTimestamps()) > 0 && (slot.Name() == "time" || slot.Name() == "date" || slot.Name() == "temporal") {
					enrichedVal = interp.ComposedTimestamps()[0]
				} else {
					for _, anchor := range interp.TemporalAnchors() {
						if anchor.Surface() == slot.Value() && anchor.Normalized() != "" {
							enrichedVal = anchor.Normalized()
							break
						}
					}
				}
				fmt.Printf("    Slot %q: %q\n", slot.Name(), enrichedVal)
			}
		}
		fmt.Println(strings.Repeat("-", 40))
	}
}


