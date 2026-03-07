#!/bin/bash

cd "$(dirname "$0")/.." || exit 1

PAUSE_FILE="nephewbot/.paused"
LOCK_FILE="nephewbot/.lock"
LOG="nephewbot/nephewbot.log.jsonl"
FAIL_COUNT_FILE="nephewbot/.fail_count"

case "${1:-}" in
  pause)
    touch "$PAUSE_FILE"
    echo "nephewbot paused"
    exit 0
    ;;
  resume)
    rm -f "$PAUSE_FILE" "$FAIL_COUNT_FILE"
    echo "nephewbot resumed (failure count reset)"
    exit 0
    ;;
  status)
    if [ -f "$PAUSE_FILE" ]; then
      echo "nephewbot is paused"
    else
      echo "nephewbot is active"
    fi
    if [ -f "$FAIL_COUNT_FILE" ]; then
      echo "consecutive failures: $(cat "$FAIL_COUNT_FILE")"
    fi
    exit 0
    ;;
esac

if [ -f "$PAUSE_FILE" ]; then
  exit 0
fi

# Prevent overlapping runs
exec 9>"$LOCK_FILE"
if ! flock -n 9; then
  echo "$(date -Iseconds) nephewbot already running, skipping" >&2
  exit 0
fi

# Pull latest before starting
git pull --rebase origin main 2>/dev/null || true

N=${1:-3}

for i in $(seq 1 "$N"); do
  TIMESTAMP=$(date -Iseconds)

  OUTPUT=$(/usr/local/bin/claude -p "/neph" --dangerously-skip-permissions --output-format json 2>&1)
  EXIT_CODE=$?

  SESSION_ID=$(echo "$OUTPUT" | jq -r '.session_id // empty' 2>/dev/null)
  RESULT=$(echo "$OUTPUT" | jq -r '.result // empty' 2>/dev/null)

  # Get a meaningful summary — skip preamble lines like "Done! Here's a summary..."
  SUMMARY=$(echo "$RESULT" | grep -v -iE '^(done|all done|here|the plan is ready)' | grep -v '^$' | head -n 1 | cut -c 1-200)
  # Fallback: second line if the filter killed everything
  if [ -z "$SUMMARY" ]; then
    SUMMARY=$(echo "$RESULT" | sed -n '2p' | cut -c 1-200)
  fi

  if [ $EXIT_CODE -ne 0 ] || [ -z "$SESSION_ID" ]; then
    # Track consecutive failures
    FAILS=$(cat "$FAIL_COUNT_FILE" 2>/dev/null || echo 0)
    FAILS=$((FAILS + 1))
    echo "$FAILS" > "$FAIL_COUNT_FILE"

    jq -n -c \
      --arg ts "$TIMESTAMP" \
      --arg status "failed" \
      --argjson exit_code "$EXIT_CODE" \
      --arg error "$(echo "$OUTPUT" | tail -n 5 | cut -c 1-500)" \
      --argjson fail_streak "$FAILS" \
      '{timestamp: $ts, status: $status, exit_code: $exit_code, error: $error, fail_streak: $fail_streak}' >> "$LOG"

    # Auto-pause after 3 consecutive failures
    if [ "$FAILS" -ge 3 ]; then
      touch "$PAUSE_FILE"
      jq -n -c \
        --arg ts "$(date -Iseconds)" \
        --arg status "auto_paused" \
        --arg reason "$FAILS consecutive failures" \
        '{timestamp: $ts, status: $status, reason: $reason}' >> "$LOG"
      break
    fi
  else
    # Reset failure counter on success
    rm -f "$FAIL_COUNT_FILE"

    jq -n -c \
      --arg ts "$TIMESTAMP" \
      --arg status "ok" \
      --arg session "$SESSION_ID" \
      --arg summary "$SUMMARY" \
      '{timestamp: $ts, status: $status, session_id: $session, summary: $summary}' >> "$LOG"
  fi
done
