## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 20 to 26, the documented test

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061522

- [review_comment] 2025-09-18T16:01:19Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:26)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 20 to 26, the documented test
names include a non-existent TestWebhookDeliveryWithRetries; update the block to
list the actual test names: keep TestHMACSigner_SignPayload and
TestHMACSigner_VerifySignature as-is, and replace the TestBackoffScheduler_*
entry with the concrete names TestBackoffScheduler_ExponentialStrategy,
TestBackoffScheduler_LinearStrategy, and TestBackoffScheduler_FixedStrategy;
remove or replace TestWebhookDeliveryWithRetries with
TestWebhookHarness_RetryOnFailure (test/integration/webhook_harness_test.go:405)
or use a generic TestWebhookHarness_* selector listing the real examples
(BasicDelivery, RetryOnFailure, ConcurrentDeliveries, SignatureValidation) so
the documentation matches existing test names and file locations.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD
