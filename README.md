# goreg

Yet another alternate `goimports` tool.

goreg is a tool for formatting Go imports while maintaining a stable order. It organizes imports into four distinct groups:

1. **Standard Library**
2. **Third-Party Libraries**
3. **Organization Modules** (optional)
4. **Local Project Imports**

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

| Option                            | Description |
|-----------------------------------|-------------|
| `-h`, `--help`                    | Show this help message and exit. |
| `-v`, `--version`                 | Show version information. |
| `-w`, `--write`                   | Write the formatted output directly to the file. (optional) |
| `-l`, `--local <local_module>`    | Specify the local module path, typically the project's module name. Used to determine whether an import is local. (optional) |
| `-o`, `--order <group_order>`     | Specify the order of import groups. Default: `"std,thirdparty,organization,local"`. Example: `"stdlib,3rd,org,local"` |
| `-n`, `--organization <org_path>` | Specify the module path of your organization. If specified, it groups imports that start with this prefix separately. (optional) |
| `-m`, `--minimize-group`          | Do not separate import groups when an alias is present. (optional) |
| `-a`, `--sort-include-alias`      | Sort imports with aliases within their respective groups. (optional) |

### Arguments

| Argument         | Description |
|------------------|-------------|
| `<file-name.go>` | The target Go file to be formatted. |
| `<local_module>` | The local module path, typically the project's module name. (optional) |
| `<org_path>`     | The organization module path. If specified, it groups imports that start with this prefix separately. (optional) |
| `<group_order>`  | Defines the order in which import groups are arranged. Must include all four: `std`, `thirdparty`, `organization`, and `local`. Example: `"stdlib,3rd,org,local"` |

## Examples

### Format a Go file and print to stdout
```sh
goreg file.go
```

### Format a Go file and overwrite it
```sh
goreg -w file.go
```

### Specify the local module path
```sh
goreg -l myproject/module file.go
```

### Set a custom import group order
```sh
goreg -o "std,org,thirdparty,local" file.go
```

### Minimize import group separation
```sh
goreg -m file.go
```

### Sort imports including aliases
```sh
goreg -a file.go
```

## See Also
- [Project Repository](https://github.com/magicdrive/goreg)
- [README.md](https://github.com/magicdrive/goreg/README.md)

## License
This project is licensed under the [MIT License](https://github.com/magicdrive/goreg/LICENSE).
