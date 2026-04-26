#!/bin/sh

# Usage: ./git-iter-diffs.sh <start-tag> <end-tag> [output-dir]

START="$1"
END="$2"
OUTDIR="${3:-diffs}"

if [ -z "$START" ] || [ -z "$END" ]; then
    echo "Usage: $0 <start-tag> <end-tag> [output-dir]"
    exit 1
fi

mkdir -p "$OUTDIR" || exit 1

# Get commit list (oldest → newest), excluding START itself
COMMITS=$(git rev-list --reverse "$START..$END") || exit 1

i=1
for COMMIT in $COMMITS; do
    PARENT=$(git rev-parse "$COMMIT^" 2>/dev/null)

    # Skip root commits (no parent)
    if [ -z "$PARENT" ]; then
        continue
    fi

    SHORT=$(git rev-parse --short "$COMMIT")
    FILE=$(printf "%s/%03d-%s.diff" "$OUTDIR" "$i" "$SHORT")

    echo "Generating $FILE"
    git diff "$PARENT" "$COMMIT" > "$FILE"

    i=$((i + 1))
done

echo "Done. Generated $((i - 1)) diff files in $OUTDIR"
