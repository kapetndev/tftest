package exec

import (
	"context"
	"fmt"

	"github.com/hashicorp/hc-install/src"
)

// Terraformer defines the interface for a Terraform installer.
type Terraformer interface {
	Ensure(ctx context.Context) (string, error)
	Remove(ctx context.Context) error
}

// installer defines the interface for a Terraform executable installer. This is
// abstracted to allow for easier testing and future extensibility.
//
// Whilst this interface appears similar to the exported [Terraformer]
// interface, it does not imply an identical responsibility. The installer is
// to isolate external dependencys (i.e., the hc-install library) from the
// rest of the application, whereas the [Terraformer] interface is to define the
// public contract for managing Terraform installations.
type installer interface {
	Ensure(ctx context.Context, source []src.Source) (string, error)
	Remove(ctx context.Context) error
}

// Terraform represents a Terraform installer. It is effectively a wrapper
// around the HashiCorp hc-install library to manage the installation and
// removal of the Terraform executable.
type Terraform struct {
	installer installer
	sources   []src.Source
}

// NewTerraform creates a new Terraform installer.
func NewTerraform(installer installer, sources []src.Source) *Terraform {
	return &Terraform{
		installer: installer,
		sources:   sources,
	}
}

// Ensure ensures that Terraform is installed and returns the path to the
// executable.
func (o *Terraform) Ensure(ctx context.Context) (string, error) {
	execPath, err := o.installer.Ensure(ctx, o.sources)
	if err != nil {
		return "", fmt.Errorf("error finding Terraform: %w", err)
	}
	if execPath == "" {
		return "", fmt.Errorf("installer returned empty path for Terraform executable")
	}
	return execPath, nil
}

// Remove removes the installed Terraform executable.
func (o *Terraform) Remove(ctx context.Context) error {
	return o.installer.Remove(ctx)
}
