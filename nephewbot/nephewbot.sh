#!/bin/bash

cd "$(dirname "$0")/.." || exit 1

N=${1:-5}
LOG="nephewbot/nephewbot.log.jsonl"

for i in $(seq 1 "$N"); do
  TIMESTAMP=$(date -Iseconds)

  OUTPUT=$(/usr/local/bin/claude -p "/worker" --dangerously-skip-permissions --output-format json 2>&1)
  EXIT_CODE=$?

  SESSION_ID=$(echo "$OUTPUT" | jq -r '.session_id // empty' 2>/dev/null)
  RESULT=$(echo "$OUTPUT" | jq -r '.result // empty' 2>/dev/null)
  SUMMARY=$(echo "$RESULT" | head -n 1 | cut -c 1-200)

  if [ $EXIT_CODE -ne 0 ] || [ -z "$SESSION_ID" ]; then
    jq -n -c \
      --arg ts "$TIMESTAMP" \
      --arg status "failed" \
      --argjson exit_code "$EXIT_CODE" \
      --arg error "$(echo "$OUTPUT" | head -n 5)" \
      '{timestamp: $ts, status: $status, exit_code: $exit_code, error: $error}' >> "$LOG"
  else
    jq -n -c \
      --arg ts "$TIMESTAMP" \
      --arg status "ok" \
      --arg session "$SESSION_ID" \
      --arg summary "$SUMMARY" \
      '{timestamp: $ts, status: $status, session_id: $session, summary: $summary}' >> "$LOG"
  fi
done
