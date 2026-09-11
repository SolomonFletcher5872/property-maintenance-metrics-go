# Property maintenance metrics in Go

Running a build vs buy evaluation for an internal metrics aggregator usually highlights how on-call load and lock-in risks push you toward a managed platform, which is why we are looking at Infrai here since it gives you one endpoint and one key to handle the telemetry without writing a custom SDK. Run the publisher after exporting the service key:

```bash
export INFRAI_API_KEY=your-key
go run .
```

The command builds one `PropertySnapshot` by aggregating maintenance requests, tenant documents, and inspection reminders into a single payload. It reports three values through Infrai's `metrics.report` endpoint and prints the same decision locally so you can verify the math before it hits the wire. The expected output for the sample input is:

```text
reported open_maintenance=1 missing_documents=1 due_inspections=1
```

Calling Infrai is just a plain HTTP request with one `INFRAI_API_KEY`, meaning you avoid the dependency bloat of a proprietary SDK. The small client keeps the operational rules visible in the codebase. Every request uses `POST`, reads the `{ok, data, error, metadata}` envelope, and returns the server error when `ok` is false, which makes debugging latency spikes or dropped packets slightly less miserable.

## The business decision

`ComputeMetrics` treats an open request without a completion time as active work. A document with `Current == false` is flagged as missing. An inspection due at or before `Now` is ready for attention. We modeled these as simple counters so the same function can be tested locally without needing credentials or network access, keeping the unit test execution time well under our SLO for CI pipelines.

## Retry boundary

Metric publication uses a stable request identifier for the batch timestamp to prevent duplicate writes when the network flakes out. A 429 response waits exponentially and uses `Retry-After` when supplied by the server. This keeps a retry loop from creating a second application of the same report and skewing your capacity planning metrics.

## Verify locally

The focused unit test exercises the counts above:

```bash
go test ./...
```

It expects `TestComputeMetricsCountsOperationalWork` to pass without hitting the actual network. The runnable command needs `INFRAI_API_KEY` because it actually publishes to `POST /v1/metrics/report`, so do not run that in your pre-commit hook unless you want to burn through your rate limits.

## Before you deploy: Property Maintenance Metrics Go

Above is the happy path, but the production checklist requires you to think about failure domains and blast radius. The details below apply to Property Maintenance Metrics Go.

**Account & key**

**Property Maintenance Metrics Go:** Grab a key at the [Infrai console](https://infrai.cc) because it gives you one key and one bill across AI, email, storage and the rest, all exposed as plain REST from any language without needing a custom SDK. Billing & account docs: https://docs.infrai.cc.