#!/usr/bin/env bash
# Script for getting one or more versions of $BIN.
# Usage: ./get_terraform.sh [version ...]
# Dependencies: curl, awk, jq, unzip

# Arguments
VERSION_ARGUMENT="$1"

# Constants
DEFAULT_VERSION="1.9.8"
BIN="terraform"
INSTALL_PATH="/usr/bin"
CACHE_PATH="/cache"
URL="https://releases.hashicorp.com/${BIN}/VERSION/${BIN}_VERSION"
URL_BIN_SUFFIX="_linux_amd64.zip"
URL_SHA_SUFFIX="_SHA256SUMS"
URL_BIN="${URL}${URL_BIN_SUFFIX}"
URL_SHA="${URL}${URL_SHA_SUFFIX}"
URL_REPLACE="VERSION"

# Check if version argument is passed.
# Override the default version if provided.
if [ -n "$VERSION_ARGUMENT" ]; then
  DEFAULT_VERSION="$VERSION_ARGUMENT"
fi

# If multiple versions are needed,
# add to space separated list.
VERSIONS="${DEFAULT_VERSION}"

# Check INSTALL_PATH is in PATH
if [[ ":$PATH:" != *":$INSTALL_PATH:"* ]]; then
  echo "Error: $INSTALL_PATH is not in PATH"
  exit 1
fi

# If Actions Runner has $BIN installed,
# remove included version.
BIN_PATH=$(command -v $BIN)
if [ -x "$BIN_PATH" ]; then
  echo "${BIN^} installed in: $BIN_PATH. Removing..."
  sudo rm -f "$BIN_PATH"
fi

# Check if $CACHE_PATH exists, and create if not
if [ ! -d "$CACHE_PATH" ]; then
  mkdir -p "$CACHE_PATH"
fi

# Download and extract requested versions
for VERSION in $VERSIONS; do
  URL_SHA="${URL_SHA//$URL_REPLACE/$VERSION}"
  EXPECTED_SHA=$(curl -s -qL "$URL_SHA" | awk -v pattern="${BIN}_${VERSION}${URL_BIN_SUFFIX}" '$0~pattern {print $1}')

  URL_BIN="${URL_BIN//$URL_REPLACE/$VERSION}"
  echo "Downloading: $URL_BIN"
  curl -s -qL -o "${BIN}.zip" "$URL_BIN"

  ACTUAL_SHA=$(sha256sum "${BIN}.zip" | awk '{print $1}')
  # Validate SHA256
  if [ "$EXPECTED_SHA" != "" ]; then
    if [ "$EXPECTED_SHA" != "$ACTUAL_SHA" ]; then
      echo "Error: SHA256 mismatch"
      echo "Expected: $EXPECTED_SHA | Actual: $ACTUAL_SHA"
      exit 1
    fi
  else
    echo "Error: EXPECTED_SHA not found"
    exit 1
  fi

  unzip -o "${BIN}.zip"
  if [ "$VERSION" == "$DEFAULT_VERSION" ]; then
    sudo cp "$BIN" "$INSTALL_PATH"
    sudo cp "$BIN" "$CACHE_PATH"
    echo "${BIN^} installed in: $(command -v $BIN)"
  fi
  sudo cp "$BIN" "${INSTALL_PATH}/${BIN}-${VERSION}"
  sudo mv "$BIN" "${CACHE_PATH}/${BIN}-${VERSION}"
done

# Validate the default version matches the installed version
INSTALLED_VERSION=$($BIN --version --json | jq -r .terraform_version)
if [ "$INSTALLED_VERSION" != "$DEFAULT_VERSION" ]; then
    echo "Error: Version Mismatch"
    echo "Requested: $DEFAULT_VERSION"
    echo "Installed: $INSTALLED_VERSION"
    exit 1
fi

# Cleanup
rm -f "$BIN"
rm -f "${BIN}.zip"