package util

import (
	"github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"

	"github.com/kapetndev/tftest/internal/cmd/flag"
	"github.com/kapetndev/tftest/internal/exec"
)

// Factory provides methods to create instances of various services.
type Factory interface {
	Terraform() exec.Terraformer
}

type factoryImpl struct {
	flags *flag.FactoryFlags
}

// NewFactory creates a new [Factory] instance.
func NewFactory(flags *flag.FactoryFlags) Factory {
	return &factoryImpl{
		flags: flags,
	}
}

// Terraform creates a new [exec.Terraformer] responsilbe for manageing
// installations of the Terraform executable.
func (f *factoryImpl) Terraform() exec.Terraformer {
	installer := install.NewInstaller()
	sources := f.terraformSources()
	return exec.NewTerraform(installer, sources)
}

func (f *factoryImpl) terraformSources() []src.Source {
	tfVersion := f.flags.TerraformVersion
	if tfVersion == "" {
		return []src.Source{
			&fs.AnyVersion{
				Product: &product.Terraform,
			},
		}
	}

	if tfVersion == "latest" {
		return []src.Source{
			&releases.LatestVersion{
				Product: product.Terraform,
			},
		}
	}

	version := version.Must(version.NewVersion(tfVersion))
	return []src.Source{
		&fs.ExactVersion{
			Product: product.Terraform,
			Version: version,
		},
		&releases.ExactVersion{
			Product: product.Terraform,
			Version: version,
		},
	}
}
