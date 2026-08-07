package main

import (
	"fmt"
	"os"
	"strings"

	"idun/intelligence/understanding"
)

func main() {
	g := understanding.NewDefaultGrammarSpecialist()

	inputs := []string{
		// Calculator - Positive
		"divide 100 by 5",
		"multiply 8 and 12",
		"add 15 and 20",
		"subtract 5 from 10",
		"calculate 15 plus 20",
		"what is 15 * 20",
		"solve 10 / 2",
		"(15+20)*3",
		// Calculator - Negative / Edge Cases
		"divide by zero",
		"multiply something",
		"calculate the meaning of life",
		
		// Reminder - Positive
		"Remind John tomorrow at 5 PM to call Sarah",
		"set a reminder tomorrow at 5 PM to call John",
		"remind me to buy milk next Monday at 9 AM",
		"remind me in 2 hours to check the oven",
		"remind us to leave in 15 minutes",
		"remind me to pay bills tomorrow",
		"remind me to feed the dog at 8 AM",
		"remind Sarah to submit the report",
		"set reminder to take out trash",
		// Reminder - Edge Cases
		"remind me to call John about the meeting tomorrow at 5 PM",
		"remind me to tell Jane to email Mark in 3 hours",
		
		// Weather - Positive
		"weather in Tokyo tomorrow",
		"what is the forecast next Friday in New York",
		"temperature in London for the next 3 days",
		"what is the weather for the next 5 days in Paris",
		"weather in Berlin",
		"temperature tomorrow",
		"forecast for the next 2 days",
		"will it rain in Seattle tomorrow",
		"will it rain in Seattle",
		"will it rain next Monday",
		// Weather - Edge Cases
		"weather",
		"will it rain",

		// Files - Positive
		"move report.pdf to C:/archive",
		"copy document.docx to /home/user/backup",
		"rename notes.txt to old_notes.txt",
		"open file C:/projects/notes.txt",
		"delete old_data.csv",
		"create directory C:/new_folder",
		"mkdir test_folder",
		"list files in C:/projects",
		"list files",
		// Files - Edge Cases
		"move my super long filename with spaces.txt to /var/log",
		"open report", // No extension
		
		// Notes - Positive
		"take a note called Ideas saying build a robot",
		"create a note titled Shopping List saying buy milk and eggs",
		"save this named Meeting Notes saying project is delayed",
		"remember that I left the keys on the table",
		"take a note saying call mom",
		"delete that note called Ideas",
		"remove the note",
		"open note titled Shopping List",
		"read the note",
		// Notes - Edge Cases
		"take a note", 
		"take a note called Empty", // Might fail if content is required

		// System - Positive
		"what time is it",
		"date today",
		"battery level",
		"how much ram",
		"how much disk space",
		"shutdown computer tomorrow",
		"turn off system",
		"restart system next Friday",
		"reboot",
		"lock screen",
		"lock computer",
		// System - Negative
		"shutdown the toaster",
		"how much water",
	}

	reportFile, err := os.Create("reports/trace_slot_extraction_results.md")
	if err != nil {
		panic(err)
	}
	defer reportFile.Close()

	fmt.Fprintln(reportFile, "# Slot Extraction Audit Report")
	fmt.Fprintln(reportFile, "\n| Input | Intent | Extracted Slots | Status |")
	fmt.Fprintln(reportFile, "|---|---|---|---|")

	for _, input := range inputs {
		norm := understanding.NormalizedText{Cleaned: input}
		hyp, ok := g.Evaluate(norm)
		
		status := "FAIL"
		intent := "N/A"
		slotsStr := "None"
		
		if ok {
			status = "PASS"
			intent = hyp.Intent
			var slots []string
			for _, slot := range hyp.Slots {
				slots = append(slots, fmt.Sprintf("%s=`%s`", slot.Name, slot.Value))
			}
			if len(slots) > 0 {
				slotsStr = strings.Join(slots, ", ")
			}
		}

		fmt.Fprintf(reportFile, "| %s | %s | %s | %s |\n", input, intent, slotsStr, status)
	}
	fmt.Println("Trace audit complete. Results written to reports/trace_slot_extraction_results.md")
}
