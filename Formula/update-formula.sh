#!/bin/bash

# Script to update the Homebrew formula with new version and checksums
# Automatically detects the latest version from git tags

set -e

FORMULA_FILE="vi-mongo.rb"

get_latest_version() {
    if command -v git >/dev/null 2>&1 && git rev-parse --git-dir >/dev/null 2>&1; then
        VERSION=$(git describe --tags --abbrev=0 2>/dev/null)
        if [ -n "$VERSION" ]; then
            echo "Detected version from git tags: $VERSION"
            return 0
        fi
    fi
    
    echo "Could not automatically detect version."
    echo "Please ensure you're in a git repository with tags."
    echo ""
    read -p "Enter version manually (e.g., v0.1.30): " VERSION
    
    if [ -z "$VERSION" ]; then
        echo "Error: Version is required"
        exit 1
    fi
}

validate_version() {
    if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo "Warning: Version '$VERSION' doesn't match expected format (vx.y.z)"
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
}

check_release_exists() {
    echo "Checking if release $VERSION exists..."
    HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "https://github.com/kopecmaciej/vi-mongo/releases/tag/$VERSION")
    
    if [ "$HTTP_STATUS" != "200" ]; then
        echo "Error: Release $VERSION not found on GitHub"
        echo "Make sure the release is published before updating the formula"
        exit 1
    fi
    echo "Release $VERSION found ✓"
}

if [ $# -eq 1 ]; then
    VERSION=$1
    echo "Using provided version: $VERSION"
else
    get_latest_version
fi

validate_version
check_release_exists

echo "Updating Homebrew formula for version $VERSION..."

if [ ! -f "$FORMULA_FILE" ]; then
    echo "Error: Formula file '$FORMULA_FILE' not found"
    exit 1
fi

cp "$FORMULA_FILE" "$FORMULA_FILE.backup"

sed -i.bak "s/version \"[^\"]*\"/version \"$VERSION\"/" "$FORMULA_FILE"

sed -i.bak "s|releases/download/[^/]*|releases/download/$VERSION|g" "$FORMULA_FILE"

echo "Updated version and URLs in $FORMULA_FILE"

echo "Calculating checksums..."

calculate_checksum() {
    local os_arch=$1
    local name=$2
    local max_attempts=3
    local attempt=1
    local base_url="https://github.com/kopecmaciej/vi-mongo/releases/download/$VERSION"
    
    while [ $attempt -le $max_attempts ]; do
        echo "Calculating $name checksum (attempt $attempt/$max_attempts)..."
        
        if checksum=$(curl -sL "$base_url/vi-mongo_$os_arch.tar.gz" | shasum -a 256 | cut -d' ' -f1); then
            if [ -n "$checksum" ] && [ ${#checksum} -eq 64 ]; then
                echo "$name: $checksum"
                return 0
            fi
        fi
        
        echo "Failed to calculate $name checksum, retrying..."
        attempt=$((attempt + 1))
        sleep 2
    done
    
    echo "Error: Failed to calculate $name checksum after $max_attempts attempts"
    return 1
}

calculate_checksum "Darwin_arm64" "macOS ARM64"
MACOS_ARM64_CHECKSUM=$checksum

calculate_checksum "Darwin_x86_64" "macOS x86_64"
MACOS_X86_64_CHECKSUM=$checksum

calculate_checksum "Linux_arm64" "Linux ARM64"
LINUX_ARM64_CHECKSUM=$checksum

calculate_checksum "Linux_x86_64" "Linux x86_64"
LINUX_X86_64_CHECKSUM=$checksum

echo "Updating checksums in formula..."

sed -i.bak "s/sha256 \"PLACEHOLDER_SHA256_ARM64\"/sha256 \"$MACOS_ARM64_CHECKSUM\"/" "$FORMULA_FILE"
sed -i.bak "s/sha256 \"PLACEHOLDER_SHA256_X86_64\"/sha256 \"$MACOS_X86_64_CHECKSUM\"/" "$FORMULA_FILE"

sed -i.bak "/on_linux/,/end/ s/sha256 \"PLACEHOLDER_SHA256_ARM64\"/sha256 \"$LINUX_ARM64_CHECKSUM\"/" "$FORMULA_FILE"
sed -i.bak "/on_linux/,/end/ s/sha256 \"PLACEHOLDER_SHA256_X86_64\"/sha256 \"$LINUX_X86_64_CHECKSUM\"/" "$FORMULA_FILE"

echo "Updated checksums in $FORMULA_FILE"

rm -f "$FORMULA_FILE.bak"
rm -f "$FORMULA_FILE.backup"

echo ""
echo "✅ Formula updated successfully!"
echo ""
echo "Changes made:"
echo "- Updated version to: $VERSION"
echo "- Updated download URLs"
echo "- Updated all platform checksums"
echo ""
echo "Next steps:"
echo "1. Review changes: git diff $FORMULA_FILE"
echo "2. Test the formula: brew install --build-from-source $FORMULA_FILE"
echo "3. Commit the changes: git add $FORMULA_FILE && git commit -m 'Update to $VERSION'"
