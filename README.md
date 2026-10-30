# goreg

Yet another alternate `goimports` tool.

goreg is a tool for formatting Go imports while maintaining a stable order. It organizes imports into four distinct groups:

1. **Standard Library**
2. **Third-Party Libraries**
3. **Organization Modules** (optional)
4. **Local Project Imports**

## Installation

### Homebrew (macOS, Linux)

```sh
brew install magicdrive/tap/goreg
brew upgrade goreg                             # later, to update it
```

### Install script (macOS and Linux, amd64 and arm64)

Download the script, read it if you like, then run it with `sh`:

```sh
curl -fsSLO https://raw.githubusercontent.com/magicdrive/goreg/main/install.sh
sh install.sh                                  # the latest release, into ~/.local/bin
"$HOME/.local/bin/goreg" --version             # prints: goreg version vX.Y.Z
```

```sh
sh install.sh                                  # install the latest release
sh install.sh --version v1.3.12                # install a given release (older ones too)
sh install.sh --upgrade                        # switch to the latest stable release
sh install.sh --upgrade --version v1.3.12      # switch to a given release (also a downgrade)
sh install.sh --dir "$HOME/.local/bin"         # choose the directory (this is the default)
sh install.sh --help
```

- **Where**: `$HOME/.local/bin/goreg` by default; `--dir DIR` installs `DIR/goreg` and creates
  `DIR` if needed (spaces in `DIR` are fine). It never uses `sudo` and never changes `PATH` or
  shell start-up files: if the script says the directory is not in `PATH`, add it, e.g.
  `export PATH="$HOME/.local/bin:$PATH"` in `~/.profile`. For a system directory such as
  `/usr/local/bin`, run it as a user who can write there (it does not ask for `sudo` itself).
- **Versions**: without `--version` it installs the latest stable release (pre-releases are not
  picked). `--version` takes a release tag, `v1.3.12` or `1.3.12`, newer or older than the
  installed one; a version that does not exist is an error, never another version. `--upgrade`
  is the same run, named for what you mean. Running it again with the installed version changes
  nothing. Every goreg release can be installed (v0.0.1 and later).
- **Updating**: run `sh install.sh --upgrade` again (with the same `--dir` if you used one).
- **Checks**: the archive must match the SHA-256 listed in the release's checksum file
  (`goreg_checksums.txt`); a missing or malformed checksum file stops the install. Only the
  `goreg` binary is taken out of the archive, after its entries have been checked (no absolute
  paths, `..`, links or devices). The new binary must run and report the version asked for
  before it replaces the old one in a single rename: a failed or interrupted run leaves the
  installed `goreg` as it was.
- **Security**: every request starts at github.com, over https, and redirects (GitHub sends the
  downloads to its asset host) are followed only to https. The checksum file comes from the same GitHub release as the archive, so it catches a
  broken or altered download, but not a release that was itself replaced on GitHub. For a
  stronger check, verify the archive's build provenance with the
  [GitHub CLI](https://cli.github.com/) before you trust it:
  `gh attestation verify goreg_<version>_<os>_<arch>.tar.gz -R magicdrive/goreg`
  (releases from v1.3.11 on have these attestations).
  The command above runs the script from the `main` branch; to run a fixed script, download it
  from a release tag instead (`https://raw.githubusercontent.com/magicdrive/goreg/<tag>/install.sh`;
  the tags from v1.4.0 on have it). Piping it straight into `sh` (`curl ... | sh`) also works, but then you
  run what you have not seen.
- **Homebrew**: the script refuses Homebrew's directories (its prefix, e.g. `/opt/homebrew`, or
  `/usr/local` on an Intel Mac) and does not replace a symbolic link such as one Homebrew made.
  Keep one way of installing: with Homebrew, upgrade with `brew upgrade goreg`; the script warns
  if another `goreg` comes first in `PATH`.
