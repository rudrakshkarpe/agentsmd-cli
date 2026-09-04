#!/bin/sh
set -eu

version=${AGENTSMD_VERSION:-v0.1.0}
release_url=${AGENTSMD_RELEASE_URL:-https://rudrakshkarpe.com/downloads/agentsmd/${version}}
install_dir=${AGENTSMD_INSTALL_DIR:-${HOME}/.local/bin}

case "$(uname -s)" in
  Darwin) os=darwin ;;
  Linux) os=linux ;;
  *)
    echo "agentsmd: unsupported operating system: $(uname -s)" >&2
    exit 1
    ;;
esac

case "$(uname -m)" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *)
    echo "agentsmd: unsupported architecture: $(uname -m)" >&2
    exit 1
    ;;
esac

archive="agentsmd_${os}_${arch}.tar.gz"
work_dir=$(mktemp -d)
trap 'rm -rf "$work_dir"' EXIT HUP INT TERM

download() {
  label=$1
  source_url=$2
  destination=$3

  echo "Downloading $label"
  curl --fail --location --show-error \
    --progress-bar \
    --retry 3 \
    --retry-delay 2 \
    --connect-timeout 30 \
    --max-time 300 \
    "$source_url" \
    --output "$destination"
}

echo "Installing agentsmd ${version} for ${os}/${arch} from rudrakshkarpe.com."
download "$archive (about 1.3 MB)..." "$release_url/$archive" "$work_dir/$archive"
download "checksums.txt..." "$release_url/checksums.txt" "$work_dir/checksums.txt"

expected=$(awk -v name="$archive" '$2 == name || $2 == "*" name { print $1 }' "$work_dir/checksums.txt")
if [ -z "$expected" ]; then
  echo "agentsmd: checksum not found for $archive" >&2
  exit 1
fi

if command -v shasum >/dev/null 2>&1; then
  actual=$(shasum -a 256 "$work_dir/$archive" | awk '{print $1}')
elif command -v sha256sum >/dev/null 2>&1; then
  actual=$(sha256sum "$work_dir/$archive" | awk '{print $1}')
else
  echo "agentsmd: shasum or sha256sum is required" >&2
  exit 1
fi

if [ "$actual" != "$expected" ]; then
  echo "agentsmd: checksum verification failed" >&2
  exit 1
fi

tar -xzf "$work_dir/$archive" -C "$work_dir" agentsmd
mkdir -p "$install_dir"
install -m 0755 "$work_dir/agentsmd" "$install_dir/agentsmd"

echo "Installed agentsmd to $install_dir/agentsmd"
case ":${PATH}:" in
  *:"$install_dir":*) ;;
  *) echo "Add $install_dir to PATH, then open a new shell." ;;
esac
"$install_dir/agentsmd" --version
