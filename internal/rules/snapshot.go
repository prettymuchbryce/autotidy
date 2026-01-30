package rules

import (
	"log/slog"
	"path/filepath"
	"sort"

	"github.com/spf13/afero"
)

// Node represents a file or directory in the tree.
type Node struct {
	Name     string
	Children []*Node
	IsDir    bool
}

// TraverseInstruction controls how traversal proceeds after visiting a node.
type TraverseInstruction int

const (
	// Continue normal traversal
	Continue TraverseInstruction = iota
	// SkipChildren skips recursing into children of this node
	SkipChildren
	// StopTraversing stops the entire traversal
	StopTraversing
)

// TraverseControl contains instructions for how to proceed after visiting a node.
type TraverseControl struct {
	Instruction TraverseInstruction
	NewPath     string // If set, use this path for children (BFS only)
}

// Visitor is a function that visits a node and returns traversal control and a value.
type Visitor[T any] func(path string) (TraverseControl, T, error)

// BuildSnapshot builds a snapshot of a directory tree.
// If recursive is false, only includes direct children.
func BuildSnapshot(fs afero.Fs, root string, recursive bool) *Node {
	info, err := fs.Stat(root)
	if err != nil {
		slog.Warn("failed to stat path", "path", root, "error", err)
		return nil
	}

	node := &Node{
		Name:  filepath.Base(root),
		IsDir: info.IsDir(),
	}

	if !info.IsDir() {
		return node
	}

	entries, err := afero.ReadDir(fs, root)
	if err != nil {
		slog.Warn("failed to read directory", "path", root, "error", err)
		return node
	}

	// Sort entries for deterministic order
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		childPath := filepath.Join(root, entry.Name())
		if entry.IsDir() && recursive {
			child := BuildSnapshot(fs, childPath, recursive)
			if child != nil {
				node.Children = append(node.Children, child)
			}
		} else {
			node.Children = append(node.Children, &Node{
				Name:  entry.Name(),
				IsDir: entry.IsDir(),
			})
		}
	}

	return node
}

// TraverseChildrenDFS traverses the children of a tree depth-first (children before parent).
// The root node itself is not visited, only its descendants.
// Returns all collected values from the visitor.
func TraverseChildrenDFS[T any](tree *Node, parentPath string, visit Visitor[T]) ([]T, error) {
	rootPath := filepath.Join(parentPath, tree.Name)
	var results []T
	for _, child := range tree.Children {
		childResults, err := traverseDFS(child, rootPath, visit)
		if err != nil {
			return results, err
		}
		results = append(results, childResults...)
	}
	return results, nil
}

// TraverseChildrenBFS traverses the children of a tree breadth-first (parent before children).
// The root node itself is not visited, only its descendants.
// Returns all collected values from the visitor.
func TraverseChildrenBFS[T any](tree *Node, parentPath string, visit Visitor[T]) ([]T, error) {
	rootPath := filepath.Join(parentPath, tree.Name)
	var results []T
	for _, child := range tree.Children {
		childResults, err := traverseBFS(child, rootPath, visit)
		if err != nil {
			return results, err
		}
		results = append(results, childResults...)
	}
	return results, nil
}

// traverseDFS traverses the tree depth-first (children before parent).
func traverseDFS[T any](n *Node, parentPath string, visit Visitor[T]) ([]T, error) {
	currentPath := filepath.Join(parentPath, n.Name)
	var results []T

	// DFS: recurse into children first (deepest first)
	for _, child := range n.Children {
		childResults, err := traverseDFS(child, currentPath, visit)
		if err != nil {
			return results, err
		}
		results = append(results, childResults...)
	}

	// Then visit this node
	ctrl, value, err := visit(currentPath)
	if err != nil {
		return results, err
	}

	// Append the value (even zero values - caller can filter if needed)
	results = append(results, value)

	if ctrl.Instruction == StopTraversing {
		return results, nil
	}

	return results, nil
}

// traverseBFS traverses the tree breadth-first (parent before children).
func traverseBFS[T any](n *Node, parentPath string, visit Visitor[T]) ([]T, error) {
	currentPath := filepath.Join(parentPath, n.Name)
	var results []T

	// BFS: visit this node first
	ctrl, value, err := visit(currentPath)
	if err != nil {
		return results, err
	}

	results = append(results, value)

	if ctrl.Instruction == StopTraversing {
		return results, nil
	}

	if ctrl.Instruction == SkipChildren {
		return results, nil
	}

	// Update path for children if specified
	if ctrl.NewPath != "" {
		currentPath = ctrl.NewPath
	}

	// Then recurse into children
	for _, child := range n.Children {
		childResults, err := traverseBFS(child, currentPath, visit)
		if err != nil {
			return results, err
		}
		results = append(results, childResults...)
	}

	return results, nil
}
