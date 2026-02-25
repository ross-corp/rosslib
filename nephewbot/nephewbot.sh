#!/bin/bash

cd "$(dirname "$0")/.." || exit 1

N=${1:-5}

echo "$(date '+%Y-%m-%d %H:%M:%S') — running worker skill $N times"
echo "---"

for i in $(seq 1 "$N"); do
  echo "iteration $i / $N"
  /usr/local/bin/claude -p "/worker" --dangerously-skip-permissions
  echo "--- completed $i / $N ---"
done

echo "$(date '+%Y-%m-%d %H:%M:%S') — done"
