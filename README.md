# essingan-ern — the ERN and ERI library of ESSINGAN

**One library for every service, the CLI and the console** (ADR-0051 D12, register D70). It is a Go module and a TypeScript package, and both run **the same test vectors**.

- **Owners:** Gébril (Go, lot L3), Eric (TypeScript). **Reviewers:** Rykiel (IAM, first consumer), the CTO office.
- **Specification:** corpus Doc 73-01 (ERN v2) and ADR-0051 (D1 to D12).

## What it does

| Function | Rule |
|---|---|
| **Build and parse** | `ern:<partition>:<service>:<location>:<account>:<resource-path>` and the ten patterns E1–E10 |
| **Validate** | Partition `essingan`, `essingan-staging`, `essingan-dev` or `essingan-<id>` (D11). Service from the closed namespace list (Doc 73-01 §5). Location `<provider>-<region>`, the provider from the **open registry** (D7), given as data, never hard-coded. Account: 12 digits |
| **Match** (policies only) | `*` and `?` inside one component or segment. `**` for any number of whole segments, only as the last element of a resource path (D10) |
| **Create ERIs** | `<prefix>-` + the first 16 characters of a ULID, lowercased (10 of time, 6 of entropy). Time-sortable. A collision is caught by the owner's unique constraint and retried (D10) |

## Layout (to build)

```
go/            Go module  github.com/ESSINGAN/essingan-ern/go
ts/            TypeScript package  @essingan/ern
testdata/      vectors.json — the single source of the cases both languages run
```

## Rules

- **A case is added to `testdata/vectors.json` first,** then both implementations pass it in CI.
- **No service keeps its own copy after M1.** IAM's `facade/internal/ern` (essingan-iam #129) seeds this repository; Gébril reconciles it with his own implementation through the vectors.
- Every PR carries the integration sheet (corpus Doc 73-15).
