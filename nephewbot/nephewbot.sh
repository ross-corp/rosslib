#!/bin/bash

cd "$(dirname "$0")/.." || exit 1

PAUSE_FILE="nephewbot/.paused"
LOG="nephewbot/nephewbot.log.jsonl"

case "${1:-}" in
  pause)
    touch "$PAUSE_FILE"
    echo "nephewbot paused"
    exit 0
    ;;
  resume)
    rm -f "$PAUSE_FILE"
    echo "nephewbot resumed"
    exit 0
    ;;
  status)
    if [ -f "$PAUSE_FILE" ]; then
      echo "nephewbot is paused"
    else
      echo "nephewbot is active"
    fi
    exit 0
    ;;
esac

if [ -f "$PAUSE_FILE" ]; then
  exit 0
fi

N=${1:-5}

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
