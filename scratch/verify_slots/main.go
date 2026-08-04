// Slot extraction verification program for Phase 4B.2 closure audit.
// Run from the idun_v3 root with: go run scratch/verify_slots/main.go
package main

import (
	"fmt"
	"strings"

	"idun/intelligence/understanding"
)

type testCase struct {
	input    string
	family   string
	wantSlot string // "slot=value" or "*" for any match
	wantFail bool
}

var cases = []testCase{
	// Calculator
	{input: "divide 100 by 5", family: "Calculator", wantSlot: "operator=divide"},
	{input: "multiply 8 and 12", family: "Calculator", wantSlot: "operand1=8"},
	{input: "15 plus 20", family: "Calculator", wantSlot: "operator=plus"},
	{input: "what is 15 * 20", family: "Calculator", wantSlot: "operand2=20"},
	{input: "(15+20)*3", family: "Calculator", wantSlot: "expression=(15+20)*3"},

	// Reminder
	{input: "Remind John tomorrow at 5 PM to call Sarah", family: "Reminder", wantSlot: "person=John"},
	{input: "remind me to buy milk next Monday at 9 AM", family: "Reminder", wantSlot: "date=next Monday"},
	{input: "remind me in 2 hours to check the oven", family: "Reminder", wantSlot: "duration=2 hours"},
	{input: "remind me to pay bills tomorrow", family: "Reminder", wantSlot: "date=tomorrow"},
	{input: "remind me to feed the dog at 8 AM", family: "Reminder", wantSlot: "time=8 AM"},

	// Weather - existing
	{input: "weather in Tokyo tomorrow", family: "Weather", wantSlot: "location=Tokyo"},
	{input: "temperature in London for the next 3 days", family: "Weather", wantSlot: "duration=next 3 days"},
	{input: "weather", family: "Weather", wantSlot: "*"},
	// Weather - daypart (new)
	{input: "weather tomorrow morning", family: "Weather", wantSlot: "daypart=morning"},
	{input: "forecast Friday evening", family: "Weather", wantSlot: "daypart=evening"},
	{input: "weather afternoon", family: "Weather", wantSlot: "daypart=afternoon"},
	// Weather - previously shadowed by Calculator
	{input: "what is the forecast next Friday in New York", family: "Weather", wantSlot: "location=New York"},

	// Files - existing
	{input: "delete old_data.csv", family: "Files", wantSlot: "operation=delete"},
	{input: "rename notes.txt to old_notes.txt", family: "Files", wantSlot: "operation=rename"},
	{input: "list files", family: "Files", wantSlot: "*"},
	// Files - path (new)
	{input: `move C:\Users\John\Documents\report.pdf to C:\archive`, family: "Files/path", wantSlot: "operation=move"},
	{input: "open file /home/john/report.pdf", family: "Files/path", wantSlot: "path=/home/john/report.pdf"},
	{input: "delete C:/Projects/IDUN/report.pdf", family: "Files/path", wantSlot: "operation=delete"},

	// Notes - previously shadowed by Files
	{input: "take a note called Ideas saying build a robot", family: "Notes", wantSlot: "title=Ideas"},
	{input: "delete that note called Ideas", family: "Notes", wantSlot: "title=Ideas"},
	{input: "open the note", family: "Notes", wantSlot: "*"},
	{input: "remember that I left the keys on the table", family: "Notes", wantSlot: "content=I left the keys on the table"},

	// System - existing
	{input: "what time is it", family: "System", wantSlot: "*"},
	{input: "battery level", family: "System", wantSlot: "target=battery"},
	{input: "shutdown computer tomorrow", family: "System", wantSlot: "target=computer"},
	// System - operation + time (new)
	{input: "shutdown the computer tomorrow at 5 PM", family: "System/op", wantSlot: "operation=shutdown"},
	{input: "shutdown the computer tomorrow at 5 PM", family: "System/time", wantSlot: "time=5 PM"},
	{input: "restart the system at 3 AM", family: "System/op", wantSlot: "operation=restart"},
	{input: "lock screen", family: "System/op", wantSlot: "operation=lock"},
}

func main() {
	g := understanding.NewDefaultGrammarSpecialist()
	norm := understanding.NewDefaultNormalizer()

	pass := 0
	fail := 0

	fmt.Printf("%-55s %-15s %-30s %s\n", "INPUT", "FAMILY", "WANT", "RESULT")
	fmt.Println(strings.Repeat("-", 120))

	for _, tc := range cases {
		normalized := norm.Normalize(tc.input)
		hyp, matched := g.Evaluate(normalized, nil)

		result := "PASS"
		detail := ""

		if tc.wantFail {
			if matched {
				result = "FAIL"
				detail = fmt.Sprintf("expected no match, got intent=%s", hyp.Intent)
			}
		} else {
			if !matched {
				result = "FAIL"
				detail = "no match"
			} else if tc.wantSlot != "*" {
				parts := strings.SplitN(tc.wantSlot, "=", 2)
				wantName, wantVal := parts[0], parts[1]
				found := false
				for _, s := range hyp.Slots {
					if s.Name == wantName && strings.EqualFold(s.Value, wantVal) {
						found = true
						break
					}
				}
				if !found {
					result = "FAIL"
					slotSummary := []string{}
					for _, s := range hyp.Slots {
						slotSummary = append(slotSummary, s.Name+"="+s.Value)
					}
					detail = fmt.Sprintf("intent=%s slots=[%s]", hyp.Intent, strings.Join(slotSummary, ", "))
				}
			}
		}

		if result == "PASS" {
			pass++
		} else {
			fail++
		}

		fmt.Printf("%-55s %-15s %-30s %s  %s\n", tc.input, tc.family, tc.wantSlot, result, detail)
	}

	fmt.Println(strings.Repeat("-", 120))
	fmt.Printf("TOTAL: %d / %d PASS  |  %d FAIL\n", pass, pass+fail, fail)
}
