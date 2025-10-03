#!/bin/sh

set -e

if [ $# -ne 1 ]; then
    echo "<rc_file>?"
    exit 1
fi

RC_FILE="$1"

get_git_tag() {
    if command -v git >/dev/null 2>&1 && git rev-parse --git-dir >/dev/null 2>&1; then
        tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "0.0.0")
        echo "$tag" | sed 's/^v//'
    else
        echo "0.0.0"
    fi
}

get_git_commit_count() {
    if command -v git >/dev/null 2>&1 && git rev-parse --git-dir >/dev/null 2>&1; then
        git rev-list --count HEAD 2>/dev/null || echo "0"
    else
        echo "0"
    fi
}

get_git_commit_hash() {
    if command -v git >/dev/null 2>&1 && git rev-parse --git-dir >/dev/null 2>&1; then
        git rev-parse --short=7 HEAD 2>/dev/null || echo "0000000"
    else
        echo "0000000"
    fi
}

TAG=$(get_git_tag)
MAJOR=$(echo "$TAG" | cut -d. -f1)
MINOR=$(echo "$TAG" | cut -d. -f2)
PATCH=$(echo "$TAG" | cut -d. -f3)
COMMIT_COUNT=$(get_git_commit_count)
COMMIT_HASH=$(get_git_commit_hash)

cat > "$RC_FILE" << EOF
1 VERSIONINFO
FILEVERSION $MAJOR,$MINOR,$PATCH,$COMMIT_COUNT
PRODUCTVERSION $MAJOR,$MINOR,$PATCH,$COMMIT_COUNT
FILEOS 0x40004
FILETYPE 0x1
{
BLOCK "StringFileInfo"
{
	BLOCK "080404B0"
	{
		VALUE "FileDescription", "适用于 Windows 的简洁极域电子教室反控软件, 使用 Go 编写."
		VALUE "CompanyName", "dotcubecn"
		VALUE "LegalCopyright", "Copyright \xA9 2025 dotcubecn. Licensed under GPL-3.0-only."
		VALUE "ProductName", "Mythgone"
		VALUE "ProductVersion", "$MAJOR.$MINOR.$PATCH.g$COMMIT_HASH"
	}
}

BLOCK "VarFileInfo"
{
	VALUE "Translation", 0x0804 0x04B0  
}
}
EOF