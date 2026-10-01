package detector

import (
	"fmt"
	"sort"
	"strings"
)

func Selector(input string) string {
	input = strings.ToLower(strings.TrimSpace(input))
	if len(input) < 10 || !strings.HasPrefix(input, "0x") {
		return ""
	}
	return input[:10]
}

func Classify(calls []Call) Classification {
	var out Classification
	out.Supported = true
	for _, call := range calls {
		selector := Selector(call.Input)
		if selector == "" {
			continue
		}
		if _, ok := TokenAndUtilitySelectors[selector]; ok {
			continue
		}
		info, ok := KnownSelectors[selector]
		if !ok {
			f := Finding{
				Depth:     call.Depth,
				To:        strings.ToLower(call.To),
				Selector:  selector,
				Label:     "unknown",
				MethodID:  0,
				Supported: false,
				Unknown:   true,
			}
			out.Calls = append(out.Calls, f)
			out.Unknowns = append(out.Unknowns, f)
			out.Unsupported = append(out.Unsupported, f)
			out.Supported = false
			continue
		}
		supported := SupportedMethodIDs[info.MethodID] && !RemovedOrUnsupportedMethodIDs[info.MethodID]
		f := Finding{
			Depth:     call.Depth,
			To:        strings.ToLower(call.To),
			Selector:  selector,
			Label:     info.Label,
			MethodID:  info.MethodID,
			Supported: supported,
		}
		out.Calls = append(out.Calls, f)
		if !supported {
			out.Unsupported = append(out.Unsupported, f)
			out.Supported = false
		}
	}
	return out
}

func (c Classification) MethodIDs() []int {
	seen := map[int]bool{}
	var ids []int
	for _, call := range c.Calls {
		if call.MethodID == 0 || seen[call.MethodID] {
			continue
		}
		seen[call.MethodID] = true
		ids = append(ids, call.MethodID)
	}
	sort.Ints(ids)
	return ids
}

func (c Classification) MethodIDString() string {
	ids := c.MethodIDs()
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, fmt.Sprintf("%d", id))
	}
	return strings.Join(parts, "-")
}
