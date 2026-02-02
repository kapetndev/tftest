package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	cliflag "github.com/tomasbasham/cli-runtime/flag"
	"github.com/tomasbasham/cli-runtime/iooption"
	"github.com/tomasbasham/cli-runtime/printer"
	"github.com/tomasbasham/cli-runtime/templates"

	"github.com/kapetndev/tftest/internal/check"
	"github.com/kapetndev/tftest/internal/cmd/flag"
	"github.com/kapetndev/tftest/internal/cmd/format"
	"github.com/kapetndev/tftest/internal/cmd/util"
	"github.com/kapetndev/tftest/internal/discovery"
	"github.com/kapetndev/tftest/internal/exec"
	"github.com/kapetndev/tftest/internal/service"
)

var (
	rootLong = templates.LongDesc(`
		Test Terraform modules by running validation and linting checks.

		This tool runs formatting checks, terraform validate, and optional TFLint
		checks against Terraform modules.`)

	rootExamples = templates.Examples(`
		# Test current directory
		tftest test .

		# Test specific module
		tftest test ./modules/network

		# Recursively test all modules
		tftest test -r ./modules

		# Verbose output
		tftest test -v ./modules/network

		# JSON output
		tftest test -f json ./modules/network`)

	// Injected at build time using ldflags.
	version = ""
	commit  = ""
)

// TestOptions defines the options for the tftest command.
type TestOptions struct {
	printerFlags *cliflag.PrinterFlags
	factoryFlags *flag.FactoryFlags
	outFile      *os.File

	FailOnError   bool
	ModulePath    string
	OutPath       string
	Recursive     bool
	SkipLint      bool
	StreamResults bool
	Terraform     exec.Terraformer
	Verbose       bool

	iooption.IOStreams
}

// NewTestOptions provides an initialised [TestOptions] instance.
func NewTestOptions(streams iooption.IOStreams) *TestOptions {
	return &TestOptions{
		printerFlags: cliflag.NewPrinterFlags(
			cliflag.FormatJSONFlag|cliflag.FormatPrettyJSONFlag|cliflag.FormatTextFlag,
			cliflag.FormatText,
		),
		factoryFlags: &flag.FactoryFlags{},
		IOStreams:    streams,
	}
}

// NewRootCommand creates the `tftest` command with default arguments.
func NewRootCommand() *cobra.Command {
	options := NewTestOptions(iooption.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	})

	factory := util.NewFactory(options.factoryFlags)
	return NewRootCommandWithArgs(factory, options)
}

// NewRootCommandWithArgs creates the `tftest` command and its nested children.
func NewRootCommandWithArgs(f util.Factory, o *TestOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "tftest [path]",
		Version:               versionInfo(),
		DisableFlagsInUseLine: true,
		Short:                 "Terraform test tool",
		Long:                  rootLong,
		Example:               rootExamples,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(f, cmd, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			return o.Run(cmd, args)
		},
	}

	printerOpts := printer.WarningPrinterOptions{Color: true}
	printer := printer.NewWarningPrinter(o.ErrOut, printerOpts)
	cmd.SetGlobalNormalizationFunc(cliflag.WarnWordSepNormalizeFunc(printer))

	// Add persistent config flags.
	pflags := cmd.PersistentFlags()
	o.printerFlags.AddFlags(pflags)
	o.factoryFlags.AddFlags(pflags)

	pflags.BoolVar(&o.FailOnError, "fail-on-error", false, "Exit with non-zero status if any check fails")
	pflags.BoolVar(&o.SkipLint, "skip-lint", false, "Skip TFLint checks")

	pflags.StringVarP(&o.OutPath, "out", "o", "", "Output file (default: stdout)")
	pflags.BoolVarP(&o.Recursive, "recursive", "r", false, "Recursively find and test all modules")
	pflags.BoolVarP(&o.StreamResults, "stream", "s", false, "Stream results as they are generated (default: false)")
	pflags.BoolVarP(&o.Verbose, "verbose", "v", false, "Enable verbose output")

	// The globlal normalisation function ensures that all flags specified meet
	// the desired format, changing users' input if necessary.
	cmd.SetGlobalNormalizationFunc(cliflag.WordSepNormalizeFunc())

	return cmd
}

// Complete sets all information required for processing the command.
func (o *TestOptions) Complete(f util.Factory, cmd *cobra.Command, args []string) error {
	o.Terraform = f.Terraform()

	// Parse module path argument.
	modulePath := "."
	if len(args) > 0 {
		modulePath = strings.TrimSpace(args[0])
	}
	o.ModulePath = modulePath

	// Setup output. If an output file is specified, create it.
	outFile := o.OutPath
	if outFile != "" {
		f, err := os.Create(outFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		o.outFile = f // store for later cleanup.
		o.Out = f
	}

	return nil
}

// Validate ensures that all required arguments and flag values are provided.
func (o *TestOptions) Validate() error {
	if len(o.ModulePath) == 0 {
		return fmt.Errorf("module must be specified")
	}

	return nil
}

// Run performs the test operation.
func (o *TestOptions) Run(cmd *cobra.Command, args []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if o.outFile != nil {
		defer o.outFile.Close()
	}

	// Ensure we have a compatible Terraform version installed.
	execPath, err := o.Terraform.Ensure(ctx)
	if err != nil {
		return fmt.Errorf("error finding Terraform: %w", err)
	}
	defer o.Terraform.Remove(ctx)

	// Prepare checks to execute.
	checks := []check.Check{
		check.NewStructureCheck(),
		check.NewFormattingCheck(execPath),
		check.NewValidationCheck(execPath),
	}
	if !o.SkipLint {
		checks = append(checks, check.NewLintingCheck())
	}

	// Select appropriate module finder strategy.
	var moduleFinder service.Finder
	if o.Recursive {
		moduleFinder = discovery.NewRecursiveModuleFinder(os.DirFS("."))
	} else {
		moduleFinder = discovery.NewSingleModuleFinder()
	}

	checker := check.NewExecutor(checks)
	svc := service.NewTestService(checker, moduleFinder)

	req := service.TestRequest{
		RootPath: o.ModulePath,
		Verbose:  o.Verbose,
	}

	printer, err := o.printerFlags.ToPrinter()
	if err != nil {
		return fmt.Errorf("failed to create printer: %w", err)
	}

	if o.StreamResults {
		return o.runWithStreaming(ctx, svc, req, printer)
	}

	results, err := svc.CollectResults(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to collect test results: %w", err)
	}

	if err := printer.Print(o.Out, format.ToModulesResult(results)); err != nil {
		return fmt.Errorf("failed to print test results: %w", err)
	}

	allPassed := check.AllPassed(results)
	if !allPassed && o.FailOnError {
		return fmt.Errorf("one or more checks failed")
	}

	return nil
}

func (o *TestOptions) runWithStreaming(ctx context.Context, svc *service.TestService, req service.TestRequest, printer printer.Printer) error {
	allPassed := true
	for results, err := range svc.Results(ctx, req) {
		if err != nil {
			return fmt.Errorf("failed to stream test results: %w", err)
		}

		if err := printer.Print(o.Out, format.ToModuleResult(results)); err != nil {
			return fmt.Errorf("failed to print test results: %w", err)
		}

		if !check.AllPassed(results) {
			allPassed = false
		}
	}

	if !allPassed && o.FailOnError {
		return fmt.Errorf("one or more checks failed")
	}

	return nil
}

func versionInfo() string {
	if version == "" {
		return ""
	}
	return fmt.Sprintf("%s (commit: %s)", version, commit)
}
