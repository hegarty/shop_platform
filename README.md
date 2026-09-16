# shop_platform

Shared Go library for the commerce-intel platform's services
([`shop_ingestor`](https://github.com/hegarty/shop_ingestor),
[`shop_analytics`](https://github.com/hegarty/shop_analytics),
[`shop_notifier`](https://github.com/hegarty/shop_notifier)). No `cmd/` binaries, no
deployment — just a Go module the other three import.

See [`shop_docs`](https://github.com/hegarty/shop_docs) for the platform's architecture,
event model, and database model.

## What's here

| Package | Purpose |
|---|---|
| `event` | Canonical event envelopes — `RawEnvelope` (source payload wrapper) and `CommerceEvent` (normalized `commerce.*` events), plus `OrderAttributes` |
| `money` | Currency amounts as integer minor units (cents), not floats — avoids rounding error across sums/refunds/discounts |
| `period` | Timezone-aware reporting period math (daily/weekly/monthly/quarterly, today/yesterday/month-to-date/etc.) |
| `redpanda` | Producer/consumer wrapper over franz-go: tenant-keyed publish, trace-context propagation via headers |
| `db` | Postgres connection pooling + a minimal migration runner (each service owns its own `migrations/` directory and calls this) |
| `otelx` | One shared OpenTelemetry bootstrap (traces + metrics via OTLP/gRPC) |
| `logging` | Structured (JSON) logging with automatic trace/span ID correlation, plus header redaction helpers |
| `config` | Environment variable loading that fails fast, reporting every missing/invalid variable at once |
| `tenant` | The shared `Tenant` shape (deliberately minimal — see ADR-007 in `shop_docs`) |

## Using this from another shop_* service

```bash
go get github.com/hegarty/shop_platform@latest
```

## Development

```bash
make test    # go test ./...
make lint    # golangci-lint run
make fmt     # gofmt -l . (fails if anything is unformatted)
```

Requires Go 1.23.8 (see `.tool-versions`; `asdf install` picks it up automatically).

## Security

See [SECURITY.md](SECURITY.md). This repo has no LICENSE file — see that omission as
intentional, not an oversight: all rights are reserved by default.
