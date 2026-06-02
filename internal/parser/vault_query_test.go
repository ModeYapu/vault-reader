package parser

import (
	"testing"
)

func TestParseVaultQuery_Table(t *testing.T) {
	input := `type: table
from: 20_Debug
where:
  status: active
sort: updated
order: desc
limit: 20
fields:
  - title
  - status
  - updated`

	q, err := ParseVaultQuery(input)
	if err != nil {
		t.Fatalf("ParseVaultQuery: %v", err)
	}
	if q.Type != "table" {
		t.Errorf("type: got %q, want 'table'", q.Type)
	}
	if q.From != "20_Debug" {
		t.Errorf("from: got %q, want '20_Debug'", q.From)
	}
	if q.Where["status"] != "active" {
		t.Errorf("where.status: got %q, want 'active'", q.Where["status"])
	}
	if q.Sort != "updated" {
		t.Errorf("sort: got %q, want 'updated'", q.Sort)
	}
	if q.Order != "desc" {
		t.Errorf("order: got %q, want 'desc'", q.Order)
	}
	if q.Limit != 20 {
		t.Errorf("limit: got %d, want 20", q.Limit)
	}
	if len(q.Fields) != 3 {
		t.Fatalf("fields: got %d, want 3", len(q.Fields))
	}
	if q.Fields[0] != "title" {
		t.Errorf("fields[0]: got %q, want 'title'", q.Fields[0])
	}
}

func TestParseVaultQuery_Defaults(t *testing.T) {
	input := `from: notes`

	q, err := ParseVaultQuery(input)
	if err != nil {
		t.Fatalf("ParseVaultQuery: %v", err)
	}
	if q.Type != "table" {
		t.Errorf("default type: got %q, want 'table'", q.Type)
	}
	if q.Limit != 20 {
		t.Errorf("default limit: got %d, want 20", q.Limit)
	}
	if q.Order != "desc" {
		t.Errorf("default order: got %q, want 'desc'", q.Order)
	}
}

func TestParseVaultQuery_List(t *testing.T) {
	input := `type: list
from: 00_Inbox
limit: 5`

	q, err := ParseVaultQuery(input)
	if err != nil {
		t.Fatalf("ParseVaultQuery: %v", err)
	}
	if q.Type != "list" {
		t.Errorf("type: got %q, want 'list'", q.Type)
	}
	if q.Limit != 5 {
		t.Errorf("limit: got %d, want 5", q.Limit)
	}
}

func TestParseVaultQuery_Cards(t *testing.T) {
	input := `type: cards
fields:
  - title
  - tags`

	q, err := ParseVaultQuery(input)
	if err != nil {
		t.Fatalf("ParseVaultQuery: %v", err)
	}
	if q.Type != "cards" {
		t.Errorf("type: got %q, want 'cards'", q.Type)
	}
	if len(q.Fields) != 2 {
		t.Errorf("fields: got %d, want 2", len(q.Fields))
	}
}

func TestParseVaultQuery_InvalidYAML(t *testing.T) {
	_, err := ParseVaultQuery(":\n  :\ninvalid: [")
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}
