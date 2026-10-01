package detector

import "testing"

func TestIdentifyOnlineArbIgnoresOldPrivateFlowHeuristic(t *testing.T) {
	block := Block{
		Number:    "0x17f852c",
		Timestamp: "0x6a590b3f",
		Miner:     "0x4838B106FCE9647bDF1e7877bF73cE8B0bad5f97",
	}
	tx := BlockTx{
		Hash:                 "0xprivate",
		TransactionIndex:     "0x4",
		From:                 "0xfrom",
		To:                   "0xto",
		Input:                "0x00000002",
		Value:                "0x17f852c",
		Gas:                  "0x7a120",
		MaxPriorityFeePerGas: "0x0",
	}
	receipt := Receipt{
		Status:            "0x1",
		GasUsed:           "0x2cf7f",
		EffectiveGasPrice: "0x61db49a",
		Logs: []ReceiptLog{
			transferLog("0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"),
			swapLog("0xcd4abaf39691dcbb877165b7ba6d04ec6436c22a", "0xc42079f94a6350d7e6235f29174924f928cc2ac818eb64fed8004e115fbcca67"),
		},
	}

	candidate := IdentifyOnlineArb(block, tx, receipt)
	if candidate.Kind != "other" {
		t.Fatalf("kind = %q, want other", candidate.Kind)
	}
	if candidate.InputSelector != "0x00000002" {
		t.Fatalf("selector = %q", candidate.InputSelector)
	}
	if candidate.ValueWei != "25134380" {
		t.Fatalf("value = %q", candidate.ValueWei)
	}
	if got := FormatReasons(candidate.Reasons); got != "" {
		t.Fatalf("reasons = %q", got)
	}
	if candidate.HopCount != 1 || len(candidate.RoutePools) != 1 {
		t.Fatalf("route = hops %d pools %v", candidate.HopCount, candidate.RoutePools)
	}
	if len(candidate.EventSummary.Events) != 0 {
		t.Fatalf("events should omit transfers and duplicate swaps, got %d", len(candidate.EventSummary.Events))
	}
	if len(candidate.EventSummary.TransferTokens) != 1 {
		t.Fatalf("transfer tokens = %v", candidate.EventSummary.TransferTokens)
	}
}

func TestParseReceiptEventsKeepsOrderedSwapPath(t *testing.T) {
	receipt := Receipt{
		Logs: []ReceiptLog{
			transferLog("0xtoken"),
			swapLog("0xpool1", "0xc42079f94a6350d7e6235f29174924f928cc2ac818eb64fed8004e115fbcca67"),
			swapLog("0xpool2", "0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822"),
		},
	}

	summary := ParseReceiptEvents(receipt)
	if summary.SwapEvents != 2 {
		t.Fatalf("swap events = %d", summary.SwapEvents)
	}
	if got := []string{summary.SwapPath[0].Pool, summary.SwapPath[1].Pool}; got[0] != "0xpool1" || got[1] != "0xpool2" {
		t.Fatalf("swap path = %v", got)
	}
	if len(summary.Events) != 0 {
		t.Fatalf("events should not duplicate swaps, got %d", len(summary.Events))
	}
	if len(summary.TransferTokens) != 1 || summary.TransferTokens[0] != "0xtoken" {
		t.Fatalf("transfer tokens = %v", summary.TransferTokens)
	}
}

func TestParseReceiptEventsInfersPoolTokenDirectionFromTransfers(t *testing.T) {
	receipt := Receipt{
		Logs: []ReceiptLog{
			erc20TransferLog("0xtokena0000000000000000000000000000000000", "0x00000000000000000000000000000000000000aa", "0x00000000000000000000000000000000000000bb", "0x1"),
			erc20TransferLog("0xtokenb0000000000000000000000000000000000", "0x00000000000000000000000000000000000000bb", "0x00000000000000000000000000000000000000cc", "0x1"),
			swapLog("0x00000000000000000000000000000000000000bb", "0xc42079f94a6350d7e6235f29174924f928cc2ac818eb64fed8004e115fbcca67"),
		},
	}

	summary := ParseReceiptEvents(receipt)
	if len(summary.SwapPath) != 1 {
		t.Fatalf("swap path len = %d", len(summary.SwapPath))
	}
	if summary.SwapPath[0].TokenIn != "0xtokena0000000000000000000000000000000000" {
		t.Fatalf("token in = %s", summary.SwapPath[0].TokenIn)
	}
	if summary.SwapPath[0].TokenOut != "0xtokenb0000000000000000000000000000000000" {
		t.Fatalf("token out = %s", summary.SwapPath[0].TokenOut)
	}
}
