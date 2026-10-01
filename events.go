package detector

import "strings"

const (
	BalancerVault = "0xba12222222228d8ba445958a75a0704d566bf2c8"
	SwaapVault    = "0xd315a9c38ec871068fec378e4ce78af528c76293"
	FluidVault    = "0x52aa899454998be5b000ad077a46bbe360f4e497"
	BancorSwap    = "0x1f573d6fb3f13d689ff844b4ce37794d79a7ff1c"
)

type EventKind string

const (
	EventKindSwap          EventKind = "swap"
	EventKindLiquidity     EventKind = "liquidity"
	EventKindPoolCreated   EventKind = "pool_created"
	EventKindLimitOrder    EventKind = "limit_order"
	EventKindTransfer      EventKind = "transfer"
	EventKindOtherProtocol EventKind = "other_protocol"
)

type ProtocolEvent struct {
	LogIndex string    `json:"logIndex,omitempty"`
	Kind     EventKind `json:"kind"`
	Protocol string    `json:"protocol"`
	Pool     string    `json:"pool,omitempty"`
	Address  string    `json:"address,omitempty"`
	TokenIn  string    `json:"tokenIn,omitempty"`
	TokenOut string    `json:"tokenOut,omitempty"`
	Topic0   string    `json:"topic0"`
}

type ReceiptEventSummary struct {
	Events          []ProtocolEvent `json:"events,omitempty"`
	SwapPath        []ProtocolEvent `json:"swapPath,omitempty"`
	SwapEvents      int             `json:"swapEvents"`
	TransferEvents  int             `json:"transferEvents"`
	TransferTokens  []string        `json:"transferTokens,omitempty"`
	Protocols       []string        `json:"protocols,omitempty"`
	Pools           []string        `json:"pools,omitempty"`
	UniqueProtocols int             `json:"uniqueProtocols"`
	UniquePools     int             `json:"uniquePools"`
}

type topicInfo struct {
	Kind     EventKind
	Protocol string
	Pool     string
}

type tokenFlow struct {
	in  map[string]struct{}
	out map[string]struct{}
}

var protocolTopicMap = map[string]topicInfo{
	"0x0d3648bd0f6ba80134a33ba9275ac585d9d315f0ad8355cddefde31afa28d0e9": {EventKindPoolCreated, "UniV2", "topic_pair_created"},
	"0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822": {EventKindSwap, "UniV2", "log_address"},
	"0x1c411e9a96e071241c2f21f7726b17ae89e3cab4c78be50e062b03a9fffbbad1": {EventKindLiquidity, "UniV2", "log_address"},

	"0x783cca1c0412dd0d695e784568c96da2e9c22ff989357a2e8b1d9b2b4e6b7118": {EventKindPoolCreated, "UniV3", "data_pool_created"},
	"0xc42079f94a6350d7e6235f29174924f928cc2ac818eb64fed8004e115fbcca67": {EventKindSwap, "UniV3", "log_address"},
	"0x7a53080ba414158be7ec69b987b5fb7d07dee101fe85488f0853ae16239d0bde": {EventKindLiquidity, "UniV3", "log_address"},
	"0x0c396cd989a39f4459b5fa1aed6a9a8dcdbc45908acfd67e028cd568da98982c": {EventKindLiquidity, "UniV3", "log_address"},

	"0xdd466e674ea557f56295e2d0218a125ea4b4f0f6f3307b95f85e6110838d6438": {EventKindPoolCreated, "UniV4", "topic_pool_id"},
	"0x40e9cecb9f5f1f1c5b9c97dec2917b7ee92e57ba5563708daca94dd84ad7112f": {EventKindSwap, "UniV4", "topic_pool_id"},
	"0xf208f4912782fd25c7f114ca3723a2d5dd6f3bcc3ac8db5af63baa85f711d5ec": {EventKindLiquidity, "UniV4", "topic_pool_id"},

	"0x19b47279256b2a23a1665c810c8d55a1758940ee09377d4f8d26497a3577dc83": {EventKindSwap, "PancakeV3", "log_address"},
	"0x2170c741c41531aec20e7c107c24eecfdd15e69c9bb0a8dd37b1840b9e0b207b": {EventKindSwap, "BalancerLike", "topic_pool_id"},
	"0xe5ce249087ce04f05a957192435400fd97868dba0e6a4b4c049abf8af80dae78": {EventKindLiquidity, "BalancerLike", "topic_pool_id"},

	"0x7154b38b5dd31bb3122436a96d4e09aba5b323ae1fd580025fab55074334c095": {EventKindSwap, "BancorSwap", "topic_pool_address"},
	"0x8b3e96f2b889fa771c53c981b40daf005f63f637f1869f707052d15a3dd97140": {EventKindSwap, "Curve10", "log_address"},
	"0xd013ca23e77a65003c2c659c5442c00c805371b7fc1ebd4c206c41d1536bd90b": {EventKindSwap, "Curve10Underlying", "log_address"},
	"0x143f1f8e861fbdeddd5b46e844b7d3ac7b86a122f36e8c463859ee6811b1f29c": {EventKindSwap, "Curve11V2", "log_address"},
	"0xb2e76ae99761dc136e598d4a629bb347eccb9532a5f8bbd72e18467c3c34cc98": {EventKindSwap, "Curve11", "log_address"},

	"0xb3e2773606abfd36b5bd91394b3a54d1398336c65005baf7bf7a05efeffaf75b": {EventKindSwap, "yETH", "log_address"},
	"0xbd99c6719f088aa0abd9e7b7a4a635d1f931601e9f304b538dc42be25d8c65c6": {EventKindSwap, "Mooniswap", "log_address"},
	"0x86c49b5d8577da08444947f1427d23ef191cfabf2c0788f93324d79e926a9302": {EventKindSwap, "MooniswapV1", "log_address"},
	"0xc2c0245e056d5fb095f04cd6373bc770802ebd1e6c918eb78fdef843cdb37b0f": {EventKindSwap, "DODO", "log_address"},
	"0x103ed084e94a44c8f5f6ba8e3011507c41063177e29949083c439777d8d63f60": {EventKindSwap, "MaverickV2", "log_address"},
	"0x4d93b232a24e82b284ced7461bf4deacffe66759d5c24513e6f29e571ad78d15": {EventKindOtherProtocol, "Fluid", "log_address"},

	"0xfec331350fce78ba658e082a71da20ac9f8d798a99b3c79681c8440cbfe77e07": {EventKindLimitOrder, "1inchLO", "log_address"},
	"0x78ad7ec0e9f89e74012afa58738b6b661c024cb0fd185ee2f616c0a28924bd66": {EventKindLimitOrder, "UnixLO", "log_address"},
}

