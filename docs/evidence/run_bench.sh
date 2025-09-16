#!/usr/bin/env bash
set -euo pipefail

# Simple, reproducible evidence harness
# - Starts a worker with the provided config
# - Optionally purges test keys
# - Captures stats-keys and /metrics before/after
# - Runs an admin bench (enqueue + wait)
# - Writes outputs into a timestamped directory under docs/evidence

COUNT=${COUNT:-1000}
RATE=${RATE:-500}
PRIORITY=${PRIORITY:-low}
CONFIG=${CONFIG:-docs/evidence/config.alpha.yaml}
BIN=${BIN:-./bin/job-queue-system}
OUTDIR=${OUTDIR:-docs/evidence/run_$(date +%Y%m%d_%H%M%S)}
PURGE=${PURGE:-1}

mkdir -p "$OUTDIR"

if [[ ! -x "$BIN" ]] && [[ -f "$BIN" ]]; then
  chmod +x "$BIN" || true
fi

if [[ ! -f "$BIN" ]]; then
  echo "Building binary..."
  make build
fi

# Extract metrics port (fallback 9191)
PORT=$(awk '/metrics_port:/ {gsub(":" , " "); print $2; exit}' "$CONFIG" || true)
PORT=${PORT:-9191}

echo "Writing outputs to: $OUTDIR"

echo "Stats (keys) BEFORE..."
"$BIN" --role=admin --config="$CONFIG" --admin-cmd=stats-keys > "$OUTDIR/stats_before.json"

if [[ "$PURGE" == "1" ]]; then
  echo "Purging test keys..."
  "$BIN" --role=admin --config="$CONFIG" --admin-cmd=purge-all --yes >/dev/null
fi

echo "Starting worker (background)..."
"$BIN" --role=worker --config="$CONFIG" > "$OUTDIR/worker.log" 2>&1 &
WPID=$!
echo $WPID > "$OUTDIR/worker.pid"

echo "Waiting for readiness on port $PORT..."
for i in {1..100}; do
  if curl -fsS "http://localhost:$PORT/readyz" >/dev/null; then
    break
  fi
  sleep 0.1
done

echo "Metrics BEFORE..."
curl -fsS "http://localhost:$PORT/metrics" | head -n 200 > "$OUTDIR/metrics_before.txt" || true

echo "Running bench: count=$COUNT rate=$RATE priority=$PRIORITY..."
"$BIN" --role=admin --config="$CONFIG" --admin-cmd=bench \
  --bench-count="$COUNT" --bench-rate="$RATE" --bench-priority="$PRIORITY" --bench-timeout=60s \
  | tee "$OUTDIR/bench.json"

echo "Metrics AFTER..."
curl -fsS "http://localhost:$PORT/metrics" | head -n 200 > "$OUTDIR/metrics_after.txt" || true

echo "Stats (keys) AFTER..."
"$BIN" --role=admin --config="$CONFIG" --admin-cmd=stats-keys > "$OUTDIR/stats_after.json"

echo "Stopping worker..."
kill "$WPID" || true
sleep 0.2

echo "Done. Outputs in: $OUTDIR"

