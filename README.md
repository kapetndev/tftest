# tftest

This README outlines the details of collaborating on this Go module. A short
introduction of this module could easily go here.

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

## License

This project is licensed under the [MIT License](LICENSE).
