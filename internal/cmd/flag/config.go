package flag

import "github.com/spf13/pflag"

// FactoryFlags holds configuration flags. This is used to configure a
// [util.Factory].
type FactoryFlags struct {
	TerraformVersion string
}

// AddFlags adds flags to the specified FlagSet.
func (f *FactoryFlags) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&f.TerraformVersion, "terraform-version", "", "Terraform version to use (defaults to system terraform)")
}
