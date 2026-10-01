package detector

import (
	"encoding/json"
	"strconv"
	"strings"
)

type RangeTx struct {
	Hash             string `json:"hash"`
	BlockNumber      string `json:"blockNumber"`
	TransactionIndex string `json:"transactionIndex"`
	From             string `json:"from"`
	To               string `json:"to"`
	Input            string `json:"input"`
}

type TraceEntry struct {
	Type         string          `json:"type"`
	TraceAddr    []int           `json:"traceAddress"`
	To           string          `json:"to"`
	Input        string          `json:"input"`
	Value        string          `json:"value"`
	Error        string          `json:"error"`
	RawTraceAddr json.RawMessage `json:"-"`
}

type TraceResult struct {
	Trace []TraceEntry `json:"trace"`
}

func CallsFromTenderlyTrace(result TraceResult) []Call {
	calls := make([]Call, 0, len(result.Trace))
	for _, entry := range result.Trace {
		calls = append(calls, Call{
			Type:  entry.Type,
			Depth: len(entry.TraceAddr),
			To:    entry.To,
			Input: entry.Input,
			Value: entry.Value,
			Error: entry.Error,
		})
	}
	return calls
}

func HexOrDecimalInt(v string) int64 {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0
	}
	if strings.HasPrefix(v, "0x") || strings.HasPrefix(v, "0X") {
		n, _ := strconv.ParseInt(v[2:], 16, 64)
		return n
	}
	n, _ := strconv.ParseInt(v, 10, 64)
	return n
}
