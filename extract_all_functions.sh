#!/bin/bash
# Extract ALL functions from OpenLoco C++ codebase
# Generate comprehensive function database for batch AI implementation

OPENLOCO_SRC="/home/joop/goloco-project/OpenLoco/src/OpenLoco/src"
OUTPUT="all_functions.json"

echo "Scanning OpenLoco codebase for ALL functions..."
echo "This will take a few minutes..."

# Use grep to find all function definitions
# Look for patterns like: type functionName(...) or Class::method(...)
cd "$OPENLOCO_SRC"

# Start JSON
echo '{' > "$OUTPUT"
echo '  "functions": [' >> "$OUTPUT"

FUNC_ID=1
FIRST=1

# Find all .cpp files and extract function signatures
find . -name "*.cpp" | while read cppfile; do
    echo "  Processing: $cppfile"
    
    # Extract functions using more sophisticated patterns
    # Look for lines that look like function definitions
    grep -n "^[a-zA-Z_].*(.*)$" "$cppfile" | grep -v "^[[:space:]]*//" | head -50 | while IFS=: read linenum line; do
        # Skip preprocessor directives and comments
        if [[ "$line" =~ ^# ]] || [[ "$line" =~ ^// ]]; then
            continue
        fi
        
        # Try to extract function name
        FUNC_NAME=$(echo "$line" | sed -E 's/.*[ \*]([a-zA-Z_][a-zA-Z0-9_:]*)\s*\(.*/\1/' | head -1)
        
        if [ ! -z "$FUNC_NAME" ] && [ "$FUNC_NAME" != "$line" ]; then
            # Determine Go package based on file path
            GO_PACKAGE=$(dirname "$cppfile" | sed 's/^\.\///' | tr '/' '_' | tr '[:upper:]' '[:lower:]' | sed 's/^src_//')
            [ -z "$GO_PACKAGE" ] && GO_PACKAGE="core"
            
            # Determine Go file name
            GO_FILE=$(basename "$cppfile" .cpp | tr '[:upper:]' '[:lower:]').go
            
            # Add comma if not first
            if [ $FIRST -eq 0 ]; then
                echo "," >> "$OUTPUT"
            fi
            FIRST=0
            
            # Generate JSON entry
            cat >> "$OUTPUT" << EOFJSON
    {
      "id": "func_${FUNC_ID}",
      "priority": ${FUNC_ID},
      "tier": $((($FUNC_ID / 50) + 1)),
      "group": "$GO_PACKAGE",
      "cppFile": "${cppfile#./}",
      "function": "$FUNC_NAME",
      "goSignature": "func ${FUNC_NAME}() error",
      "goPackage": "$GO_PACKAGE",
      "goFile": "$GO_FILE",
      "dependencies": [],
      "complexity": "medium",
      "estimatedTokens": 500,
      "description": "Auto-extracted from $cppfile:$linenum"
    }
EOFJSON
            
            FUNC_ID=$((FUNC_ID + 1))
            
            # Limit to prevent huge file during testing
            if [ $FUNC_ID -gt 500 ]; then
                break 2
            fi
        fi
    done
done

# Close JSON
echo '' >> "$OUTPUT"
echo '  ]' >> "$OUTPUT"
echo '}' >> "$OUTPUT"

echo "Extracted $FUNC_ID functions to $OUTPUT"
