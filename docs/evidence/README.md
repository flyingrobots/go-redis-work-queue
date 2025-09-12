# Evidence for v0.4.0-alpha Promotion

- CI run: see `ci_run.json` (contains URL to the successful workflow run)
- Bench JSON: `bench.json` (admin bench with 1000 jobs at 500 rps)
- Config used: `config.alpha.yaml`
- Metrics snapshots: `metrics_before.txt`, `metrics_after.txt`

Reproduce locally

1) Ensure Redis is running on `localhost:6379` (e.g., `docker run -p 6379:6379 redis:7-alpine`)
2) Build binary: `make build`
3) Start worker: `./bin/job-queue-system --role=worker --config=docs/evidence/config.alpha.yaml`
4) In another terminal, run bench: `./bin/job-queue-system --role=admin --config=docs/evidence/config.alpha.yaml --admin-cmd=bench --bench-count=1000 --bench-rate=500 --bench-priority=low --bench-timeout=60s`
5) Capture metrics: `curl -sS localhost:9191/metrics | head -n 200 > docs/evidence/metrics_after.txt`

Notes
- The simple latency reported in `bench.json` is measured by comparing current time to each job's creation_time after completion sampling and is a coarse approximation. For precise latency distributions, prefer Prometheus histogram `job_processing_duration_seconds` and compute quantiles there.
