package actions

import (
	"testing"

	"gopkg.in/yaml.v3"
	"github.com/prettymuchbryce/autotidy/internal/rules"
)

func TestDeserializeTrash(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name:    "bare action name",
			yaml:    "trash",
			wantErr: false,
		},
		{
			name:    "null value",
			yaml:    "trash: null",
			wantErr: false,
		},
		{
			name:    "empty mapping",
			yaml:    "trash: {}",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a rules.Action
			err := yaml.Unmarshal([]byte(tt.yaml), &a)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if a.Name != "trash" {
				t.Errorf("action name = %q, want %q", a.Name, "trash")
				return
			}

			_, ok := a.Inner.(*Trash)
			if !ok {
				t.Errorf("inner is not *Trash, got %T", a.Inner)
			}
		})
	}
}
