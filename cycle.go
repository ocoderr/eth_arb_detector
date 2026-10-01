package detector

type CycleInfo struct {
	Status      string   `json:"status"`
	Reason      string   `json:"reason,omitempty"`
	TokenPath   []string `json:"tokenPath,omitempty"`
	KnownHops   int      `json:"knownHops"`
	UnknownHops int      `json:"unknownHops"`
}

func DetectCycle(events ReceiptEventSummary, economics TxEconomics) CycleInfo {
	if len(events.SwapPath) == 0 {
		return CycleInfo{Status: "unknown", Reason: "no_swap_path"}
	}

	path := make([]string, 0, len(events.SwapPath)+1)
	known := 0
	unknown := 0
	continuous := true
	for _, hop := range events.SwapPath {
		if hop.TokenIn == "" || hop.TokenOut == "" {
			unknown++
			continue
		}
		known++
		if len(path) == 0 {
			path = append(path, hop.TokenIn, hop.TokenOut)
			continue
		}
		if path[len(path)-1] != hop.TokenIn {
			continuous = false
		}
		path = append(path, hop.TokenOut)
	}

	if unknown > 0 {
		info := CycleInfo{
			Status:      "unknown",
			Reason:      "missing_token_direction",
			TokenPath:   path,
			KnownHops:   known,
			UnknownHops: unknown,
		}
		if economics.WethTransferInWei != "" && economics.WethTransferOutWei != "" &&
			economics.WethTransferInWei != "0" && economics.WethTransferOutWei != "0" {
			info.Reason = "missing_token_direction_with_weth_roundtrip"
		}
		return info
	}
	if len(path) < 2 {
		return CycleInfo{Status: "unknown", Reason: "no_known_token_path", KnownHops: known}
	}
	if !continuous {
		return CycleInfo{
			Status:    "non_cycle",
			Reason:    "token_path_not_continuous",
			TokenPath: path,
			KnownHops: known,
		}
	}
	if path[0] == path[len(path)-1] {
		return CycleInfo{
			Status:    "cycle",
			Reason:    "closed_token_path",
			TokenPath: path,
			KnownHops: known,
		}
	}
	return CycleInfo{
		Status:    "non_cycle",
		Reason:    "open_token_path",
		TokenPath: path,
		KnownHops: known,
	}
}
