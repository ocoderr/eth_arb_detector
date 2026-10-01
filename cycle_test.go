package detector

import "testing"

func TestDetectCycleClosedTokenPath(t *testing.T) {
	info := DetectCycle(ReceiptEventSummary{SwapPath: []ProtocolEvent{
		{TokenIn: "weth", TokenOut: "usdc"},
		{TokenIn: "usdc", TokenOut: "weth"},
	}}, TxEconomics{})

	if info.Status != "cycle" {
		t.Fatalf("status = %s", info.Status)
	}
	if info.Reason != "closed_token_path" {
		t.Fatalf("reason = %s", info.Reason)
	}
	if got := len(info.TokenPath); got != 3 {
		t.Fatalf("token path len = %d", got)
	}
}

func TestDetectCycleOpenTokenPath(t *testing.T) {
	info := DetectCycle(ReceiptEventSummary{SwapPath: []ProtocolEvent{
		{TokenIn: "weth", TokenOut: "usdc"},
		{TokenIn: "usdc", TokenOut: "dai"},
	}}, TxEconomics{})

	if info.Status != "non_cycle" {
		t.Fatalf("status = %s", info.Status)
	}
	if info.Reason != "open_token_path" {
		t.Fatalf("reason = %s", info.Reason)
	}
}

func TestDetectCycleUnknownWhenDirectionMissing(t *testing.T) {
	info := DetectCycle(ReceiptEventSummary{SwapPath: []ProtocolEvent{
		{Protocol: "UniV3", Pool: "0xpool"},
		{TokenIn: "usdc", TokenOut: "weth"},
	}}, TxEconomics{
		WethTransferInWei:  "100",
		WethTransferOutWei: "90",
	})

	if info.Status != "unknown" {
		t.Fatalf("status = %s", info.Status)
	}
	if info.Reason != "missing_token_direction_with_weth_roundtrip" {
		t.Fatalf("reason = %s", info.Reason)
	}
	if info.KnownHops != 1 || info.UnknownHops != 1 {
		t.Fatalf("known/unknown = %d/%d", info.KnownHops, info.UnknownHops)
	}
}
