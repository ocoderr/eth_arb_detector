# ETH Arbitrage Detector

An offline extraction of the Ethereum arbitrage recognition core. Go standard
library only: no RPC client, daemon, keys, notifications, server scripts or
production configuration. The original analyzer is not modified to depend on
this project.

## Run

Requires Go 1.22 or newer.

```sh
go test ./...
go run ./cmd/arb-detect --input testdata/synthetic-cycle.json
go build -o bin/arb-detect ./cmd/arb-detect
```

Input is one JSON object containing `block`, `transaction`, and `receipt`,
using Ethereum JSON-RPC result field names (not JSON-RPC envelopes).
Use `--input -` for stdin. The CLI rejects missing objects, conflicting tx
hashes/block numbers, missing receipt status and inputs larger than 16 MiB.
It does not fetch or authenticate data. Obtain evidence independently.

The included fixture is synthetic: a three-transfer cycle with positive gross
WETH delta but fees greater than that delta. A positive candidate is deliberately
NOT a claim of positive net profit. Both candidate and noncandidate results exit
successfully; malformed input exits nonzero.

## Core API

```go
import detector "eth-arb-detector"

candidate := detector.IdentifyOnlineArb(block, tx, receipt)
isCandidate := candidate.IsCandidate(50)
transfers := detector.AnalyzeTransferTable(receipt, tx)
```

- transfer_table.go: transfer graph, strongly connected components and token-net
  heuristics for candidate cycles.
- online_arb.go: main recognition entrypoint (legacy name; no network activity).
- events.go, discover.go: receipt-event evidence and transaction summaries.
- cycle.go, economics.go: route-cycle and limited WETH/fee diagnostics.
- classify.go, protocols.go: optional static selector classification.
- tenderly.go: optional trace data types/conversion only; no provider calls.

## Interpretation Limits

This preserves the source heuristics, including their limitations. Confidence
is a rule score, not a calibrated probability. Receipts may be incomplete or
contain misleading logs; known topics/selectors alone do not authenticate pools.
The CLI validates only basic input consistency, not canonical chain inclusion.

Transfer cycles can miss cross-transaction, CEX or cross-chain arbitrage, and
can confuse other circular flows with arbitrage. WETH-specific economics do not
account for every asset, private payment, flash-loan cost or reverted trace
ancestor. Do not interpret named revenue/profit fields as verified net profit.

Selector support flags are an inherited static vocabulary, not a live inventory
of any deployed bot. Unknown calls are research leads, not proven new protocols.
Signal competition requires separate transaction ordering and signal evidence.
Ethereum mainnet WETH and protocol constants are retained; this is not a
chain-independent implementation.

## Extraction Provenance

Extracted from eth_arb_analyzer/internal/analyzer at commit
`6622a3b`. All 16 Go source/test files except contracts.go are retained with
only the package renamed from analyzer to detector. contracts.go, which holds
owner-specific contract/recipient addresses, is intentionally excluded.
CLI, fixture and CLI tests are new; recognition logic is unchanged.

This directory does not copy the original .git history or configuration.
No remote, scheduled job, license grant or publication has been configured.
