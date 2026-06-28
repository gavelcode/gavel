#!/bin/sh
# Gavel CLI installer.
#
#   curl -fsSL https://raw.githubusercontent.com/gavelcode/gavel/main/install.sh | sh
#
# Environment overrides:
#   GAVEL_VERSION   tag to install (default: latest release, e.g. v0.1.0)
#   GAVEL_BIN_DIR   install directory (default: /usr/local/bin, or ~/.local/bin
#                   when /usr/local/bin is not writable)
set -eu

REPO="gavelcode/gavel"

err() {
	echo "gavel-install: $*" >&2
	exit 1
}

need() {
	command -v "$1" >/dev/null 2>&1 || err "required command not found: $1"
}

need uname
need tar
if command -v curl >/dev/null 2>&1; then
	dl() { curl -fsSL "$1" -o "$2"; }
	fetch() { curl -fsSL "$1"; }
elif command -v wget >/dev/null 2>&1; then
	dl() { wget -qO "$2" "$1"; }
	fetch() { wget -qO - "$1"; }
else
	err "need curl or wget"
fi

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
	linux | darwin) ;;
	*) err "unsupported OS: $os (linux and darwin only)" ;;
esac

arch=$(uname -m)
case "$arch" in
	x86_64 | amd64) arch=amd64 ;;
	arm64 | aarch64) arch=arm64 ;;
	*) err "unsupported architecture: $arch (amd64 and arm64 only)" ;;
esac

version=${GAVEL_VERSION:-}
if [ -z "$version" ]; then
	version=$(fetch "https://api.github.com/repos/${REPO}/releases/latest" |
		grep '"tag_name"' | head -1 | cut -d'"' -f4)
	[ -n "$version" ] || err "could not determine latest release; set GAVEL_VERSION"
fi

# Archive name as produced by goreleaser: gavel_<version-without-v>_<os>_<arch>.tar.gz
nov=${version#v}
archive="gavel_${nov}_${os}_${arch}.tar.gz"
base="https://github.com/${REPO}/releases/download/${version}"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "gavel-install: downloading ${archive} (${version})"
dl "${base}/${archive}" "${tmp}/${archive}" || err "download failed: ${base}/${archive}"

# Verify checksum when the binary is available.
if command -v sha256sum >/dev/null 2>&1 || command -v shasum >/dev/null 2>&1; then
	echo "gavel-install: verifying checksum"
	fetch "${base}/checksums.txt" >"${tmp}/checksums.txt" || err "could not fetch checksums.txt"
	want=$(grep " ${archive}\$" "${tmp}/checksums.txt" | cut -d' ' -f1)
	[ -n "$want" ] || err "checksum for ${archive} not found"
	if command -v sha256sum >/dev/null 2>&1; then
		got=$(sha256sum "${tmp}/${archive}" | cut -d' ' -f1)
	else
		got=$(shasum -a 256 "${tmp}/${archive}" | cut -d' ' -f1)
	fi
	[ "$want" = "$got" ] || err "checksum mismatch: want ${want}, got ${got}"
fi

tar -xzf "${tmp}/${archive}" -C "$tmp" gavel || err "extract failed"

bindir=${GAVEL_BIN_DIR:-/usr/local/bin}
if [ ! -d "$bindir" ] || [ ! -w "$bindir" ]; then
	if [ -z "${GAVEL_BIN_DIR:-}" ]; then
		bindir="${HOME}/.local/bin"
		mkdir -p "$bindir"
	else
		err "install dir not writable: $bindir"
	fi
fi

install -m 0755 "${tmp}/gavel" "${bindir}/gavel" 2>/dev/null ||
	{ mv "${tmp}/gavel" "${bindir}/gavel" && chmod 0755 "${bindir}/gavel"; } ||
	err "could not install to ${bindir}"

echo "gavel-install: installed gavel ${version} to ${bindir}/gavel"
case ":${PATH}:" in
	*":${bindir}:"*) ;;
	*) echo "gavel-install: add ${bindir} to your PATH" ;;
esac
