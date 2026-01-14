#!/bin/bash
# Progress checker with per-tier breakdown

echo "==================================================================="
echo "  Batch AI Progress Check"
echo "==================================================================="
echo ""

# Check if still running
if ps aux | grep -q "[b]atch_ai_impl"; then
    CURRENT_TIER=$(ps aux | grep "[b]atch_ai_impl" | grep -v grep | grep -oP 'tier \K[0-9]+' | head -1)
    echo "✅ Status: RUNNING (Currently on Tier $CURRENT_TIER)"
    echo ""
else
    echo "⏹️  Status: STOPPED"
    echo ""
fi

# Per-tier function counts (from all_functions_massive.json)
echo "📊 Progress by Tier:"
echo ""
echo "Tier 1 (Graphics & UI):      40 functions"
echo "Tier 2 (Paint & World):      15 functions"  
echo "Tier 3 (Game Logic):         17 functions"
echo "Tier 4 (Input/Audio/Config): 20 functions"
echo "Tier 5 (Multiplayer/Extra):   8 functions"
echo "                             ───────────────"
echo "Total:                      100 functions"
echo ""

# Count recently generated files by package (rough tier estimate)
echo "📁 Recently generated files by system:"
echo "   Graphics:   $(find pkg/graphics -name "*.go" -mmin -300 2>/dev/null | wc -l) files"
echo "   UI:         $(find pkg/ui -name "*.go" -mmin -300 2>/dev/null | wc -l) files"
echo "   Paint:      $(find pkg/paint -name "*.go" -mmin -300 2>/dev/null | wc -l) files"
echo "   World:      $(find pkg/world -name "*.go" -mmin -300 2>/dev/null | wc -l) files"
echo "   Vehicle:    $(find pkg/vehicle -name "*.go" -mmin -300 2>/dev/null | wc -l) files"
echo "   Company:    $(find pkg/company -name "*.go" -mmin -300 2>/dev/null | wc -l) files"
echo "   Economy:    $(find pkg/economy -name "*.go" -mmin -300 2>/dev/null | wc -l) files"
echo "   Other:      $(find pkg -name "*.go" -mmin -300 2>/dev/null | grep -v -E "(graphics|ui|paint|world|vehicle|company|economy)" | wc -l) files"
echo ""

# Time-based counts
NEW_FILES=$(find pkg -name "*.go" -mmin -60 2>/dev/null | wc -l)
echo "⏱️  Generation rate:"
echo "   Last hour:  $NEW_FILES files"
echo "   Last 3 hrs: $(find pkg -name "*.go" -mmin -180 2>/dev/null | wc -l) files"
echo ""

# Show most recent
echo "📄 Most recently generated (last 5):"
find pkg -name "*.go" -mmin -60 -printf "%T@ %p\n" 2>/dev/null | sort -rn | head -5 | cut -d' ' -f2- | sed 's/^/   /'
echo ""

# Estimate based on tier
if [ ! -z "$CURRENT_TIER" ]; then
    COMPLETED_TIERS=$((CURRENT_TIER - 1))
    if [ $CURRENT_TIER -eq 1 ]; then
        FUNC_DONE=0
        FUNC_TOTAL=40
    elif [ $CURRENT_TIER -eq 2 ]; then
        FUNC_DONE=40
        FUNC_TOTAL=55
    elif [ $CURRENT_TIER -eq 3 ]; then
        FUNC_DONE=55
        FUNC_TOTAL=72
    elif [ $CURRENT_TIER -eq 4 ]; then
        FUNC_DONE=72
        FUNC_TOTAL=92
    elif [ $CURRENT_TIER -eq 5 ]; then
        FUNC_DONE=92
        FUNC_TOTAL=100
    fi
    
    PERCENT=$((FUNC_DONE * 100 / 100))
    echo "🎯 Estimated overall progress: ~${PERCENT}% complete"
    echo "   (Tier $CURRENT_TIER in progress, Tiers 1-$COMPLETED_TIERS done)"
fi

echo ""
echo "💡 Tip: Some complex functions (paint, sprites) take longer!"
echo "💡 Run './check_progress.sh' anytime for updates"
