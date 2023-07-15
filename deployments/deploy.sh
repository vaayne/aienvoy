#!/bin/bash

# Usage: ./download_release.sh <version>

if [ $# -ne 1 ]; then
  echo "Usage: $0 <version>"
  exit 1
fi

VERSION=$1

# Download release
DOWNLOAD_URL="https://github.com/Vaayne/aienvoy/releases/download/v${VERSION}/aienvoy_Linux_x86_64.tar.gz"
DOWNLOAD_FILE="/tmp/release.tar.gz"
wget -O "${DOWNLOAD_FILE}" "${DOWNLOAD_URL}"

# Extract release
EXTRACT_PATH="/tmp/aienvoy"
mkdir -p "${EXTRACT_PATH}"
tar -xzf "${DOWNLOAD_FILE}" -C "${EXTRACT_PATH}"

# Copy binary
BINARY_FILE="${EXTRACT_PATH}/aienvoy"
DESTINATION="/usr/local/bin/aienvoy"
cp "${BINARY_FILE}" "${DESTINATION}"

echo "Downloaded and extracted release ${VERSION} to ${DESTINATION}"

# Restart odb service
systemctl restart odb
