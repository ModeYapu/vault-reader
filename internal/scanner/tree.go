package scanner

import (
	"sort"
	"strings"
)

// TreeNode represents a file or directory in the vault tree.
type TreeNode struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	Type     string      `json:"type"` // "dir" or "file"
	IsCanvas bool        `json:"isCanvas,omitempty"`
	Children []*TreeNode `json:"children,omitempty"`
}

// BuildTree converts a flat list of VaultFiles into a tree structure.
func BuildTree(files []VaultFile) *TreeNode {
	root := &TreeNode{
		Name:     "Vault",
		Path:     "",
		Type:     "dir",
		Children: make([]*TreeNode, 0),
	}

	for _, f := range files {
		parts := strings.Split(f.Path, "/")
		current := root

		for i, part := range parts {
			isFile := (i == len(parts)-1)

			if isFile {
				current.Children = append(current.Children, &TreeNode{
					Name:     f.Name,
					Path:     f.Path,
					Type:     "file",
					IsCanvas: f.IsCanvas,
				})
			} else {
				// Find or create directory
				found := false
				for _, child := range current.Children {
					if child.Type == "dir" && child.Name == part {
						current = child
						found = true
						break
					}
				}
				if !found {
					dirPath := strings.Join(parts[:i+1], "/")
					newDir := &TreeNode{
						Name:     part,
						Path:     dirPath,
						Type:     "dir",
						Children: make([]*TreeNode, 0),
					}
					current.Children = append(current.Children, newDir)
					current = newDir
				}
			}
		}
	}

	// Sort children: directories first, then files, alphabetically within each group
	sortNodes(root)

	return root
}

func sortNodes(node *TreeNode) {
	sort.Slice(node.Children, func(i, j int) bool {
		a, b := node.Children[i], node.Children[j]
		if a.Type != b.Type {
			return a.Type == "dir"
		}
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	})
	for _, child := range node.Children {
		if child.Type == "dir" {
			sortNodes(child)
		}
	}
}
