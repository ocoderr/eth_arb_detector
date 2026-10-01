package detector

import (
	"math/big"
	"sort"
	"strings"
)

const (
	wethDepositTopic    = "0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c"
	wethWithdrawalTopic = "0x7fcf532c15f0a6db0bd6d0e038bea71d30d808c7d98cb3bf7268a95bf5081b65"
	zeroAddress         = "0x0000000000000000000000000000000000000000"
)

type TransferTableAnalysis struct {
	Status             string              `json:"status"`
	Reason             string              `json:"reason,omitempty"`
	TransferCount      int                 `json:"transferCount"`
	AddressCount       int                 `json:"addressCount"`
	TokenCount         int                 `json:"tokenCount"`
	SCCCount           int                 `json:"sccCount"`
	ProfitedCycleCount int                 `json:"profitedCycleCount"`
	ProfitedCycles     []TransferCycleInfo `json:"profitedCycles,omitempty"`
}

type TransferCycleInfo struct {
	AddressCount       int      `json:"addressCount"`
	TransferCount      int      `json:"transferCount"`
	TradeAddressCount  int      `json:"tradeAddressCount"`
	ProfitAddressCount int      `json:"profitAddressCount"`
	Tokens             []string `json:"tokens,omitempty"`
	ProfitTokens       []string `json:"profitTokens,omitempty"`
}

type transferRow struct {
	From   string
	To     string
	Token  string
	Amount *big.Int
}

type transferVertex struct {
	address string
	edges   []int
	index   int
	low     int
	onStack bool
}

func AnalyzeTransferTable(receipt Receipt, tx BlockTx) TransferTableAnalysis {
	transfers := receiptTransfers(receipt)
	addresses := map[string]struct{}{}
	tokens := map[string]struct{}{}
	for _, tr := range transfers {
		if tr.From != "" {
			addresses[tr.From] = struct{}{}
		}
		if tr.To != "" {
			addresses[tr.To] = struct{}{}
		}
		if tr.Token != "" {
			tokens[tr.Token] = struct{}{}
		}
	}
	out := TransferTableAnalysis{
		TransferCount: len(transfers),
		AddressCount:  len(addresses),
		TokenCount:    len(tokens),
	}
	if len(transfers) < 3 {
		out.Status = "unknown"
		out.Reason = "too_few_transfers"
		return out
	}

	components := transferSCCs(transfers)
	out.SCCCount = len(components)
	txFrom := strings.ToLower(tx.From)
	for _, component := range components {
		if len(component) <= 1 {
			continue
		}
		info, ok := analyzeTransferComponent(transfers, component, txFrom)
		if !ok {
			continue
		}
		out.ProfitedCycles = append(out.ProfitedCycles, info)
	}
	out.ProfitedCycleCount = len(out.ProfitedCycles)
	if out.ProfitedCycleCount > 0 {
		out.Status = "profited_cycle"
		out.Reason = "transfer_scc_net_profit"
		if len(out.ProfitedCycles) > 3 {
			out.ProfitedCycles = out.ProfitedCycles[:3]
		}
		return out
	}
	if len(components) > 0 {
		out.Status = "non_profit_cycle"
		out.Reason = "scc_without_profit"
		return out
	}
	out.Status = "non_cycle"
	out.Reason = "no_scc"
	return out
}

func receiptTransfers(receipt Receipt) []transferRow {
	out := make([]transferRow, 0)
	for _, eventLog := range receipt.Logs {
		if len(eventLog.Topics) == 0 {
			continue
		}
		topic0 := strings.ToLower(eventLog.Topics[0])
		token := strings.ToLower(eventLog.Address)
		if topic0 == ERC20TransferTopic {
			if len(eventLog.Topics) < 3 {
				continue
			}
			from := indexedAddress(eventLog.Topics, 1)
			to := indexedAddress(eventLog.Topics, 2)
			amount := hexBig(eventLog.Data)
			if token == "" || from == "" || to == "" || amount == nil || amount.Sign() == 0 {
				continue
			}
			out = append(out, transferRow{From: from, To: to, Token: token, Amount: amount})
			continue
		}
		if token == WETH && topic0 == wethDepositTopic {
			if len(eventLog.Topics) < 2 {
				continue
			}
			to := indexedAddress(eventLog.Topics, 1)
			amount := hexBig(eventLog.Data)
			if to == "" || amount == nil || amount.Sign() == 0 {
				continue
			}
			out = append(out, transferRow{From: zeroAddress, To: to, Token: WETH, Amount: amount})
			continue
		}
		if token == WETH && topic0 == wethWithdrawalTopic {
			if len(eventLog.Topics) < 2 {
				continue
			}
			to := indexedAddress(eventLog.Topics, 1)
			amount := hexBig(eventLog.Data)
			if to == "" || amount == nil || amount.Sign() == 0 {
				continue
			}
			out = append(out, transferRow{From: zeroAddress, To: to, Token: WETH, Amount: amount})
			continue
		}
	}
	return out
}

