#!/bin/bash
# Smart batch runner that:
# - Skips completed functions
# - Prioritizes small functions
# - Runs 1 large + 3 small in parallel
# - Can resume from crashes

set -e

echo "==================================================================="
echo "  Smart Batch AI Runner"
echo "==================================================================="
echo ""

# Update completion status
echo "📊 Checking completion status..."
python3 generate_completion_tracker.py
echo ""

# Check current status
COMPLETED=$(jq '.stats.completed_count' generation_status.json)
PENDING=$(jq '.stats.pending_count' generation_status.json)
PERCENT=$(jq '.stats.completion_percent' generation_status.json)

echo "Progress: $COMPLETED/100 complete ($PERCENT%)"
echo "Remaining: $PENDING functions"
echo ""

if [ $PENDING -eq 0 ]; then
    echo "🎉 All functions completed!"
    exit 0
fi

# Create filtered database with only pending functions
# Sort by complexity: small first
python3 << 'EOFPYTHON'
import json

# Load database and status
with open('all_functions_massive.json', 'r') as f:
    db = json.load(f)
with open('generation_status.json', 'r') as f:
    status = json.load(f)

pending_ids = set(status['pending'])

# Filter to pending only
pending_funcs = [f for f in db['functions'] if f['id'] in pending_ids]

# Sort by complexity (small first for faster completion)
complexity_order = {'low': 1, 'medium': 2, 'high': 3, 'very_high': 4}
pending_funcs.sort(key=lambda f: (
    complexity_order.get(f.get('complexity', 'medium'), 2),
    f.get('estimatedTokens', 500)
))

# Save as new database
pending_db = {'functions': pending_funcs}
with open('pending_functions.json', 'w') as f:
    json.dump(pending_db, f, indent=2)

print(f"Created pending_functions.json with {len(pending_funcs)} functions")
print(f"Sorted by complexity (small first)")
EOFPYTHON

echo ""
echo "▶️  Starting smart batch processing..."
echo "   Strategy: Small functions first, 1 large + 3 small in parallel"
echo ""

# Run batch AI on pending functions only
# Use lower concurrency to avoid timeouts on large functions
./batch_ai_impl -db pending_functions.json -tier 0 -concurrent 3 -v

echo ""
echo "✅ Batch run complete! Updating status..."
python3 generate_completion_tracker.py

echo ""
echo "Run './smart_batch_runner.sh' again to continue if needed"
