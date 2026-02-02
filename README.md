# tftest [![test](https://github.com/kapetndev/tftest/actions/workflows/test.yaml/badge.svg?event=push)](https://github.com/kapetndev/tftest/actions/workflows/test.yaml)

A concurrent Terraform module testing tool that validates module structure,
formatting, configuration validity, and linting across your entire repository.

## Features

- **Structure validation**: Verifies that modules exist and contain Terraform
  files.
- **Formatting checks**: Ensures all `.tf` files follow Terraform formatting
  standards.
- **Configuration validation**: Runs `terraform validate` to catch syntax and
  configuration errors.
- **Linting support**: Integrates with TFLint for advanced code quality checks.
- **Concurrent execution**: Tests multiple modules in parallel, limited by
  available CPU cores.
- **Recursive scanning**: Automatically discovers and tests all Terraform
  modules in a directory tree.
- **Streaming results**: Optionally returns test results as they complete rather
  than waiting for all tests to finish.

## Example Output

```bash
tftest -r modules
╔══════════════════════════════════════════════════════════════════════════════╗
║ Testing Module: modules/storage/gcs_bucket                                   ║
╚══════════════════════════════════════════════════════════════════════════════╗
 [PASS] All files properly formatted
 [PASS] Terraform configuration valid
╚══════════════════════════════════════════════════════════════════════════════╝
✓ modules/storage/gcs_bucket - ALL CHECKS PASSED

╔══════════════════════════════════════════════════════════════════════════════╗
║ Testing Module: modules/iam/workload_identity_pool                           ║
╚══════════════════════════════════════════════════════════════════════════════╗
 [PASS] All files properly formatted
 [PASS] Terraform configuration valid
╚══════════════════════════════════════════════════════════════════════════════╝
✓ modules/iam/workload_identity_pool - ALL CHECKS PASSED
```

## Prerequisites

You will need the following things properly installed on your computer:

- [Terraform](https://developer.hashicorp.com/terraform) (optional, will
  be installed automatically if not found in `PATH`)
- [TFLint](https://github.com/terraform-linters/tflint) (optional, for linting
  checks)

## Installation

Grab the binary for your system from the [releases
page](https://github.com/kapetndev/tftest/releases).

Alternatively, install with Go:

```bash
go install github.com/kapetndev/tftest/cmd/tftest@latest
```

> [!NOTE]
> Installing from source using `go install` will build from the master branch
> and won't report version information via `tftest --version`.

## Using

Run `tftest` against a directory containing Terraform modules:

```bash
tftest [path]
```

If no path is specified, the current working directory is used. The tool will:

1. Recursively discover all Terraform modules
1. Run checks concurrently (structure, formatting, validation, linting)
1. Stream results as each module completes
1. Exit with status 0 if all checks pass, non-zero otherwise

### Options

<!-- pyml disable md013 -->
| Flag | Short | Description | Example |
|------|-------|-------------|---------|
| `--fail-on-error` | | Exit with non-zero status if any check fails (default: false) | `tftest --fail-on-error ./modules` |
| `--format` | `-f` | Output format: `text` (default), `json`, or `prettyjson` | `tftest -f json ./modules` |
| `--out` | `-o` | Write results to file instead of stdout | `tftest -o results.json ./modules` |
| `--recursive` | `-r` | Recursively discover and test all modules in subdirectories | `tftest -r ./modules` |
| `--skip-lint` | Skip TFLint checks | `tftest --skip-lint ./modules` |
| `--stream` | `-s` | Stream results as each module completes instead of waiting for all to finish | `tftest -s ./modules` |
| `--terraform-version` | | Use a specific Terraform version instead of system default | `tftest --terraform-version=1.5.0 ./modules` |
| `--verbose` | `-v` | Display detailed Terraform output and diagnostic information | `tftest -v ./modules` |
<!-- pyml enable md013 -->

**Combining options:**

```bash
tftest -r -v -f prettyjson -o results.json ./modules
```

### GitHub Actions

Use `tftest` in your CI/CD workflows with the [dedicated
action](https://github.com/kapetndev/tftest-action).

## License

This project is licensed under the [MIT License](LICENSE).
