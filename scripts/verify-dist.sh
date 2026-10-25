#!/usr/bin/env bash
# Checks that a Hephaestus dist directory has exactly the release layout goreg
# has always shipped (v1.3.10 and earlier, built by GoReleaser):
#
#   goreg_<version>_<os>_<arch>.tar.gz for the eight targets below, tar.gz on
#   Windows too, each containing exactly
#     goreg (goreg.exe on Windows)               mode 0755
#     completions/goreg-completion.sh            mode 0644
#     completions/bash/goreg-completion.bash     mode 0644
#     completions/zsh/_goreg                     mode 0644
#   (plus their parent directory entries), with the completion scripts equal to
#   misc/completions, and goreg_checksums.txt listing every archive.
#
# When an archive for the host platform exists, its binary must print
# "goreg version v<version>" (skipped with VERIFY_DIST_NO_EXEC=1, used where
# release credentials are present, so that no built code runs there).
#
# Usage: scripts/verify-dist.sh <dist-dir> <version without v>
set -euo pipefail
export LC_ALL=C

if [[ $# -ne 2 ]]; then
  echo "usage: $0 <dist-dir> <version>" >&2
  exit 2
fi
dist="$1" version="$2"
repo="$(cd "$(dirname "$0")/.." && pwd)"
targets=(darwin_amd64 darwin_arm64 linux_386 linux_amd64 linux_arm64 windows_386 windows_amd64 windows_arm64)
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

fail() { echo "verify-dist: $*" >&2; exit 1; }

[[ -d "$dist" ]] || fail "$dist is not a directory"

# Exactly the expected archives, and no zip.
expected_archives="$(for t in "${targets[@]}"; do echo "goreg_${version}_${t}.tar.gz"; done | sort)"
actual_archives="$(cd "$dist" && find . -maxdepth 1 -type f \( -name '*.tar.gz' -o -name '*.zip' \) | sed 's|^\./||' | sort)"
if [[ "$actual_archives" != "$expected_archives" ]]; then
  echo "expected:"; echo "$expected_archives"; echo "actual:"; echo "$actual_archives"
  fail "unexpected set of archives"
fi

# The checksum file lists exactly these archives and matches them.
sums="$dist/goreg_checksums.txt"
[[ -f "$sums" ]] || fail "missing goreg_checksums.txt"
listed="$(awk '{print $2}' "$sums" | sort)"
[[ "$listed" == "$expected_archives" ]] || fail "goreg_checksums.txt does not list exactly the archives"
grep -Eqv '^[0-9a-f]{64}  goreg_[^ ]+\.tar\.gz$' "$sums" && fail "goreg_checksums.txt has a malformed line"
if command -v sha256sum >/dev/null; then
  (cd "$dist" && sha256sum --quiet -c goreg_checksums.txt) || fail "checksum mismatch"
else
  (cd "$dist" && shasum -a 256 -c goreg_checksums.txt >/dev/null) || fail "checksum mismatch"
fi

host_os="$(uname -s | tr '[:upper:]' '[:lower:]')"
host_arch="$(uname -m)"
case "$host_arch" in x86_64) host_arch=amd64 ;; aarch64) host_arch=arm64 ;; esac

for t in "${targets[@]}"; do
  archive="$dist/goreg_${version}_${t}.tar.gz"
  exe=goreg
  [[ "$t" == windows_* ]] && exe=goreg.exe
  # mode and path of every entry, directories marked by a trailing slash.
  listing="$(tar -tvzf "$archive" | awk '{print $1, $NF}' | sort -k2)"
  expected="$(printf '%s\n' \
    "drwxr-xr-x completions/" \
    "drwxr-xr-x completions/bash/" \
    "-rw-r--r-- completions/bash/goreg-completion.bash" \
    "-rw-r--r-- completions/goreg-completion.sh" \
    "drwxr-xr-x completions/zsh/" \
    "-rw-r--r-- completions/zsh/_goreg" \
    "-rwxr-xr-x $exe" | sort -k2)"
  if [[ "$listing" != "$expected" ]]; then
    echo "expected:"; echo "$expected"; echo "actual:"; echo "$listing"
    fail "unexpected contents in $(basename "$archive")"
  fi
  dir="$work/$t"
  mkdir -p "$dir"
  tar -xzf "$archive" -C "$dir"
  for f in goreg-completion.sh bash/goreg-completion.bash zsh/_goreg; do
    cmp -s "$dir/completions/$f" "$repo/misc/completions/$f" || fail "$t: completions/$f differs from misc/completions/$f"
  done
  if [[ -z "${VERIFY_DIST_NO_EXEC:-}" && "$t" == "${host_os}_${host_arch}" ]]; then
    out="$("$dir/$exe" --version)"
    [[ "$out" == "goreg version v$version" ]] || fail "$t: --version printed '$out', want 'goreg version v$version'"
    echo "verify-dist: $t binary: $out"
  fi
done

echo "verify-dist: ok (${#targets[@]} archives for $version, goreg_checksums.txt)"
