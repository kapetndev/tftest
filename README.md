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

## Signing

The binaries on the [releases
page](https://github.com/kapetndev/tftest/releases) are signed with the signing
key `49F7948E9BD5370F9F2B68B572CEA43B14FC7612`.

### Public Key

```text
-----BEGIN PGP PUBLIC KEY BLOCK-----

mDMEaYipHxYJKwYBBAHaRw8BAQdA/l8/Ex2YIZni68fieEvQpluv/DxsamZzjbWC
H7fotui0PUthcGV0biBTaWduaW5nIChDb2RlIHNpZ25pbmcgZm9yIGthcGV0bikg
PHNpZ25pbmdAa2FwZXRuLmRldj6ImQQTFgoAQRYhBEn3lI6b1TcPnytotXLOpDsU
/HYSBQJpiKkfAhsDBQkJNToABQsJCAcCAiICBhUKCQgLAgQWAgMBAh4HAheAAAoJ
EHLOpDsU/HYSVPcBAKmon0iQ0tzByes2jA49GPkBYiJ0FZIz8pq7JRDjEqFYAQDY
aP6/+H4ndK1REOtPuyCnlmxPVSW5Wk/oQATJt4m+Brg4BGmIqR8SCisGAQQBl1UB
BQEBB0D067Q4E9MidLnOSJb9IYQ7InHvUG/TkZyP9CEbO6qrYQMBCAeIfgQYFgoA
JhYhBEn3lI6b1TcPnytotXLOpDsU/HYSBQJpiKkfAhsMBQkJNToAAAoJEHLOpDsU
/HYSVwYBAL6lXcNMhmDInuQYE5RsQa43D7vu25Bc7H2SiexsW5tdAP9xi6mWEVP0
vzHROcMpkAyC3GkBmn8FMxvvuYhwEB59Dw==
=pWQI
-----END PGP PUBLIC KEY BLOCK-----
```

### Verifying Binaries

1. Import the signing key:

   ```bash
   gpg --keyserver keyserver.ubuntu.com --recv-keys 49F7948E9BD5370F9F2B68B572CEA43B14FC7612
   ```

1. Verify that the key has been imported:

   ```bash
   gpg --list-keys signing@kapetn.dev
   ```

   Check the output is:

   ```bash
   pub   ed25519 2026-02-08 [SC] [expires: 2031-01-01]
         49F7948E9BD5370F9F2B68B572CEA43B14FC7612
   ```

1. Download the binary and its corresponding shasums and signature from the
   [releases page](/releases).

1. Verify the signature of the shasums file:

   ```bash
   gpg --verify tftest_x.y.z_SHA256SUMS.sig tftest_x.y.z_SHA256SUMS
   ```

   This should give you a similar output to that below:

<!-- pyml disable md013 -->
   ```bash
   gpg: Signature made Sun Feb  8 15:30:24 2026 GMT
   gpg:                using EDDSA key 49F7948E9BD5370F9F2B68B572CEA43B14FC7612
   gpg:                issuer "signing@kapetn.dev"
   gpg: Good signature from "Kapetn Signing (Code signing for kapetn) <signing@kapetn.dev>" [ultimate]
   ```
<!-- pyml enable md013 -->

1. Verify the shasum of the binary matches the value in the shasums file:

   ```bash
   sha256sum -c --ignore-missing tftest_x.y.z_SHA256SUMS
   ```

   This should give you a similar output to that below:

   ```bash
   tftest_x.y.z_linux_amd64.tar.gz: OK
   ```

## License

This project is licensed under the [MIT License](LICENSE).