func transferSCCs(transfers []transferRow) [][]string {
	graph := map[string]*transferVertex{}
	for i, tr := range transfers {
		if _, ok := graph[tr.From]; !ok {
			graph[tr.From] = &transferVertex{address: tr.From, index: -1}
		}
		if _, ok := graph[tr.To]; !ok {
			graph[tr.To] = &transferVertex{address: tr.To, index: -1}
		}
		graph[tr.From].edges = append(graph[tr.From].edges, i)
	}

	index := 0
	stack := []*transferVertex{}
	var components [][]string
	var strongConnect func(v *transferVertex)
	strongConnect = func(v *transferVertex) {
		v.index = index
		v.low = index
		index++
		stack = append(stack, v)
		v.onStack = true

		for _, edgeIdx := range v.edges {
			w := graph[transfers[edgeIdx].To]
			if w.index < 0 {
				strongConnect(w)
				if w.low < v.low {
					v.low = w.low
				}
			} else if w.onStack && w.index < v.low {
				v.low = w.index
			}
		}

		if v.low == v.index {
			var component []string
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				w.onStack = false
				component = append(component, w.address)
				if w.address == v.address {
					break
				}
			}
			components = append(components, component)
		}
	}

	addrs := make([]string, 0, len(graph))
	for addr := range graph {
		addrs = append(addrs, addr)
	}
	sort.Strings(addrs)
	for _, addr := range addrs {
		v := graph[addr]
		if v.index < 0 {
			strongConnect(v)
		}
	}
	return components
}

func analyzeTransferComponent(transfers []transferRow, component []string, txFrom string) (TransferCycleInfo, bool) {
	addrSet := map[string]struct{}{}
	for _, addr := range component {
		addrSet[addr] = struct{}{}
	}
	return analyzeTransferNet(transfers, addrSet, txFrom)
}

func analyzeTransferNet(transfers []transferRow, addrSet map[string]struct{}, txFrom string) (TransferCycleInfo, bool) {
	net := map[string]map[string]*big.Int{}
	tokens := map[string]struct{}{}
	addresses := map[string]struct{}{}
	transferCount := 0
	for _, tr := range transfers {
		if addrSet != nil {
			if _, ok := addrSet[tr.From]; !ok {
				continue
			}
			if _, ok := addrSet[tr.To]; !ok {
				continue
			}
		}
		transferCount++
		addresses[tr.From] = struct{}{}
		addresses[tr.To] = struct{}{}
		tokens[tr.Token] = struct{}{}
		addNet(net, tr.From, tr.Token, new(big.Int).Neg(tr.Amount))
		addNet(net, tr.To, tr.Token, tr.Amount)
	}
	if transferCount < 3 {
		return TransferCycleInfo{}, false
	}

	tradeCount := 0
	profitCount := 0
	profitTokens := map[string]struct{}{}
	fromAsTrade := false
	for addr, tokenNet := range net {
		pos, neg := 0, 0
		for token, amount := range tokenNet {
			switch amount.Sign() {
			case 1:
				pos++
				profitTokens[token] = struct{}{}
			case -1:
				neg++
			}
		}
		if pos > 0 && neg > 0 {
			tradeCount++
			if addr == txFrom {
				fromAsTrade = true
			}
			continue
		}
		if pos > 0 && neg == 0 && addr != zeroAddress {
			profitCount++
			for token, amount := range tokenNet {
				if amount.Sign() > 0 {
					profitTokens[token] = struct{}{}
				}
			}
		}
	}
	minTradeCount := 2
	if _, hasZero := addresses[zeroAddress]; hasZero {
		minTradeCount = 1
	}
	if fromAsTrade || tradeCount < minTradeCount || profitCount == 0 {
		return TransferCycleInfo{}, false
	}
	return TransferCycleInfo{
		AddressCount:       len(addresses),
		TransferCount:      transferCount,
		TradeAddressCount:  tradeCount,
		ProfitAddressCount: profitCount,
		Tokens:             sortedKeys(tokens),
		ProfitTokens:       sortedKeys(profitTokens),
	}, true
}

func addNet(net map[string]map[string]*big.Int, addr, token string, delta *big.Int) {
	if _, ok := net[addr]; !ok {
		net[addr] = map[string]*big.Int{}
	}
	if _, ok := net[addr][token]; !ok {
		net[addr][token] = new(big.Int)
	}
	net[addr][token].Add(net[addr][token], delta)
}
