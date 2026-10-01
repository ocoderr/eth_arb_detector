package detector

import "testing"

func TestAnalyzeTransferTableFindsProfitedSCC(t *testing.T) {
	bot := "0x00000000000000000000000000000000000000aa"
	pool1 := "0x00000000000000000000000000000000000000bb"
	pool2 := "0x00000000000000000000000000000000000000cc"
	weth := "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"
	tokenX := "0x00000000000000000000000000000000000000dd"
	receipt := Receipt{Logs: []ReceiptLog{
		erc20TransferLog(weth, bot, pool1, "0x64"),
		erc20TransferLog(tokenX, pool1, pool2, "0xc8"),
		erc20TransferLog(weth, pool2, bot, "0x6e"),
	}}

	analysis := AnalyzeTransferTable(receipt, BlockTx{From: "0x00000000000000000000000000000000000000ee"})
	if analysis.Status != "profited_cycle" {
		t.Fatalf("status = %s, want profited_cycle: %+v", analysis.Status, analysis)
	}
	if analysis.ProfitedCycleCount != 1 {
		t.Fatalf("profited cycles = %d", analysis.ProfitedCycleCount)
	}
	cycle := analysis.ProfitedCycles[0]
	if cycle.TradeAddressCount != 2 || cycle.ProfitAddressCount != 1 {
		t.Fatalf("cycle counts = %+v", cycle)
	}
}

func TestIdentifyOnlineArbUsesTransferTableCycle(t *testing.T) {
	bot := "0x00000000000000000000000000000000000000aa"
	pool1 := "0x00000000000000000000000000000000000000bb"
	pool2 := "0x00000000000000000000000000000000000000cc"
	weth := "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"
	tokenX := "0x00000000000000000000000000000000000000dd"
	block := Block{Number: "0x1", Timestamp: "0x2"}
	tx := BlockTx{
		Hash:             "0xtransfercycle",
		TransactionIndex: "0x1",
		From:             "0x00000000000000000000000000000000000000ee",
		To:               bot,
		Gas:              "0x5208",
	}
	receipt := Receipt{
		Status:            "0x1",
		GasUsed:           "0x5208",
		EffectiveGasPrice: "0x1",
		Logs: []ReceiptLog{
			erc20TransferLog(weth, bot, pool1, "0x64"),
			erc20TransferLog(tokenX, pool1, pool2, "0xc8"),
			erc20TransferLog(weth, pool2, bot, "0x6e"),
		},
	}

	candidate := IdentifyOnlineArb(block, tx, receipt)
	if candidate.Kind != "transfer_cycle_candidate" {
		t.Fatalf("kind = %s", candidate.Kind)
	}
	if candidate.TransferAnalysis.ProfitedCycleCount != 1 {
		t.Fatalf("transfer analysis = %+v", candidate.TransferAnalysis)
	}
	if got := FormatReasons(candidate.Reasons); got != "transfer-table-profited-cycle" {
		t.Fatalf("reasons = %q", got)
	}
}

func TestIdentifyOnlineArbRejectsOpenPathWETHRevenue(t *testing.T) {
	bot := "0x00000000000000000000000000000000000000aa"
	pool := "0x00000000000000000000000000000000000000bb"
	manager := "0x00000000000000000000000000000000000000cc"
	usdt := "0xdac17f958d2ee523a2206206994597c13d831ec7"
	tokenX := "0xe0e6f1575be24bbf7e30eea7b924d655ead061e7"
	block := Block{Number: "0x1", Timestamp: "0x2"}
	tx := BlockTx{
		Hash:             "0xpartialcycle",
		TransactionIndex: "0x1",
		From:             "0x00000000000000000000000000000000000000ee",
		To:               bot,
		Gas:              "0x5208",
	}
	receipt := Receipt{
		Status:            "0x1",
		GasUsed:           "0x5208",
		EffectiveGasPrice: "0x1",
		Logs: []ReceiptLog{
			erc20TransferLog(usdt, pool, bot, "0x64"),
			erc20TransferLog(usdt, bot, manager, "0x64"),
			erc20TransferLog(tokenX, manager, bot, "0xc8"),
			erc20TransferLog(tokenX, bot, manager, "0xc8"),
			wethDepositLog(bot, "0x6e"),
			erc20TransferLog(WETH, bot, pool, "0x64"),
			swapLog(pool, "0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822"),
		},
	}

	candidate := IdentifyOnlineArb(block, tx, receipt)
	if candidate.Kind != "other" {
		t.Fatalf("kind = %s, want other: %+v", candidate.Kind, candidate.TransferAnalysis)
	}
	if got := FormatReasons(candidate.Reasons); got != "" {
		t.Fatalf("reasons = %q", got)
	}
}

func TestIdentifyOnlineArbRejectsVaultRebalanceWithoutCycle(t *testing.T) {
	vault := "0x00000000000000000000000000000000000000a1"
	router := "0x00000000000000000000000000000000000000a2"
	swapper := "0x00000000000000000000000000000000000000a3"
	pool := "0x00000000000000000000000000000000000000a4"
	wstETH := "0x7f39c581f595b53c5cb19bd0b3f8da6c935e2ca0"
	wrapper := "0x8ff4d4a12a4df051abfbd1ff94f6b17d42cc3856"
	block := Block{Number: "0x1", Timestamp: "0x2"}
	tx := BlockTx{
		Hash:             "0xvaultrebalance",
		TransactionIndex: "0x1",
		From:             "0x00000000000000000000000000000000000000ee",
		To:               wrapper,
		Input:            "0xf845c3ae",
		Gas:              "0x5208",
	}
	receipt := Receipt{
		Status:            "0x1",
		GasUsed:           "0x5208",
		EffectiveGasPrice: "0x1",
		Logs: []ReceiptLog{
			erc20TransferLog(WETH, vault, router, "0x64"),
			erc20TransferLog(WETH, router, swapper, "0x64"),
			erc20TransferLog(wstETH, pool, swapper, "0x5a"),
			erc20TransferLog(WETH, swapper, pool, "0x64"),
			swapLog(pool, "0xc42079f94a6350d7e6235f29174924f928cc2ac818eb64fed8004e115fbcca67"),
			erc20TransferLog(wstETH, swapper, vault, "0x5a"),
		},
	}

	candidate := IdentifyOnlineArb(block, tx, receipt)
	if candidate.Kind != "other" {
		t.Fatalf("kind = %s, want other: %+v", candidate.Kind, candidate.TransferAnalysis)
	}
	if candidate.TransferAnalysis.Status != "non_profit_cycle" {
		t.Fatalf("transfer status = %s, want non_profit_cycle: %+v", candidate.TransferAnalysis.Status, candidate.TransferAnalysis)
	}
	if candidate.TransferAnalysis.ProfitedCycleCount != 0 {
		t.Fatalf("profited cycles = %d", candidate.TransferAnalysis.ProfitedCycleCount)
	}
}

func wethDepositLog(dst, amountHex string) ReceiptLog {
	return ReceiptLog{
		Address: WETH,
		Topics: []string{
			wethDepositTopic,
			"0x000000000000000000000000" + dst[2:],
		},
		Data: amountHex,
	}
}

func wethWithdrawalLog(src, amountHex string) ReceiptLog {
	return ReceiptLog{
		Address: WETH,
		Topics: []string{
			wethWithdrawalTopic,
			"0x000000000000000000000000" + src[2:],
		},
		Data: amountHex,
	}
}
