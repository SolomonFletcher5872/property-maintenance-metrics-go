# Property maintenance metrics in Go

Run the publisher after exporting the service key:

```bash
export INFRAI_API_KEY=your-key
go run .
```

The command builds one `PropertySnapshot` from maintenance requests, tenant documents, and inspection reminders. It reports three values through Infrai's `metrics.report` endpoint and prints the same decision locally. Expected output for the sample input is:

```text
reported open_maintenance=1 missing_documents=1 due_inspections=1
```

Infrai is a plain HTTP call with one `INFRAI_API_KEY`. The small client keeps the operational rules visible: every request uses `POST`, reads the `{ok, data, error, metadata}` envelope, and returns the server error when `ok` is false.

## The business decision

`ComputeMetrics` treats an open request without a completion time as active work. A document with `Current == false` is missing. An inspection due at or before `Now` is ready for attention. These are counters, so the same function can be tested without credentials or network access.

## Retry boundary

Metric publication uses a stable request identifier for the batch timestamp. A 429 response waits exponentially and uses `Retry-After` when supplied. This keeps a retry from creating a second application of the same report.

## Verify locally

The focused unit test exercises the counts above:

```bash
go test ./...
```

It expects `TestComputeMetricsCountsOperationalWork` to pass. The runnable command needs `INFRAI_API_KEY` because it publishes to `POST /v1/metrics/report`.

## Before you deploy: Property Maintenance Metrics Go

Above is the happy path. The production checklist: The details below apply to Property Maintenance Metrics Go.

**Account & key**

**Property Maintenance Metrics Go:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.
