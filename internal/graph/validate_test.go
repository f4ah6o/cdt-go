package graph

import "testing"

func TestCoverage(t *testing.T) {
	g := &Graph{
		Version: Version,
		Nodes: []Node{
			{ID: "concept:Add", Kind: "concept", Name: "Add"},
			{ID: "doc:add", Kind: "doc"},
			{ID: "code:add", Kind: "code"},
			{ID: "test:add", Kind: "test"},
		},
		Edges: []Edge{
			{From: "doc:add", To: "concept:Add", Kind: "describes"},
			{From: "code:add", To: "concept:Add", Kind: "implements"},
			{From: "test:add", To: "concept:Add", Kind: "verifies"},
		},
	}
	if err := Validate(g); err != nil {
		t.Fatal(err)
	}
	rows := Coverage(g)
	if len(rows) != 1 || rows[0].Status != "ok" {
		t.Fatalf("unexpected coverage: %+v", rows)
	}
}

func TestValidateDuplicateNodeMessage(t *testing.T) {
	g := &Graph{Version: Version, Nodes: []Node{
		{ID: "same", Kind: "file"},
		{ID: "same", Kind: "code"},
	}}
	err := Validate(g)
	if err == nil || err.Error() != `duplicate node id "same" (kinds file and code)` {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateRejectsUnusedTestsEdgeKind(t *testing.T) {
	g := &Graph{Version: Version, Nodes: []Node{
		{ID: "test:add", Kind: "test"},
		{ID: "concept:Add", Kind: "concept"},
	}, Edges: []Edge{{From: "test:add", To: "concept:Add", Kind: "tests"}}}
	err := Validate(g)
	if err == nil {
		t.Fatal("Validate() error = nil, want unsupported edge kind")
	}
}
