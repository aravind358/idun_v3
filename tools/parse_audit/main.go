package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Payload struct {
	Status           string
	PrimaryIntent    string
	PrimaryHypothesis struct {
		Slots []struct {
			Name  string
			Value string
		}
	}
	Entities []struct {
		Surface string
		Type    string
	}
	References []struct {
		Surface string
		Type    string
	}
	TemporalAnchors []struct {
		Surface string
		Type    string
	}
}

func main() {
	file, err := os.Open("robustness_audit_utf8.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentInput string
	var jsonBuffer []string
	inJson := false

	fmt.Println("Input|Slots|Entities|References|Temporal Anchors|Primary Intent|Status|Result")
	fmt.Println("---|---|---|---|---|---|---|---")

	processPayload := func(input string, data string) {
		var p Payload
		if err := json.Unmarshal([]byte(data), &p); err != nil {
			fmt.Printf("%s|ERROR|ERROR|ERROR|ERROR|ERROR|ERROR|FAIL\n", input)
			return
		}

		var slots []string
		for _, s := range p.PrimaryHypothesis.Slots {
			slots = append(slots, s.Name+"="+s.Value)
		}

		var entities []string
		for _, e := range p.Entities {
			entities = append(entities, e.Surface+"("+e.Type+")")
		}

		var refs []string
		for _, r := range p.References {
			refs = append(refs, r.Surface+"("+r.Type+")")
		}

		var temps []string
		for _, t := range p.TemporalAnchors {
			temps = append(temps, t.Surface+"("+t.Type+")")
		}

		// Evaluate Result
		result := "FAIL"
		if p.Status == "UNAMBIGUOUS" {
			result = "PASS"
			// Check if intent is unresolved
			if p.PrimaryIntent == "unresolved_intent" {
				result = "FAIL"
			}
		} else if p.Status == "FAILED_IMPASSE" {
			result = "FAIL"
		} else if p.Status == "AMBIGUOUS" {
			result = "PARTIAL"
		}

		fmt.Printf("%s|%s|%s|%s|%s|%s|%s|%s\n",
			input,
			strings.Join(slots, ", "),
			strings.Join(entities, ", "),
			strings.Join(refs, ", "),
			strings.Join(temps, ", "),
			p.PrimaryIntent,
			p.Status,
			result,
		)
	}

	for scanner.Scan() {
		line := scanner.Text()
		
		if strings.HasPrefix(line, "--- INPUT: ") {
			if inJson {
				processPayload(currentInput, strings.Join(jsonBuffer, "\n"))
				inJson = false
				jsonBuffer = nil
			}
			currentInput = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "--- INPUT: "), "---"))
		} else if strings.HasPrefix(line, "Parsed Payload:") {
			inJson = true
		} else if inJson {
			if strings.HasPrefix(line, "}") {
				jsonBuffer = append(jsonBuffer, "}")
				processPayload(currentInput, strings.Join(jsonBuffer, "\n"))
				inJson = false
				jsonBuffer = nil
			} else {
				jsonBuffer = append(jsonBuffer, line)
			}
		}
	}
	
	if inJson {
		processPayload(currentInput, strings.Join(jsonBuffer, "\n"))
	}
}
