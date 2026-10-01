package detector

import (
	"fmt"
	"math/big"
	"sort"
	"strings"
)

const ERC20TransferTopic = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"

var SwapLikeEventTopics = map[string]string{
	"0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822": "UniV2Swap",
	"0xc42079f94a6350d7e6235f29174924f928cc2ac818eb64fed8004e115fbcca67": "UniV3Swap",
	"0x40e9cecb9f5f1f1c5b9c97dec2917b7ee92e57ba5563708daca94dd84ad7112f": "UniV4Swap",
	"0x19b47279256b2a23a1665c810c8d55a1758940ee09377d4f8d26497a3577dc83": "PancakeV3Swap",
	"0x2170c741c41531aec20e7c107c24eecfdd15e69c9bb0a8dd37b1840b9e0b207b": "BalancerSwap",
	"0x7154b38b5dd31bb3122436a96d4e09aba5b323ae1fd580025fab55074334c095": "BancorSwap",
	"0x8b3e96f2b889fa771c53c981b40daf005f63f637f1869f707052d15a3dd97140": "Curve10Swap",
	"0xd013ca23e77a65003c2c659c5442c00c805371b7fc1ebd4c206c41d1536bd90b": "Curve10UnderlyingSwap",
	"0x143f1f8e861fbdeddd5b46e844b7d3ac7b86a122f36e8c463859ee6811b1f29c": "Curve11V2Swap",
	"0xb2e76ae99761dc136e598d4a629bb347eccb9532a5f8bbd72e18467c3c34cc98": "Curve11Swap",
	"0xbd99c6719f088aa0abd9e7b7a4a635d1f931601e9f304b538dc42be25d8c65c6": "MooniswapSwap",
	"0x86c49b5d8577da08444947f1427d23ef191cfabf2c0788f93324d79e926a9302": "MooniswapV1Swap",
	"0xc2c0245e056d5fb095f04cd6373bc770802ebd1e6c918eb78fdef843cdb37b0f": "DODOSwap",
	"0x103ed084e94a44c8f5f6ba8e3011507c41063177e29949083c439777d8d63f60": "MaverickV2Swap",
}

type Block struct {
	Number       string    `json:"number"`
	Hash         string    `json:"hash"`
	Timestamp    string    `json:"timestamp"`
	Miner        string    `json:"miner"`
	Transactions []BlockTx `json:"transactions"`
}

type BlockTx struct {
	Hash                 string `json:"hash"`
	TransactionIndex     string `json:"transactionIndex"`
	From                 string `json:"from"`
	To                   string `json:"to"`
	Input                string `json:"input"`
	Value                string `json:"value"`
	Gas                  string `json:"gas"`
	GasPrice             string `json:"gasPrice"`
	MaxFeePerGas         string `json:"maxFeePerGas"`
	MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas"`
}

type Receipt struct {
	TransactionHash   string       `json:"transactionHash"`
	BlockNumber       string       `json:"blockNumber"`
	TransactionIndex  string       `json:"transactionIndex"`
	Status            string       `json:"status"`
	GasUsed           string       `json:"gasUsed"`
	EffectiveGasPrice string       `json:"effectiveGasPrice"`
	Logs              []ReceiptLog `json:"logs"`
}

type ReceiptLog struct {
	Address string   `json:"address"`
	Topics  []string `json:"topics"`
	Data    string   `json:"data"`
}

type DiscoveryCandidate struct {
	TxHash             string
	TxIndex            int64
	From               string
	To                 string
	Status             string
	InputSelector      string
	ValueWei           string
	GasLimit           int64
	GasPriceWei        string
	MaxFeePerGasWei    string
	MaxPriorityFeeWei  string
	GasUsed            int64
	EffectiveGasPrice  int64
	Logs               int
	UniqueLogContracts int
	TransferLogs       int
	SwapLikeLogs       int
	UniqueTokenLogs    int
	Score              int
	Reasons            []string
}

