package detector

import "testing"

func TestClassifySupportedAndUnsupported(t *testing.T) {
	classification := Classify([]Call{
		{Depth: 1, To: "0xpool", Input: "0x022c0d9f00000000"},
		{Depth: 1, To: "0xreactor", Input: "0x3f62192e00000000"},
		{Depth: 1, To: "0xtoken", Input: "0xa9059cbb00000000"},
	})

	if classification.Supported {
		t.Fatal("classification should be unsupported when a removed protocol is present")
	}
	if got := classification.MethodIDString(); got != "2-45" {
		t.Fatalf("method ids = %q, want 2-45", got)
	}
	if len(classification.Calls) != 2 {
		t.Fatalf("calls = %d, want 2", len(classification.Calls))
	}
	if len(classification.Unsupported) != 1 {
		t.Fatalf("unsupported = %d, want 1", len(classification.Unsupported))
	}
	if classification.Unsupported[0].MethodID != 45 {
		t.Fatalf("unsupported method = %d, want 45", classification.Unsupported[0].MethodID)
	}
}

func TestClassifyUnknownSelector(t *testing.T) {
	classification := Classify([]Call{
		{Depth: 1, To: "0xunknown", Input: "0xdeadbeef00000000"},
	})

	if classification.Supported {
		t.Fatal("unknown selector should make classification unsupported")
	}
	if len(classification.Unknowns) != 1 {
		t.Fatalf("unknowns = %d, want 1", len(classification.Unknowns))
	}
	if classification.Unknowns[0].Selector != "0xdeadbeef" {
		t.Fatalf("selector = %q, want 0xdeadbeef", classification.Unknowns[0].Selector)
	}
}
