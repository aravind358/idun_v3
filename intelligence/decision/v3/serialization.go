package v3

import (
	"encoding/json"
	"fmt"
	"idun/core/foundation"
	"time"
)

type jsonDecisionFinding struct {
	ValidatorType string `json:"ValidatorType"`
	NodeID        string `json:"NodeID"`
	Passed        bool   `json:"Passed"`
	Code          string `json:"Code"`
	Message       string `json:"Message"`
}

type jsonDecisionRecord struct {
	SpecVersion       string                `json:"SpecVersion"`
	ArtifactID        string                `json:"ArtifactID"`
	ParentArtifactID  string                `json:"ParentArtifactID"`
	EnvelopeID        string                `json:"EnvelopeID"`
	Timestamp         time.Time             `json:"Timestamp"`
	Resolution        string                `json:"Resolution"`
	Reason            string                `json:"Reason"`
	SafetyPassed      bool                  `json:"SafetyPassed"`
	PolicyPassed      bool                  `json:"PolicyPassed"`
	BudgetPassed      bool                  `json:"BudgetPassed"`
	PermissionsPassed bool                  `json:"PermissionsPassed"`
	Findings          []jsonDecisionFinding `json:"Findings"`
}

func (d *DecisionRecord) MarshalJSON() ([]byte, error) {
	if err := d.Validate(); err != nil {
		return nil, fmt.Errorf("cannot marshal invalid DecisionRecord: %w", err)
	}

	j := jsonDecisionRecord{
		SpecVersion:       d.specVersion,
		ArtifactID:        string(d.artifactID),
		ParentArtifactID:  string(d.parentArtifactID),
		EnvelopeID:        string(d.envelopeID),
		Timestamp:         time.Time(d.timestamp),
		Resolution:        string(d.resolution),
		Reason:            d.reason,
		SafetyPassed:      d.safetyPassed,
		PolicyPassed:      d.policyPassed,
		BudgetPassed:      d.budgetPassed,
		PermissionsPassed: d.permissionsPassed,
	}

	for _, f := range d.findings {
		j.Findings = append(j.Findings, jsonDecisionFinding{
			ValidatorType: f.ValidatorType(),
			NodeID:        f.NodeID(),
			Passed:        f.Passed(),
			Code:          f.Code(),
			Message:       f.Message(),
		})
	}

	return json.Marshal(j)
}

func (d *DecisionRecord) UnmarshalJSON(data []byte) error {
	var j jsonDecisionRecord
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}

	d.specVersion = j.SpecVersion
	d.artifactID = foundation.ArtifactID(j.ArtifactID)
	d.parentArtifactID = foundation.ParentArtifactID(j.ParentArtifactID)
	d.envelopeID = foundation.EnvelopeID(j.EnvelopeID)
	d.timestamp = foundation.Timestamp(j.Timestamp)
	d.resolution = ResolutionStatus(j.Resolution)
	d.reason = j.Reason
	d.safetyPassed = j.SafetyPassed
	d.policyPassed = j.PolicyPassed
	d.budgetPassed = j.BudgetPassed
	d.permissionsPassed = j.PermissionsPassed

	d.findings = make([]DecisionFinding, len(j.Findings))
	for i, f := range j.Findings {
		d.findings[i] = NewDecisionFinding(f.ValidatorType, f.NodeID, f.Passed, f.Code, f.Message)
	}

	return d.Validate()
}
