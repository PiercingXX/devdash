#!/bin/bash
# Download the latest release of DevDash for Linux, and install it in /usr/local/bin.
# Inspired by https://github.com/jesseduffield/lazydocker/blob/master/scripts/install_update_linux.sh
set -euo pipefail

case $(uname -m) in
    x86_64 | amd64) ARCH=x86_64 ;;
    aarch64 | arm64) ARCH=arm64 ;;
    *)
        echo "No release for the architecture $(uname -m): install DevDash with Go instead." >&2
        exit 1
        ;;
esac

VERSION=$(curl -fsSL -H 'Accept: application/json' https://github.com/PiercingXX/devdash/releases/latest | sed -e 's/.*"tag_name":"\([^"]*\)".*/\1/')
FILE="devdash_${VERSION#v}_Linux_${ARCH}.tar.gz"
URL="https://github.com/PiercingXX/devdash/releases/download/${VERSION}/${FILE}"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

echo "Downloading $URL"
curl -fsSL "$URL" -o "$TMP/devdash.tar.gz"
tar -xzf "$TMP/devdash.tar.gz" -C "$TMP" devdash
sudo install -m 755 "$TMP/devdash" /usr/local/bin/devdash
echo "DevDash $VERSION is installed in /usr/local/bin/devdash"
