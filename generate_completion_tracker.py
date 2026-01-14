#!/usr/bin/env python3
"""Track which functions have been completed"""
import json
import os
import glob

# Load function database
with open('all_functions_massive.json', 'r') as f:
    db = json.load(f)

# Check which functions have generated files
completed = []
failed = []
pending = []

for func in db['functions']:
    go_file = f"pkg/{func['goPackage']}/{func['goFile']}"
    
    if os.path.exists(go_file):
        # Check if file contains the function
        with open(go_file, 'r', encoding='utf-8', errors='ignore') as f:
            content = f.read()
            # Extract function name from signature
            sig = func['goSignature']
            if 'func ' in sig:
                func_name = sig.split('(')[0].split()[-1]
                if func_name in content and len(content) > 500:  # Has substantial content
                    completed.append(func['id'])
                else:
                    pending.append(func['id'])
            else:
                pending.append(func['id'])
    else:
        pending.append(func['id'])

# Generate status report
status = {
    'completed': completed,
    'pending': pending,
    'failed': failed,
    'stats': {
        'completed_count': len(completed),
        'pending_count': len(pending),
        'failed_count': len(failed),
        'total': len(db['functions']),
        'completion_percent': round(len(completed) * 100 / len(db['functions']), 1)
    }
}

# Save status
with open('generation_status.json', 'w') as f:
    json.dump(status, f, indent=2)

# Print summary
print(f"✅ Completed: {len(completed)}/{len(db['functions'])} ({status['stats']['completion_percent']}%)")
print(f"⏳ Pending: {len(pending)}")
print(f"❌ Failed: {len(failed)}")
print(f"\nStatus saved to: generation_status.json")
