package v3

import (
	"encoding/json"
	"idun/core/foundation"
	"testing"
	"time"
)

func TestExecutionPlan_Validation(t *testing.T) {
	artID, _ := foundation.NewUUID()
	parentID, _ := foundation.NewUUID()
	envID, _ := foundation.NewUUID()

	// Valid DAG: n1 -> n2 -> n3, n1 -> n3
	n1 := NewPlanNode("n1", CapabilityID("cap.one"), nil, "Node 1")
	n2 := NewPlanNode("n2", CapabilityID("cap.two"), nil, "Node 2")
	n3 := NewPlanNode("n3", CapabilityID("cap.three"), nil, "Node 3")

	validBuilder := NewBuilder().
		ArtifactID(foundation.ArtifactID(artID)).
		ParentArtifactID(foundation.ParentArtifactID(parentID)).
		EnvelopeID(foundation.EnvelopeID(envID)).
		Timestamp(foundation.Timestamp(time.Now())).
		Nodes([]PlanNode{n1, n2, n3}).
		Edges([]Dependency{
			NewDependency("n1", "n2"),
			NewDependency("n2", "n3"),
			NewDependency("n1", "n3"),
		})

	if _, err := validBuilder.Build(); err != nil {
		t.Errorf("expected valid DAG, got err: %v", err)
	}

	// Invalid DAG: Cycle n3 -> n1
	cycleBuilder := NewBuilder().
		ArtifactID(foundation.ArtifactID(artID)).
		ParentArtifactID(foundation.ParentArtifactID(parentID)).
		EnvelopeID(foundation.EnvelopeID(envID)).
		Timestamp(foundation.Timestamp(time.Now())).
		Nodes([]PlanNode{n1, n2, n3}).
		Edges([]Dependency{
			NewDependency("n1", "n2"),
			NewDependency("n2", "n3"),
			NewDependency("n3", "n1"),
		})

	if _, err := cycleBuilder.Build(); err == nil {
		t.Errorf("expected cycle error, got none")
	}

	// Invalid DAG: Duplicate node IDs
	dupBuilder := NewBuilder().
		ArtifactID(foundation.ArtifactID(artID)).
		ParentArtifactID(foundation.ParentArtifactID(parentID)).
		EnvelopeID(foundation.EnvelopeID(envID)).
		Timestamp(foundation.Timestamp(time.Now())).
		Nodes([]PlanNode{n1, n1})

	if _, err := dupBuilder.Build(); err == nil {
		t.Errorf("expected duplicate node error, got none")
	}
}

func TestExecutionPlan_Serialization(t *testing.T) {
	artID, _ := foundation.NewUUID()
	parentID, _ := foundation.NewUUID()
	envID, _ := foundation.NewUUID()

	n1 := NewPlanNode("n1", CapabilityID("cap.test"), map[string]any{"param1": "val1", "param2": 42.0}, "Test Node")

	plan, _ := NewBuilder().
		ArtifactID(foundation.ArtifactID(artID)).
		ParentArtifactID(foundation.ParentArtifactID(parentID)).
		EnvelopeID(foundation.EnvelopeID(envID)).
		Timestamp(foundation.Timestamp(time.Now())).
		Nodes([]PlanNode{n1}).
		Build()

	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var p2 ExecutionPlan
	if err := json.Unmarshal(data, &p2); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if p2.Nodes()[0].BoundParams()["param1"] != "val1" {
		t.Errorf("mismatch param1")
	}
	// json unmarshals numbers to float64
	if p2.Nodes()[0].BoundParams()["param2"] != 42.0 {
		t.Errorf("mismatch param2")
	}
}
