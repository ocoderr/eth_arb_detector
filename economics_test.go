package detector

import "testing"

func TestBuildTxEconomicsWethRevenueAndProfit(t *testing.T) {
	receipt := Receipt{
		GasUsed:           "0x64",
		EffectiveGasPrice: "0xa",
		Logs: []ReceiptLog{
			erc20TransferLog(WETH, "0x0000000000000000000000000000000000000001", "0x0000000000000000000000000000000000000002", "0x3e8"),
			erc20TransferLog(WETH, "0x0000000000000000000000000000000000000002", "0x0000000000000000000000000000000000000003", "0x64"),
		},
	}

	economics := BuildTxEconomics(BlockTx{To: "0x0000000000000000000000000000000000000002"}, receipt)
	if economics.TxFeeWei != "1000" {
		t.Fatalf("tx fee = %s", economics.TxFeeWei)
	}
	if economics.WethDeltaToToWei != "900" {
		t.Fatalf("weth delta = %s", economics.WethDeltaToToWei)
	}
	if economics.WethProfitToToWei != "-100" {
		t.Fatalf("weth profit = %s", economics.WethProfitToToWei)
	}
}

func TestBuildTxEconomicsCountsWETHDepositAsRevenue(t *testing.T) {
	target := "0x0000000000000000000000000000000000000002"
	receipt := Receipt{
		GasUsed:           "0x0",
		EffectiveGasPrice: "0x0",
		Logs: []ReceiptLog{
			wethDepositLog(target, "0x6e"),
			erc20TransferLog(WETH, target, "0x0000000000000000000000000000000000000003", "0x64"),
			wethWithdrawalLog(target, "0xa"),
		},
	}

	economics := BuildTxEconomics(BlockTx{To: target}, receipt)
	if economics.WethRevenueToToWei != "10" {
		t.Fatalf("revenue = %s", economics.WethRevenueToToWei)
	}
	if economics.WethTransferInWei != "110" || economics.WethTransferOutWei != "100" {
		t.Fatalf("flow in=%s out=%s", economics.WethTransferInWei, economics.WethTransferOutWei)
	}
}

func erc20TransferLog(token, from, to, amountHex string) ReceiptLog {
	return ReceiptLog{
		Address: token,
		Topics: []string{
			ERC20TransferTopic,
			"0x000000000000000000000000" + from[2:],
			"0x000000000000000000000000" + to[2:],
		},
		Data: amountHex,
	}
}
