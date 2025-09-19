# Harden Ack Path with Durable Ledger

## Summary
Acknowledgement path only updates Redis lists; historical records can be lost. Emit events to a durable sink (S3/Kafka/etc.) and add fallbacks/alerts when the sink is unavailable.

## Acceptance Criteria
- Ack flow publishes completion/failure events to a durable store with defined retention.
- Local NDJSON (if retained) rotates safely (size caps, gzip, PII scrubbing).
- Monitoring alerts when ledger writes fail; workers fail fast instead of silently dropping history.

## Dependencies / Inputs
- Worker ack/cleanup logic.
- Available durable sink client.

## Deliverables / Outputs
- Updated ack implementation with durable write path and error handling.
- Tests or integration checks demonstrating ledger persistence and alerting.