func DiscoverTransaction(tx BlockTx, receipt Receipt) DiscoveryCandidate {
	logContracts := map[string]bool{}
	tokenContracts := map[string]bool{}
	swapKinds := map[string]bool{}
	c := DiscoveryCandidate{
		TxHash:            tx.Hash,
		TxIndex:           HexOrDecimalInt(tx.TransactionIndex),
		From:              strings.ToLower(tx.From),
		To:                strings.ToLower(tx.To),
		Status:            receipt.Status,
		InputSelector:     Selector(tx.Input),
		ValueWei:          HexOrDecimalBigString(tx.Value),
		GasLimit:          HexOrDecimalInt(tx.Gas),
		GasPriceWei:       HexOrDecimalBigString(tx.GasPrice),
		MaxFeePerGasWei:   HexOrDecimalBigString(tx.MaxFeePerGas),
		MaxPriorityFeeWei: HexOrDecimalBigString(tx.MaxPriorityFeePerGas),
		GasUsed:           HexOrDecimalInt(receipt.GasUsed),
		EffectiveGasPrice: HexOrDecimalInt(receipt.EffectiveGasPrice),
		Logs:              len(receipt.Logs),
	}
	for _, log := range receipt.Logs {
		addr := strings.ToLower(log.Address)
		if addr != "" {
			logContracts[addr] = true
		}
		if len(log.Topics) == 0 {
			continue
		}
		topic0 := strings.ToLower(log.Topics[0])
		if topic0 == ERC20TransferTopic {
			c.TransferLogs++
			if addr != "" {
				tokenContracts[addr] = true
			}
			continue
		}
		if kind, ok := SwapLikeEventTopics[topic0]; ok {
			c.SwapLikeLogs++
			swapKinds[kind] = true
		}
	}
	c.UniqueLogContracts = len(logContracts)
	c.UniqueTokenLogs = len(tokenContracts)
	c.Score = discoveryScore(c)
	c.Reasons = discoveryReasons(c, swapKinds)
	return c
}

func (c DiscoveryCandidate) IsCandidate() bool {
	return len(c.Reasons) > 0
}

func SortDiscoveryCandidates(candidates []DiscoveryCandidate) {
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score > candidates[j].Score
		}
		return candidates[i].TxIndex < candidates[j].TxIndex
	})
}

func FormatReasons(reasons []string) string {
	return strings.Join(reasons, ",")
}

func discoveryScore(c DiscoveryCandidate) int {
	score := 0
	score += c.SwapLikeLogs * 12
	score += c.TransferLogs * 2
	score += c.UniqueTokenLogs * 5
	score += c.UniqueLogContracts * 2
	if c.SwapLikeLogs >= 2 {
		score += 20
	}
	if c.TransferLogs >= 6 && c.UniqueTokenLogs >= 2 && c.UniqueLogContracts >= 4 {
		score += 15
	}
	if c.Status != "" && c.Status != "0x1" {
		score -= 30
	}
	return score
}

func discoveryReasons(c DiscoveryCandidate, swapKinds map[string]bool) []string {
	var reasons []string
	if c.Status != "" && c.Status != "0x1" {
		return reasons
	}
	if c.SwapLikeLogs >= 2 {
		reasons = append(reasons, fmt.Sprintf("multi-swap-events:%d", c.SwapLikeLogs))
	}
	if c.SwapLikeLogs >= 1 && c.TransferLogs >= 4 && c.UniqueTokenLogs >= 2 {
		reasons = append(reasons, "swap-plus-token-flow")
	}
	if c.TransferLogs >= 6 && c.UniqueTokenLogs >= 2 && c.UniqueLogContracts >= 4 {
		reasons = append(reasons, "transfer-fanout-possible-unknown-arb")
	}
	if len(swapKinds) >= 2 {
		reasons = append(reasons, fmt.Sprintf("multiple-swap-families:%d", len(swapKinds)))
	}
	if c.MaxPriorityFeeWei == "0" && c.ValueWei != "" && c.ValueWei != "0" {
		reasons = append(reasons, "zero-priority-with-value")
	}
	return reasons
}

func HexOrDecimalBigString(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	n := new(big.Int)
	if strings.HasPrefix(v, "0x") || strings.HasPrefix(v, "0X") {
		if len(v) == 2 {
			return "0"
		}
		if _, ok := n.SetString(v[2:], 16); !ok {
			return ""
		}
		return n.String()
	}
	if _, ok := n.SetString(v, 10); !ok {
		return ""
	}
	return n.String()
}
