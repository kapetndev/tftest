package discovery_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/kapetndev/tftest/internal/discovery"
)

func TestSingle_FindModules(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		rootPath string
		want     []string
		wantErr  bool
	}{
		"returns root path": {
			rootPath: ".",
			want:     []string{"."},
		},
		"returns custom path": {
			rootPath: "/path/to/module",
			want:     []string{"/path/to/module"},
		},
		"returns relative path": {
			rootPath: "modules/app",
			want:     []string{"modules/app"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			finder := discovery.NewSingleModuleFinder()

			got, err := finder.FindModules(context.Background(), tt.rootPath)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}
			if !tt.wantErr {
				if diff := cmp.Diff(got, tt.want); diff != "" {
					t.Errorf("mismatch (-got +want):\n%s", diff)
				}
			}
		})
	}
}
