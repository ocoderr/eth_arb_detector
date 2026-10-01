package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

func TestSyntheticCycle(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"--input", "../../testdata/synthetic-cycle.json"}, strings.NewReader(""), &out, io.Discard); err != nil {
		t.Fatal(err)
	}
	var report struct {
		IsCandidate       bool
		ExecutionVerified bool
		Candidate         struct{ Kind string }
	}
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if !report.IsCandidate || report.ExecutionVerified || report.Candidate.Kind != "transfer_cycle_candidate" {
		t.Fatalf("unexpected result: %+v", report)
	}
}

func TestInvalidInput(t *testing.T) {
	for _, data := range []string{"{}", "null", "not json", "{} {}"} {
		if err := run(nil, strings.NewReader(data), io.Discard, io.Discard); err == nil {
			t.Fatal("invalid input accepted")
		}
	}
	if err := run([]string{"--min-confidence", "0"}, strings.NewReader("{}"), io.Discard, io.Discard); err == nil {
		t.Fatal("invalid threshold")
	}
}

func TestReceiptConsistencyAndFailedTransaction(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/synthetic-cycle.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"hash", "block", "status", "failed", "no-logs"} {
		var data input
		if err := json.Unmarshal(raw, &data); err != nil {
			t.Fatal(err)
		}
		switch mode {
		case "hash":
			data.Receipt.TransactionHash = "other"
		case "block":
			data.Receipt.BlockNumber = "0x2"
		case "status":
			data.Receipt.Status = ""
		case "failed":
			data.Receipt.Status = "0x0"
		case "no-logs":
			data.Receipt.Logs = nil
		}
		payload, err := json.Marshal(data)
		if err != nil {
			t.Fatal(err)
		}
		var out bytes.Buffer
		err = run(nil, bytes.NewReader(payload), &out, io.Discard)
		if mode == "failed" || mode == "no-logs" {
			if err != nil {
				t.Fatal(err)
			}
			var report struct{ IsCandidate bool }
			if err := json.Unmarshal(out.Bytes(), &report); err != nil {
				t.Fatal(err)
			}
			if report.IsCandidate {
				t.Fatal("negative case classified as candidate")
			}
		} else if err == nil {
			t.Fatal("inconsistent input accepted")
		}
	}
}
