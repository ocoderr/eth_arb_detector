package detector

import "testing"

func TestDiscoverTransactionFlagsUnknownStyleCandidate(t *testing.T) {
	receipt := Receipt{
		Status: "0x1",
		Logs: []ReceiptLog{
			transferLog("0xtoken1"),
			transferLog("0xtoken2"),
			transferLog("0xtoken1"),
			transferLog("0xtoken3"),
			transferLog("0xtoken2"),
			transferLog("0xtoken4"),
		},
	}
	candidate := DiscoverTransaction(BlockTx{Hash: "0xabc", TransactionIndex: "0x2"}, receipt)

	if !candidate.IsCandidate() {
		t.Fatal("expected transfer fanout to be a discovery candidate")
	}
	if candidate.SwapLikeLogs != 0 {
		t.Fatalf("swap-like logs = %d, want 0", candidate.SwapLikeLogs)
	}
	if got := FormatReasons(candidate.Reasons); got != "transfer-fanout-possible-unknown-arb" {
		t.Fatalf("reasons = %q", got)
	}
}

func TestDiscoverTransactionFlagsMultiSwapCandidate(t *testing.T) {
	receipt := Receipt{
		Status: "0x1",
		Logs: []ReceiptLog{
			swapLog("0xpool1", "0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822"),
			swapLog("0xpool2", "0xc42079f94a6350d7e6235f29174924f928cc2ac818eb64fed8004e115fbcca67"),
			transferLog("0xtoken1"),
			transferLog("0xtoken2"),
			transferLog("0xtoken1"),
			transferLog("0xtoken2"),
		},
	}
	candidate := DiscoverTransaction(BlockTx{Hash: "0xdef", TransactionIndex: "3"}, receipt)

	if !candidate.IsCandidate() {
		t.Fatal("expected multi-swap transaction to be a discovery candidate")
	}
	if candidate.SwapLikeLogs != 2 {
		t.Fatalf("swap-like logs = %d, want 2", candidate.SwapLikeLogs)
	}
	if candidate.Score <= 0 {
		t.Fatalf("score = %d, want positive", candidate.Score)
	}
}

func transferLog(address string) ReceiptLog {
	return ReceiptLog{Address: address, Topics: []string{ERC20TransferTopic}}
}

func swapLog(address string, topic string) ReceiptLog {
	return ReceiptLog{Address: address, Topics: []string{topic}}
}
