#!/bin/bash

cd "$(dirname "$0")/.." || exit 1

PAUSE_FILE="nephewbot/.paused"
LOCK_FILE="nephewbot/.lock"
LOG="nephewbot/nephewbot.log.jsonl"

if [ -f "$PAUSE_FILE" ]; then
  exit 0
fi

# Share the same lock as nephewbot — only one can run at a time
exec 9>"$LOCK_FILE"
if ! flock -n 9; then
  echo "$(date -Iseconds) nephewbot is running, deferring busybody" >&2
  exit 0
fi

git pull --rebase origin main 2>/dev/null || true

TIMESTAMP=$(date -Iseconds)

OUTPUT=$(/usr/local/bin/claude -p "/busybody" --dangerously-skip-permissions --output-format json 2>&1)
EXIT_CODE=$?

SESSION_ID=$(echo "$OUTPUT" | jq -r '.session_id // empty' 2>/dev/null)
RESULT=$(echo "$OUTPUT" | jq -r '.result // empty' 2>/dev/null)

SUMMARY=$(echo "$RESULT" | grep -v -iE '^(done|all done|here|the plan is ready)' | grep -v '^$' | head -n 1 | cut -c 1-200)
if [ -z "$SUMMARY" ]; then
  SUMMARY=$(echo "$RESULT" | sed -n '2p' | cut -c 1-200)
fi

if [ $EXIT_CODE -ne 0 ] || [ -z "$SESSION_ID" ]; then
  jq -n -c \
    --arg ts "$TIMESTAMP" \
    --arg status "failed" \
    --arg skill "busybody" \
    --argjson exit_code "$EXIT_CODE" \
    --arg error "$(echo "$OUTPUT" | tail -n 5 | cut -c 1-500)" \
    '{timestamp: $ts, status: $status, skill: $skill, exit_code: $exit_code, error: $error}' >> "$LOG"
else
  jq -n -c \
    --arg ts "$TIMESTAMP" \
    --arg status "ok" \
    --arg skill "busybody" \
    --arg session "$SESSION_ID" \
    --arg summary "$SUMMARY" \
    '{timestamp: $ts, status: $status, skill: $skill, session_id: $session, summary: $summary}' >> "$LOG"
fi
