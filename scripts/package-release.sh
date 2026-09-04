#!/bin/sh
set -eu

version=${1:?usage: package-release.sh VERSION [OUTPUT_DIR]}
output_dir=${2:-dist/release}
module=github.com/rudrakshkarpe/agentsmd-cli

case "$version" in
  v*) version=${version#v} ;;
esac

case "$version" in
  *[!0-9A-Za-z.-]*|'')
    echo "invalid release version: $version" >&2
    exit 1
    ;;
esac

mkdir -p "$output_dir"
rm -f "$output_dir"/agentsmd_*.tar.gz "$output_dir"/checksums.txt

for target in darwin/amd64 darwin/arm64 linux/amd64 linux/arm64; do
  os=${target%/*}
  arch=${target#*/}
  archive="agentsmd_${os}_${arch}.tar.gz"
  staging=$(mktemp -d)

  GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags "-s -w -X ${module}/cli.Version=${version}" \
    -o "$staging/agentsmd" ./cmd/agentsmd
  cp LICENSE README.md "$staging/"
  tar -C "$staging" -czf "$output_dir/$archive" agentsmd LICENSE README.md
  rm -rf "$staging"
done

(
  cd "$output_dir"
  shasum -a 256 agentsmd_*.tar.gz > checksums.txt
)
