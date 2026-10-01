package detector

import (
	"math/big"
	"strings"
)

const WETH = "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"

type TxEconomics struct {
	TxFeeWei             string `json:"txFeeWei,omitempty"`
	CoinbasePaymentWei   string `json:"coinbasePaymentWei,omitempty"`
	WethDeltaToToWei     string `json:"wethDeltaToToWei,omitempty"`
	WethRevenueToToWei   string `json:"wethRevenueToToWei,omitempty"`
	WethProfitToToWei    string `json:"wethProfitToToWei,omitempty"`
	WethNetProfitWei     string `json:"wethNetProfitWei,omitempty"`
	WethTransferInWei    string `json:"wethTransferInWei,omitempty"`
	WethTransferOutWei   string `json:"wethTransferOutWei,omitempty"`
	ProfitCalcConfidence string `json:"profitCalcConfidence,omitempty"`
}

func BuildTxEconomics(tx BlockTx, receipt Receipt) TxEconomics {
	txFee := new(big.Int).Mul(
		big.NewInt(HexOrDecimalInt(receipt.GasUsed)),
		big.NewInt(HexOrDecimalInt(receipt.EffectiveGasPrice)),
	)
	in, out := wethTransferFlow(receipt, tx.To)
	delta := new(big.Int).Sub(in, out)
	economics := TxEconomics{
		TxFeeWei:           txFee.String(),
		WethDeltaToToWei:   delta.String(),
		WethTransferInWei:  in.String(),
		WethTransferOutWei: out.String(),
	}
	if delta.Sign() > 0 {
		economics.WethRevenueToToWei = delta.String()
		economics.WethProfitToToWei = new(big.Int).Sub(delta, txFee).String()
		economics.ProfitCalcConfidence = "weth-delta-minus-tx-fee"
	} else {
		economics.ProfitCalcConfidence = "tx-fee-only"
	}
	return economics
}

func BuildTracedTxEconomics(tx BlockTx, receipt Receipt, miner string, trace TraceResult) TxEconomics {
	economics := BuildTxEconomics(tx, receipt)
	coinbasePayment := TraceValueTo(trace, miner)
	economics.CoinbasePaymentWei = coinbasePayment.String()
	revenue := decimalBig(economics.WethRevenueToToWei)
	txFee := decimalBig(economics.TxFeeWei)
	if revenue.Sign() > 0 {
		net := new(big.Int).Sub(revenue, txFee)
		net.Sub(net, coinbasePayment)
		economics.WethNetProfitWei = net.String()
		economics.ProfitCalcConfidence = "weth-delta-minus-tx-fee-minus-coinbase-payment"
	}
	return economics
}

func TraceValueTo(trace TraceResult, target string) *big.Int {
	target = strings.ToLower(target)
	out := new(big.Int)
	if target == "" {
		return out
	}
	for _, entry := range trace.Trace {
		if entry.Error != "" || strings.ToLower(entry.To) != target {
			continue
		}
		value := decimalOrHexBig(entry.Value)
		if value == nil {
			continue
		}
		out.Add(out, value)
	}
	return out
}

func wethTransferFlow(receipt Receipt, target string) (*big.Int, *big.Int) {
	target = strings.ToLower(target)
	in := new(big.Int)
	out := new(big.Int)
	if target == "" {
		return in, out
	}
	for _, eventLog := range receipt.Logs {
		if strings.ToLower(eventLog.Address) != WETH || len(eventLog.Topics) == 0 {
			continue
		}
		topic0 := strings.ToLower(eventLog.Topics[0])
		if topic0 == wethDepositTopic {
			if len(eventLog.Topics) < 2 || indexedAddress(eventLog.Topics, 1) != target {
				continue
			}
			amount := hexBig(eventLog.Data)
			if amount != nil {
				in.Add(in, amount)
			}
			continue
		}
		if topic0 != ERC20TransferTopic || len(eventLog.Topics) < 3 {
			continue
		}
		amount := hexBig(eventLog.Data)
		if amount == nil {
			continue
		}
		from := indexedAddress(eventLog.Topics, 1)
		to := indexedAddress(eventLog.Topics, 2)
		switch {
		case to == target:
			in.Add(in, amount)
		case from == target:
			out.Add(out, amount)
		}
	}
	return in, out
}

func hexBig(s string) *big.Int {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	s = strings.TrimPrefix(strings.TrimPrefix(s, "0x"), "0X")
	if s == "" {
		return big.NewInt(0)
	}
	n := new(big.Int)
	if _, ok := n.SetString(s, 16); !ok {
		return nil
	}
	return n
}

func decimalBig(s string) *big.Int {
	n := new(big.Int)
	if s == "" {
		return n
	}
	if _, ok := n.SetString(s, 10); !ok {
		return big.NewInt(0)
	}
	return n
}

func decimalOrHexBig(s string) *big.Int {
	s = strings.TrimSpace(s)
	if s == "" {
		return big.NewInt(0)
	}
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		return hexBig(s)
	}
	return decimalBig(s)
}