- **Uninstall**: delete the file: `rm "$HOME/.local/bin/goreg"` (or `DIR/goreg`).
- **Needs**: a POSIX `sh`, `curl` (or GNU Wget 1.17+; BusyBox wget is refused because it does
  not check certificates, and wget2 is not supported), `tar`, `gzip`, and `sha256sum`, `shasum`
  or `openssl`. A proxy set in `https_proxy` is used. `INSTALLER_TIMEOUT` (seconds, default 60)
  is how long it waits for a connection or more data.
- **Windows**: the script does not run on Windows (Git Bash, MSYS2 and Cygwin are refused). Use
  the `goreg_<version>_windows_<arch>.tar.gz` of the [Releases page](https://github.com/magicdrive/goreg/releases) (it holds `goreg.exe`). Inside WSL, the Linux instructions apply.
- **Troubleshooting**: *not writable* — choose a directory of your own with `--dir`;
  *not in PATH* — see above; *does not run in DIR (noexec)* — choose a directory on another file
  system; *SHA-256 mismatch* or *rejected* — nothing was installed; try again later and report it
  if it persists; *release not found* — check the version on the
  [Releases page](https://github.com/magicdrive/goreg/releases); *inside the Homebrew prefix* —
  use `brew upgrade goreg`, or another `--dir`.

### From source

```sh
git clone https://github.com/magicdrive/goreg.git
cd goreg
make build                                     # needs Go (see go.mod); the binary is build/goreg
```

### Pre-built binaries

Download an archive from the [Releases](https://github.com/magicdrive/goreg/releases) page:
`goreg_<version>_<os>_<arch>.tar.gz` for macOS (amd64, arm64), Linux (386, amd64, arm64) and
Windows (386, amd64, arm64), with `goreg_checksums.txt` (SHA-256) to check them against.

## Usage

```sh
goreg [OPTIONS] <file-name.go>
goreg init
```

### Subcommands

| Subcommand | Description |
|------------|-------------|
| `init`     | Create a default `goreg.toml` configuration file in the current directory. |

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
| `-r`, `--remove-import-comment`   | Remove the comments in the import. (optional) |

### Arguments

| Argument         | Description |
|------------------|-------------|
| `<file-name.go>` | The target Go file to be formatted. |
| `<local_module>` | The local module path, typically the project's module name. (optional) |
| `<org_path>`     | The organization module path. If specified, it groups imports that start with this prefix separately. (optional) |
| `<group_order>`  | Defines the order in which import groups are arranged. Must include all four: `std`, `thirdparty`, `organization`, and `local`. Example: `"stdlib,3rd,org,local"` |

### Enviroments

| Enviroment                 | Description |
|----------------------------|-------------|
| `GOREG_NOT_USE_CONFIGFILE` | if anything other than `""` is set, goreg.toml will not be searched for. |

## Configuration

goreg supports configuration via a `goreg.toml` file. This allows users to define default options instead of specifying them via CLI arguments every time.

### Creating a Configuration File

You can quickly create a default `goreg.toml` configuration file using the `init` subcommand:

```sh
goreg init
```

This will create a `goreg.toml` file in the current directory with default settings. If a `goreg.toml` file already exists, the command will not overwrite it and will display an error message.

### Example `goreg.toml`

```toml
[import]
local_module = "myproject"  # Defines the local module path. If blank, it will be automatically guessed.
organization_module = "github.com/myorg"  # Defines the organization's module path.
order = "std,thirdparty,organization,local"  # Specifies the order of import groups.

[format]
minimize_group = false  # Do not separate import groups when an alias is present.
sort_include_alias = false  # Sort imports with aliases within their respective groups.
remove_import_comment = false  # Remove comments in the import.
```

### Using `goreg.toml`

goreg will automatically search for `goreg.toml` in the current directory and its parent directories. If no configuration file is found, it will check `~/.config/goreg/goreg.toml` as a fallback.

To override settings from the configuration file, you can specify options via CLI arguments.

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

## Author

Copyright (c) 2025 Hiroshi IKEGAMI

## License

This project is licensed under the [MIT License](https://github.com/magicdrive/goreg/LICENSE).


