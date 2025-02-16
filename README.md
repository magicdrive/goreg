# goreg

Yet another alternate `goimports` tool.

goreg is a tool for formatting Go imports while maintaining a stable order. It organizes imports into three distinct groups:

1. **Standard Library**
2. **Third-Party Libraries**
3. **Local Project Imports**

## Installation

You can install `goreg` using `go install`:

```sh
go install github.com/magicdrive/goreg@latest
```

Alternatively, you can download a pre-built binary from the [Releases](https://github.com/magicdrive/goreg/releases) page.

## Usage

```sh
goreg [OPTIONS] <file-name.go>
```

### Options

| Option            | Description |
|-------------------|-------------|
| `-h`, `--help`    | Show this help message and exit. |
| `-v`, `--version` | Show version information. |
| `-w`, `--write`   | Write the formatted output directly to the file. (optional)|
| `-l`, `--local`   | Specify your local modulepath. (optional)|

### Arguments

| Argument         | Description |
|------------------|-------------|
| `<file-name.go>` | Target Go file to be formatted. |

## Examples

### Format a Go file and print to stdout
```sh
goreg file.go
```

### Format a Go file and overwrite it
```sh
goreg -w file.go
```

## See Also
- [Project Repository](https://github.com/magicdrive/goreg)
- [README.md](https://github.com/magicdrive/goreg/README.md)


LICENCE
-----

[MIT License](https://github.com/magicdrive/goreg/LICENCE)
