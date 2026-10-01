package detector

import (
	"fmt"
	"strings"
)

type OnlineArbCandidate struct {
	BlockNumber        int64                 `json:"blockNumber"`
	BlockTimestamp     int64                 `json:"blockTimestamp"`
	BlockMiner         string                `json:"blockMiner"`
	TxHash             string                `json:"txHash"`
	TxIndex            int64                 `json:"txIndex"`
	From               string                `json:"from"`
	To                 string                `json:"to"`
	Status             string                `json:"status"`
	InputSelector      string                `json:"inputSelector,omitempty"`
	ValueWei           string                `json:"valueWei,omitempty"`
	GasLimit           int64                 `json:"gasLimit,omitempty"`
	GasPriceWei        string                `json:"gasPriceWei,omitempty"`
	MaxFeePerGasWei    string                `json:"maxFeePerGasWei,omitempty"`
	MaxPriorityFeeWei  string                `json:"maxPriorityFeeWei,omitempty"`
	GasUsed            int64                 `json:"gasUsed"`
	EffectiveGasPrice  int64                 `json:"effectiveGasPrice"`
	Kind               string                `json:"kind"`
	Confidence         int                   `json:"confidence"`
	Score              int                   `json:"score"`
	Reasons            []string              `json:"reasons"`
	RouteProtocols     []string              `json:"routeProtocols,omitempty"`
	RoutePools         []string              `json:"routePools,omitempty"`
	HopCount           int                   `json:"hopCount,omitempty"`
	Logs               int                   `json:"logs"`
	SwapLikeLogs       int                   `json:"swapLikeLogs"`
	TransferLogs       int                   `json:"transferLogs"`
	UniqueTokenLogs    int                   `json:"uniqueTokenLogs"`
	UniqueLogContracts int                   `json:"uniqueLogContracts"`
	Economics          TxEconomics           `json:"economics"`
	Cycle              CycleInfo             `json:"cycle"`
	TransferAnalysis   TransferTableAnalysis `json:"transferAnalysis"`
	EventSummary       ReceiptEventSummary   `json:"eventSummary"`
}

func IdentifyOnlineArb(block Block, tx BlockTx, receipt Receipt) OnlineArbCandidate {
	discovery := DiscoverTransaction(tx, receipt)
	events := ParseReceiptEvents(receipt)
	economics := BuildTxEconomics(tx, receipt)
	transferAnalysis := AnalyzeTransferTable(receipt, tx)
	cycle := DetectCycle(events, economics)
	kind, confidence := classifyTransferTableArb(discovery, transferAnalysis)
	blockNumber := HexOrDecimalInt(block.Number)
	reasons := transferTableReasons(transferAnalysis)
	return OnlineArbCandidate{
		BlockNumber:        blockNumber,
		BlockTimestamp:     HexOrDecimalInt(block.Timestamp),
		BlockMiner:         strings.ToLower(block.Miner),
		TxHash:             discovery.TxHash,
		TxIndex:            discovery.TxIndex,
		From:               strings.ToLower(discovery.From),
		To:                 strings.ToLower(discovery.To),
		Status:             discovery.Status,
		InputSelector:      discovery.InputSelector,
		ValueWei:           discovery.ValueWei,
		GasLimit:           discovery.GasLimit,
		GasPriceWei:        discovery.GasPriceWei,
		MaxFeePerGasWei:    discovery.MaxFeePerGasWei,
		MaxPriorityFeeWei:  discovery.MaxPriorityFeeWei,
		GasUsed:            discovery.GasUsed,
		EffectiveGasPrice:  discovery.EffectiveGasPrice,
		Kind:               kind,
		Confidence:         confidence,
		Score:              confidence,
		Reasons:            reasons,
		RouteProtocols:     orderedProtocols(events.SwapPath),
		RoutePools:         orderedPools(events.SwapPath),
		HopCount:           len(events.SwapPath),
		Logs:               discovery.Logs,
		SwapLikeLogs:       discovery.SwapLikeLogs,
		TransferLogs:       discovery.TransferLogs,
		UniqueTokenLogs:    discovery.UniqueTokenLogs,
		UniqueLogContracts: discovery.UniqueLogContracts,
		Economics:          economics,
		Cycle:              cycle,
		TransferAnalysis:   transferAnalysis,
		EventSummary:       events,
	}
}

func orderedProtocols(events []ProtocolEvent) []string {
	out := make([]string, 0, len(events))
	for _, event := range events {
		if event.Protocol == "" {
			continue
		}
		out = append(out, event.Protocol)
	}
	return out
}

func orderedPools(events []ProtocolEvent) []string {
	out := make([]string, 0, len(events))
	for _, event := range events {
		if event.Pool == "" {
			continue
		}
		out = append(out, event.Pool)
	}
	return out
}

func (c OnlineArbCandidate) IsCandidate(minConfidence int) bool {
	return c.Kind != "other" && c.Confidence >= minConfidence
}

func classifyTransferTableArb(c DiscoveryCandidate, transferAnalysis TransferTableAnalysis) (string, int) {
	if c.Status != "" && c.Status != "0x1" {
		return "failed", 0
	}
	if transferAnalysis.ProfitedCycleCount == 0 {
		return "other", 0
	}
	confidence := 65
	confidence += transferAnalysis.ProfitedCycleCount * 5
	if transferAnalysis.TransferCount >= 6 {
		confidence += 10
	}
	if transferAnalysis.TokenCount >= 2 {
		confidence += 10
	}
	if transferAnalysis.SCCCount > 0 {
		confidence += 5
	}
	return "transfer_cycle_candidate", clampConfidence(confidence)
}

func transferTableReasons(transferAnalysis TransferTableAnalysis) []string {
	if transferAnalysis.ProfitedCycleCount > 0 {
		return []string{"transfer-table-profited-cycle"}
	}
	return nil
}

func clampConfidence(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func FormatOnlineArbSummary(c OnlineArbCandidate) string {
	return fmt.Sprintf("block=%d txIndex=%d kind=%s confidence=%d score=%d tx=%s swaps=%d transfers=%d tokens=%d contracts=%d",
		c.BlockNumber, c.TxIndex, c.Kind, c.Confidence, c.Score, c.TxHash,
		c.SwapLikeLogs, c.TransferLogs, c.UniqueTokenLogs, c.UniqueLogContracts)
}
