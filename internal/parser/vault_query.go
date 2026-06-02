package parser

import (
	"gopkg.in/yaml.v3"
)

// VaultQuery represents a parsed vault-query code block.
type VaultQuery struct {
	Type   string   `yaml:"type"`   // table, list, cards
	From   string   `yaml:"from"`   // folder prefix
	Where  map[string]string `yaml:"where"` // key=value filters
	Sort   string   `yaml:"sort"`   // field name
	Order  string   `yaml:"order"`  // desc or asc
	Limit  int      `yaml:"limit"`  // max results
	Fields []string `yaml:"fields"` // columns to display
}

// ParseVaultQuery parses the YAML content of a vault-query code block.
func ParseVaultQuery(content string) (*VaultQuery, error) {
	var q VaultQuery
	if err := yaml.Unmarshal([]byte(content), &q); err != nil {
		return nil, err
	}
	if q.Type == "" {
		q.Type = "table"
	}
	if q.Limit <= 0 {
		q.Limit = 20
	}
	if q.Order == "" {
		q.Order = "desc"
	}
	return &q, nil
}
