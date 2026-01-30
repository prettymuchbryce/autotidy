package rules

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/prettymuchbryce/autotidy/internal/testutil"
)

func TestBuildSnapshot(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(fs afero.Fs)
		path           string
		recursive      bool
		expectNil      bool
		expectedName   string
		expectedChildren []string // names of direct children, sorted
		checkNested    func(t *testing.T, node *Node) // optional deeper checks
	}{
		{
			name: "flat directory",
			setup: func(fs afero.Fs) {
				fs.MkdirAll("/root", 0755)
				afero.WriteFile(fs, "/root/a.txt", []byte("a"), 0644)
				afero.WriteFile(fs, "/root/b.txt", []byte("b"), 0644)
				afero.WriteFile(fs, "/root/c.txt", []byte("c"), 0644)
			},
			path:             "/root",
			recursive:        false,
			expectedName:     "root",
			expectedChildren: []string{"a.txt", "b.txt", "c.txt"},
		},
		{
			name: "non-recursive stops at subdirectories",
			setup: func(fs afero.Fs) {
				fs.MkdirAll("/root/sub", 0755)
				afero.WriteFile(fs, "/root/file.txt", []byte("content"), 0644)
				afero.WriteFile(fs, "/root/sub/nested.txt", []byte("nested"), 0644)
			},
			path:             "/root",
			recursive:        false,
			expectedName:     "root",
			expectedChildren: []string{"file.txt", "sub"},
			checkNested: func(t *testing.T, node *Node) {
				for _, child := range node.Children {
					if child.Name == "sub" {
						if !child.IsDir {
							t.Error("sub should be a directory")
						}
						if len(child.Children) != 0 {
							t.Errorf("non-recursive: sub should have 0 children, got %d", len(child.Children))
						}
						return
					}
				}
				t.Error("expected to find 'sub' directory")
			},
		},
		{
			name: "recursive includes nested children",
			setup: func(fs afero.Fs) {
				fs.MkdirAll("/root/sub/deeper", 0755)
				afero.WriteFile(fs, "/root/a.txt", []byte("a"), 0644)
				afero.WriteFile(fs, "/root/sub/b.txt", []byte("b"), 0644)
				afero.WriteFile(fs, "/root/sub/deeper/c.txt", []byte("c"), 0644)
			},
			path:             "/root",
			recursive:        true,
			expectedName:     "root",
			expectedChildren: []string{"a.txt", "sub"},
			checkNested: func(t *testing.T, node *Node) {
				var subDir *Node
				for _, child := range node.Children {
					if child.Name == "sub" {
						subDir = child
						break
					}
				}
				if subDir == nil {
					t.Fatal("expected to find 'sub' directory")
				}
				if len(subDir.Children) != 2 {
					t.Fatalf("sub should have 2 children, got %d", len(subDir.Children))
				}

				var deeperDir *Node
				for _, child := range subDir.Children {
					if child.Name == "deeper" {
						deeperDir = child
						break
					}
				}
				if deeperDir == nil {
					t.Fatal("expected to find 'deeper' directory")
				}
				if len(deeperDir.Children) != 1 || deeperDir.Children[0].Name != "c.txt" {
					t.Error("deeper should have one child: c.txt")
				}
			},
		},
		{
			name:      "non-existent path returns nil",
			setup:    func(fs afero.Fs) {},
			path:      "/nonexistent",
			recursive: false,
			expectNil: true,
		},
		{
			name: "empty directory",
			setup: func(fs afero.Fs) {
				fs.MkdirAll("/empty", 0755)
			},
			path:             "/empty",
			recursive:        false,
			expectedName:     "empty",
			expectedChildren: []string{},
		},
		{
			name: "children are sorted alphabetically",
			setup: func(fs afero.Fs) {
				fs.MkdirAll("/root", 0755)
				afero.WriteFile(fs, "/root/zebra.txt", []byte("z"), 0644)
				afero.WriteFile(fs, "/root/apple.txt", []byte("a"), 0644)
				afero.WriteFile(fs, "/root/mango.txt", []byte("m"), 0644)
			},
			path:             "/root",
			recursive:        false,
			expectedName:     "root",
			expectedChildren: []string{"apple.txt", "mango.txt", "zebra.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			tt.setup(fs)

			node := BuildSnapshot(fs, tt.path, tt.recursive)

			if tt.expectNil {
				if node != nil {
					t.Error("expected nil node")
				}
				return
			}

			if node == nil {
				t.Fatal("expected node, got nil")
			}
			if node.Name != tt.expectedName {
				t.Errorf("Name = %q, want %q", node.Name, tt.expectedName)
			}
			if !node.IsDir {
				t.Error("expected IsDir = true")
			}
			if len(node.Children) != len(tt.expectedChildren) {
				t.Fatalf("expected %d children, got %d", len(tt.expectedChildren), len(node.Children))
			}
			for i, expectedName := range tt.expectedChildren {
				if node.Children[i].Name != expectedName {
					t.Errorf("child[%d].Name = %q, want %q", i, node.Children[i].Name, expectedName)
				}
			}

			if tt.checkNested != nil {
				tt.checkNested(t, node)
			}
		})
	}
}

func TestTraverseChildren_Order(t *testing.T) {
	fsRoot := testutil.Path("/")
	root := testutil.Path("/", "root")
	sub := testutil.Path("/", "root", "sub")
	tests := []struct {
		name     string
		traverse func(tree *Node, parent string, visit Visitor[string]) ([]string, error)
		expected []string
	}{
		{
			name:     "DFS visits children before parent",
			traverse: TraverseChildrenDFS[string],
			expected: []string{testutil.Path("/", "root", "a.txt"), testutil.Path("/", "root", "sub", "b.txt"), sub},
		},
		{
			name:     "BFS visits parent before children",
			traverse: TraverseChildrenBFS[string],
			expected: []string{testutil.Path("/", "root", "a.txt"), sub, testutil.Path("/", "root", "sub", "b.txt")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			fs.MkdirAll(sub, 0755)
			afero.WriteFile(fs, testutil.Path("/", "root", "a.txt"), []byte("a"), 0644)
			afero.WriteFile(fs, testutil.Path("/", "root", "sub", "b.txt"), []byte("b"), 0644)

			tree := BuildSnapshot(fs, root, true)
			if tree == nil {
				t.Fatal("failed to build tree")
			}

			var visited []string
			visitor := func(path string) (TraverseControl, string, error) {
				visited = append(visited, path)
				return TraverseControl{}, path, nil
			}

			tt.traverse(tree, fsRoot, visitor)

			if len(visited) != len(tt.expected) {
				t.Fatalf("expected %d visits, got %d: %v", len(tt.expected), len(visited), visited)
			}

			for i, path := range tt.expected {
				if visited[i] != path {
					t.Errorf("visit[%d] = %q, want %q", i, visited[i], path)
				}
			}
		})
	}
}

func TestTraverseChildren_CollectsValues(t *testing.T) {
	fs := afero.NewMemMapFs()
	fs.MkdirAll("/root", 0755)
	afero.WriteFile(fs, "/root/a.txt", []byte("a"), 0644)
	afero.WriteFile(fs, "/root/b.txt", []byte("b"), 0644)

	tree := BuildSnapshot(fs, "/root", false)
	if tree == nil {
		t.Fatal("failed to build snapshot")
	}

	visitor := func(path string) (TraverseControl, string, error) {
		return TraverseControl{}, path, nil
	}

	paths, err := TraverseChildrenDFS(tree, "/", visitor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(paths) != 2 {
		t.Errorf("expected 2 paths, got %d", len(paths))
	}
}