func ParseReceiptEvents(receipt Receipt) ReceiptEventSummary {
	var out ReceiptEventSummary
	protocols := map[string]struct{}{}
	pools := map[string]struct{}{}
	transferTokens := map[string]struct{}{}
	flowByAddress := receiptTokenFlows(receipt)
	for _, eventLog := range receipt.Logs {
		if len(eventLog.Topics) == 0 {
			continue
		}
		topic0 := strings.ToLower(eventLog.Topics[0])
		address := strings.ToLower(eventLog.Address)
		if topic0 == ERC20TransferTopic {
			out.TransferEvents++
			if address != "" {
				transferTokens[address] = struct{}{}
			}
			continue
		}
		info, ok := protocolTopicMap[topic0]
		if !ok {
			continue
		}
		protocol := refineProtocol(info.Protocol, address)
		pool := eventPool(info.Pool, eventLog)
		if protocol != "" {
			protocols[protocol] = struct{}{}
		}
		if pool != "" {
			pools[pool] = struct{}{}
		}
		event := ProtocolEvent{
			Kind:     info.Kind,
			Protocol: protocol,
			Pool:     pool,
			Address:  address,
			Topic0:   topic0,
		}
		event.TokenIn, event.TokenOut = eventTokenDirection(protocol, eventLog, pool, flowByAddress)
		if info.Kind == EventKindSwap {
			out.SwapEvents++
			out.SwapPath = append(out.SwapPath, event)
			continue
		}
		out.Events = append(out.Events, event)
	}
	out.TransferTokens = sortedKeys(transferTokens)
	out.Protocols = sortedKeys(protocols)
	out.Pools = sortedKeys(pools)
	out.UniqueProtocols = len(out.Protocols)
	out.UniquePools = len(out.Pools)
	return out
}

func eventTokenDirection(protocol string, eventLog ReceiptLog, pool string, flows map[string]tokenFlow) (string, string) {
	switch protocol {
	case "Balancer", "Swaap", "BalancerLike":
		return indexedAddress(eventLog.Topics, 2), indexedAddress(eventLog.Topics, 3)
	}
	if flow, ok := flows[pool]; ok {
		return singleToken(flow.in), singleToken(flow.out)
	}
	return "", ""
}

func refineProtocol(protocol, address string) string {
	if protocol != "BalancerLike" {
		return protocol
	}
	switch strings.ToLower(address) {
	case BalancerVault:
		return "Balancer"
	case SwaapVault:
		return "Swaap"
	default:
		return "BalancerLike"
	}
}

func eventPool(kind string, eventLog ReceiptLog) string {
	switch kind {
	case "log_address":
		return strings.ToLower(eventLog.Address)
	case "topic_pool_id":
		if len(eventLog.Topics) > 1 {
			return strings.ToLower(eventLog.Topics[1])
		}
	case "topic_pool_address":
		return indexedAddress(eventLog.Topics, 1)
	}
	return ""
}

func indexedAddress(topics []string, idx int) string {
	if len(topics) <= idx {
		return ""
	}
	topic := strings.TrimPrefix(strings.ToLower(topics[idx]), "0x")
	if len(topic) < 40 {
		return ""
	}
	return "0x" + topic[len(topic)-40:]
}

func receiptTokenFlows(receipt Receipt) map[string]tokenFlow {
	flows := map[string]tokenFlow{}
	for _, eventLog := range receipt.Logs {
		if len(eventLog.Topics) < 3 || strings.ToLower(eventLog.Topics[0]) != ERC20TransferTopic {
			continue
		}
		token := strings.ToLower(eventLog.Address)
		from := indexedAddress(eventLog.Topics, 1)
		to := indexedAddress(eventLog.Topics, 2)
		if token == "" {
			continue
		}
		if from != "" {
			flow := flows[from]
			if flow.out == nil {
				flow.out = map[string]struct{}{}
			}
			flow.out[token] = struct{}{}
			flows[from] = flow
		}
		if to != "" {
			flow := flows[to]
			if flow.in == nil {
				flow.in = map[string]struct{}{}
			}
			flow.in[token] = struct{}{}
			flows[to] = flow
		}
	}
	return flows
}

func singleToken(tokens map[string]struct{}) string {
	if len(tokens) != 1 {
		return ""
	}
	for token := range tokens {
		return token
	}
	return ""
}

func sortedKeys(values map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}
