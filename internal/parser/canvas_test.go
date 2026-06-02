package parser

import (
	"testing"
)

func TestParseCanvasBasic(t *testing.T) {
	input := `{
	  "nodes": [
	    {
	      "id": "node1",
	      "type": "text",
	      "text": "Hello World",
	      "x": 0,
	      "y": 0,
	      "width": 300,
	      "height": 200
	    }
	  ],
	  "edges": []
	}`

	doc, err := ParseCanvas(input, "test.canvas")
	if err != nil {
		t.Fatalf("ParseCanvas failed: %v", err)
	}

	if len(doc.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(doc.Nodes))
	}
	n := doc.Nodes[0]
	if n.ID != "node1" {
		t.Errorf("node ID = %q, want node1", n.ID)
	}
	if n.Type != "text" {
		t.Errorf("node type = %q, want text", n.Type)
	}
	if n.Text != "Hello World" {
		t.Errorf("node text = %q, want Hello World", n.Text)
	}
	if n.X != 0 || n.Y != 0 || n.Width != 300 || n.Height != 200 {
		t.Errorf("node position = %d,%d %dx%d", n.X, n.Y, n.Width, n.Height)
	}
}

func TestParseCanvasWithEdges(t *testing.T) {
	input := `{
	  "nodes": [
	    {"id": "a", "type": "text", "text": "A", "x": 0, "y": 0, "width": 100, "height": 100},
	    {"id": "b", "type": "text", "text": "B", "x": 200, "y": 0, "width": 100, "height": 100}
	  ],
	  "edges": [
	    {"id": "e1", "fromNode": "a", "fromSide": "right", "toNode": "b", "toSide": "left", "label": "connects"}
	  ]
	}`

	doc, err := ParseCanvas(input, "test.canvas")
	if err != nil {
		t.Fatalf("ParseCanvas failed: %v", err)
	}

	if len(doc.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(doc.Nodes))
	}
	if len(doc.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(doc.Edges))
	}
	e := doc.Edges[0]
	if e.FromNode != "a" || e.ToNode != "b" {
		t.Errorf("edge from=%q to=%q", e.FromNode, e.ToNode)
	}
	if e.Label != "connects" {
		t.Errorf("edge label = %q, want connects", e.Label)
	}
}

func TestParseCanvasFileNode(t *testing.T) {
	input := `{
	  "nodes": [
	    {"id": "f1", "type": "file", "file": "notes/test.md", "x": 100, "y": 100, "width": 300, "height": 200}
	  ],
	  "edges": []
	}`

	doc, err := ParseCanvas(input, "test.canvas")
	if err != nil {
		t.Fatalf("ParseCanvas failed: %v", err)
	}

	if doc.Nodes[0].Type != "file" {
		t.Errorf("node type = %q, want file", doc.Nodes[0].Type)
	}
	if doc.Nodes[0].File != "notes/test.md" {
		t.Errorf("node file = %q, want notes/test.md", doc.Nodes[0].File)
	}
}

func TestParseCanvasLinkNode(t *testing.T) {
	input := `{
	  "nodes": [
	    {"id": "l1", "type": "link", "url": "https://example.com", "x": 0, "y": 0, "width": 300, "height": 200}
	  ],
	  "edges": []
	}`

	doc, err := ParseCanvas(input, "test.canvas")
	if err != nil {
		t.Fatalf("ParseCanvas failed: %v", err)
	}

	if doc.Nodes[0].Type != "link" {
		t.Errorf("node type = %q, want link", doc.Nodes[0].Type)
	}
	if doc.Nodes[0].URL != "https://example.com" {
		t.Errorf("node url = %q", doc.Nodes[0].URL)
	}
}

func TestParseCanvasGroupNode(t *testing.T) {
	input := `{
	  "nodes": [
	    {"id": "g1", "type": "group", "label": "My Group", "x": -50, "y": -50, "width": 500, "height": 400},
	    {"id": "n1", "type": "text", "text": "Inside", "x": 0, "y": 0, "width": 200, "height": 100}
	  ],
	  "edges": []
	}`

	doc, err := ParseCanvas(input, "test.canvas")
	if err != nil {
		t.Fatalf("ParseCanvas failed: %v", err)
	}

	if doc.Nodes[0].Type != "group" {
		t.Errorf("node type = %q, want group", doc.Nodes[0].Type)
	}
	if doc.Nodes[0].Label != "My Group" {
		t.Errorf("node label = %q, want My Group", doc.Nodes[0].Label)
	}
}

func TestParseCanvasColoredEdge(t *testing.T) {
	input := `{
	  "nodes": [
	    {"id": "a", "type": "text", "text": "A", "x": 0, "y": 0, "width": 100, "height": 100},
	    {"id": "b", "type": "text", "text": "B", "x": 200, "y": 0, "width": 100, "height": 100}
	  ],
	  "edges": [
	    {"id": "e1", "fromNode": "a", "toNode": "b", "color": "red"}
	  ]
	}`

	doc, err := ParseCanvas(input, "test.canvas")
	if err != nil {
		t.Fatalf("ParseCanvas failed: %v", err)
	}

	if doc.Edges[0].Color != "red" {
		t.Errorf("edge color = %q, want red", doc.Edges[0].Color)
	}
}

func TestParseCanvasEmpty(t *testing.T) {
	input := `{"nodes": [], "edges": []}`
	doc, err := ParseCanvas(input, "test.canvas")
	if err != nil {
		t.Fatalf("ParseCanvas failed: %v", err)
	}
	if len(doc.Nodes) != 0 || len(doc.Edges) != 0 {
		t.Errorf("expected empty canvas")
	}
}

func TestParseCanvasInvalidJSON(t *testing.T) {
	_, err := ParseCanvas("not json", "test.canvas")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseCanvasColoredNode(t *testing.T) {
	input := `{
	  "nodes": [
	    {"id": "c1", "type": "text", "text": "Colored", "x": 0, "y": 0, "width": 200, "height": 100, "color": "4"}
	  ],
	  "edges": []
	}`

	doc, err := ParseCanvas(input, "test.canvas")
	if err != nil {
		t.Fatalf("ParseCanvas failed: %v", err)
	}
	if doc.Nodes[0].Color != "4" {
		t.Errorf("node color = %q, want 4", doc.Nodes[0].Color)
	}
}
