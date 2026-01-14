#!/bin/bash
# Clean up extracted type files to remove C++ remnants and fix syntax

set -e

LOCO_DIR="pkg/loco"

echo "🧹 Cleaning up extracted types..."

cleanup_file() {
    local file="$1"

    if [ ! -f "$file" ]; then
        return
    fi

    echo "  Cleaning: $file"

    # Use sed to clean up common issues
    sed -i \
        -e '/^\/\/ char /d' \
        -e '/^\/\/ std::/d' \
        -e '/^\/\/ return /d' \
        -e '/^\/\/ func /d' \
        -e '/^\/\/ \/\//d' \
        -e '/^\/\/ See below/d' \
        -e '/^\/\/ If object/d' \
        -e '/^\/\/ when looking/d' \
        -e '/^\/\/ This means/d' \
        -e '/^\/\/ Most custom/d' \
        -e '/^\/\/ Use the isVanilla/d' \
        -e '/^\/\/ a lookup/d' \
        -e 's/= ObjectHeader{[^}]*}/= ObjectHeader{}/g' \
        -e 's/std\.numeric_limits.*max()/0xFFFFFFFF/g' \
        -e 's/LoadedObjectId/uint32/g' \
        -e 's/StringId/uint16/g' \
        -e "s/0x[0-9A-Fa-f]*U/0x\1/g" \
        -e "s/'\\\\xFF'/0xFF/g" \
        "$file"

    # Remove multi-line C++ comments
    awk '
        BEGIN { in_cpp = 0 }
        /^\/\/ \/\/.*{/ { in_cpp = 1; next }
        in_cpp && /^}/ { in_cpp = 0; next }
        in_cpp { next }
        { print }
    ' "$file" > "$file.tmp"
    mv "$file.tmp" "$file"

    # Fix invalid const declarations
    sed -i \
        -e '/^const.*ObjectHeader{.*{/d' \
        -e '/^const.*std::/d' \
        -e '/^const.*0x.*{.*0xFF.*}/d' \
        "$file"
}

# Clean all extracted files
find "$LOCO_DIR" -name "*.go" | while read file; do
    cleanup_file "$file"
done

# Add missing type aliases
echo "📝 Adding type aliases..."

cat >> "$LOCO_DIR/objects/types.go" << 'EOF'

// Type aliases for consistency
type LoadedObjectId = uint32
const NullObjectId LoadedObjectId = 0xFFFFFFFF
EOF

cat >> "$LOCO_DIR/objects/cargo.go" << 'EOF'

// Type alias for string IDs
type StringId = uint16
EOF

echo ""
echo "✅ Cleanup complete!"
echo "Testing compilation..."
