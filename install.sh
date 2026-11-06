#!/bin/sh
set -e

repo=ZXlinxk/nosub
dir=${NOSUB_DIR:-/usr/local/bin}

case $(uname -s) in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  *) echo "unsupported system: $(uname -s)" >&2; exit 1 ;;
esac
case $(uname -m) in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) echo "unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

name=nosub_${os}_${arch}
base=https://github.com/$repo/releases/latest/download
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

curl -fsSL "$base/$name" -o "$tmp/$name"
curl -fsSL "$base/checksums.txt" -o "$tmp/checksums.txt"
(cd "$tmp" && grep " $name\$" checksums.txt | { sha256sum -c - 2>/dev/null || shasum -a 256 -c -; } >/dev/null)

chmod +x "$tmp/$name"
if [ -w "$dir" ]; then
  mv "$tmp/$name" "$dir/nosub"
else
  sudo mv "$tmp/$name" "$dir/nosub"
fi
echo "installed to $dir/nosub"
