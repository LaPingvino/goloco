#!/bin/bash
# Extract clean type definitions from converted code into pkg/loco

set -e

CONVERTED_DIR="pkg/_conversion_generated"
OUTPUT_DIR="pkg/loco"

echo "🔧 Extracting clean types from converted code..."

# Create output directory structure
mkdir -p "$OUTPUT_DIR"/{objects,worldmap,vehicles,economy,s5,ui}

# Function to extract and clean types from a file
extract_types() {
    local input_file="$1"
    local output_file="$2"
    local package_name="$3"

    echo "  Processing: $(basename "$input_file")"

    # Start with package declaration
    echo "package $package_name" > "$output_file"
    echo "" >> "$output_file"
    echo "// Extracted from: $input_file" >> "$output_file"
    echo "// This file contains clean type definitions, enums, and constants" >> "$output_file"
    echo "" >> "$output_file"

    # Extract type definitions, enums, and constants
    # Skip C++ comments, function bodies, and malformed code
    awk '
        BEGIN {
            in_type = 0
            in_const = 0
            in_comment_block = 0
            type_buffer = ""
        }

        # Skip AUTO-GENERATED headers
        /AUTO-GENERATED|WILL NOT COMPILE/ { next }

        # Skip C++ includes and namespace comments
        /^\/\/ #include|^\/\/ namespace|^\/\/ forward:/ { next }

        # Skip method and function comments
        /^\/\/ method:|^\/\/ func |^\/\/ orphan member:/ { next }

        # Track const blocks
        /^const \(/ {
            in_const = 1
            print $0
            next
        }

        # End of const block
        in_const && /^\)/ {
            in_const = 0
            print $0
            next
        }

        # Print const entries
        in_const {
            print $0
            next
        }

        # Standalone const declarations
        /^const [A-Z]/ {
            # Only print if it looks valid (not C++ syntax)
            if ($0 !~ /std::/ && $0 !~ /any \/\*/) {
                print $0
            }
            next
        }

        # Type definitions
        /^type [A-Z]/ {
            in_type = 1
            type_buffer = $0
            # If its a simple type alias, print immediately
            if ($0 ~ /^type [A-Za-z_]+ (int|uint8|uint16|uint32|uint64|bool|string)$/) {
                print $0
                print ""
                in_type = 0
                type_buffer = ""
            }
            next
        }

        # Inside type definition
        in_type {
            type_buffer = type_buffer "\n" $0

            # Check for end of struct
            if ($0 ~ /^}/) {
                # Only print if it has at least one field
                if (type_buffer ~ /\n\t[A-Z]/) {
                    print type_buffer
                    print ""
                }
                in_type = 0
                type_buffer = ""
            }
            next
        }

        # Var declarations with array initializers
        /^var [A-Z].*= \[/ {
            # Skip if it has commented values
            if ($0 !~ /\/\//) {
                print $0
            }
            next
        }

    ' "$input_file" >> "$output_file"

    # Clean up empty lines (max 2 consecutive)
    awk 'NF || ++n < 2 { print; n = 0 }' "$output_file" > "$output_file.tmp"
    mv "$output_file.tmp" "$output_file"
}

# Extract objects
echo "📦 Extracting objects..."
extract_types "$CONVERTED_DIR/objects/object.go" "$OUTPUT_DIR/objects/types.go" "objects"
extract_types "$CONVERTED_DIR/objects/cargo_object.go" "$OUTPUT_DIR/objects/cargo.go" "objects"
extract_types "$CONVERTED_DIR/objects/train_station_object.go" "$OUTPUT_DIR/objects/station.go" "objects"
extract_types "$CONVERTED_DIR/objects/road_object.go" "$OUTPUT_DIR/objects/road.go" "objects"
extract_types "$CONVERTED_DIR/objects/vehicle_object.go" "$OUTPUT_DIR/objects/vehicle.go" "objects" 2>/dev/null || true

# Extract worldmap types
echo "🗺️  Extracting worldmap types..."
extract_types "$CONVERTED_DIR/worldmap/types.go" "$OUTPUT_DIR/worldmap/types.go" "worldmap"
extract_types "$CONVERTED_DIR/worldmap/tile.go" "$OUTPUT_DIR/worldmap/tile.go" "worldmap"

# Extract vehicle types
echo "🚂 Extracting vehicle types..."
extract_types "$CONVERTED_DIR/vehicles/vehicle.go" "$OUTPUT_DIR/vehicles/types.go" "vehicles"

# Extract S5 format
echo "💾 Extracting S5 save format..."
extract_types "$CONVERTED_DIR/s5/s5.go" "$OUTPUT_DIR/s5/format.go" "s5"

# Extract economy types
echo "💰 Extracting economy types..."
extract_types "$CONVERTED_DIR/economy/currency.go" "$OUTPUT_DIR/economy/currency.go" "economy" 2>/dev/null || true
extract_types "$CONVERTED_DIR/economy/expenditures.go" "$OUTPUT_DIR/economy/expenditures.go" "economy" 2>/dev/null || true

echo ""
echo "✅ Type extraction complete!"
echo "📁 Clean types are in: $OUTPUT_DIR"
echo ""
echo "Next steps:"
echo "  1. Run: go build ./pkg/loco/..."
echo "  2. Fix any compilation errors"
echo "  3. Generate function stubs: go run ./cmd/scaffold_generator"
