package parser

import (
	"encoding/json"
)

// CanvasDocument represents a parsed Obsidian .canvas file.
type CanvasDocument struct {
	Nodes []CanvasNode `json:"nodes"`
	Edges []CanvasEdge `json:"edges"`
}

// CanvasNode represents a single node in a canvas.
type CanvasNode struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Text   string `json:"text,omitempty"`
	File   string `json:"file,omitempty"`
	URL    string `json:"url,omitempty"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Color  string `json:"color,omitempty"`
	Label  string `json:"label,omitempty"`
}

// CanvasEdge represents a connection between two canvas nodes.
type CanvasEdge struct {
	ID       string `json:"id"`
	FromNode string `json:"fromNode"`
	FromSide string `json:"fromSide,omitempty"`
	ToNode   string `json:"toNode"`
	ToSide   string `json:"toSide,omitempty"`
	Label    string `json:"label,omitempty"`
	Color    string `json:"color,omitempty"`
}

// ParseCanvas parses a .canvas JSON file into a CanvasDocument.
func ParseCanvas(content string, filePath string) (*CanvasDocument, error) {
	var doc CanvasDocument
	if err := json.Unmarshal([]byte(content), &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}
