package exec_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hashicorp/hc-install/src"

	"github.com/kapetndev/tftest/internal/exec"
)

// mockInstaller implements the installer interface for testing.
type mockInstaller struct {
	ensurePath string
	ensureErr  error
	removeErr  error
}

func (m *mockInstaller) Ensure(ctx context.Context, sources []src.Source) (string, error) {
	return m.ensurePath, m.ensureErr
}

func (m *mockInstaller) Remove(ctx context.Context) error {
	return m.removeErr
}

func TestTerraform_Ensure(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		installer       *mockInstaller
		sources         []src.Source
		want            string
		wantErr         bool
		wantMsgContains string
	}{
		"successful installation": {
			installer: &mockInstaller{
				ensurePath: "/usr/local/bin/terraform",
			},
			sources: []src.Source{},
			want:    "/usr/local/bin/terraform",
		},
		"installation with multiple sources": {
			installer: &mockInstaller{
				ensurePath: "/usr/local/bin/terraform",
			},
			sources: []src.Source{nil, nil},
			want:    "/usr/local/bin/terraform",
		},
		"installer returns error": {
			installer: &mockInstaller{
				ensureErr: errors.New("download failed"),
			},
			sources:         []src.Source{},
			wantErr:         true,
			wantMsgContains: "error finding Terraform",
		},
		"installer returns empty path": {
			installer: &mockInstaller{
				ensurePath: "",
			},
			sources:         []src.Source{},
			wantErr:         true,
			wantMsgContains: "installer returned empty path for Terraform executable",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			terraform := exec.NewTerraform(tt.installer, tt.sources)

			got, err := terraform.Ensure(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("mismatch:\n  got:  %q\n  want: %q", got, tt.want)
			}

			if tt.wantErr && tt.wantMsgContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantMsgContains) {
					t.Errorf("error = %v, should contain %q", err, tt.wantMsgContains)
				}
			}
		})
	}
}

func TestTerraform_Ensure_ContextCancellation(t *testing.T) {
	t.Parallel()

	installer := &mockInstaller{
		ensureErr: context.Canceled,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	terraform := exec.NewTerraform(installer, []src.Source{})

	_, err := terraform.Ensure(ctx)
	if err == nil {
		t.Fatal("expected error due to cancelled context, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got: %v", err)
	}
}

func TestTerraform_Remove(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		installer *mockInstaller
		wantErr   bool
	}{
		"successful removal": {
			installer: &mockInstaller{},
		},
		"removal fails": {
			installer: &mockInstaller{
				removeErr: errors.New("failed to remove executable"),
			},
			wantErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			terraform := exec.NewTerraform(tt.installer, []src.Source{})

			err := terraform.Remove(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestTerraform_Remove_ContextCancellation(t *testing.T) {
	t.Parallel()

	installer := &mockInstaller{
		removeErr: context.Canceled,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	terraform := exec.NewTerraform(installer, []src.Source{})

	err := terraform.Remove(ctx)
	if err == nil {
		t.Fatal("expected error due to cancelled context, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got: %v", err)
	}
}
