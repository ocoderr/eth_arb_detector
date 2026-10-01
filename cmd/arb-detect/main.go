package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	detector "eth-arb-detector"
)

type input struct {
	Block       *detector.Block   `json:"block"`
	Transaction *detector.BlockTx `json:"transaction"`
	Receipt     *detector.Receipt `json:"receipt"`
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("arb-detect", flag.ContinueOnError)
	flags.SetOutput(stderr)
	path := flags.String("input", "-", "local JSON file or - for stdin; no RPC")
	minimum := flags.Int("min-confidence", 50, "heuristic threshold, 1..100 (not a probability)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected positional arguments")
	}
	if *minimum < 1 || *minimum > 100 {
		return errors.New("min-confidence must be 1..100")
	}
	reader := stdin
	if *path != "-" {
		file, err := os.Open(*path)
		if err != nil {
			return err
		}
		defer file.Close()
		reader = file
	}
	const maxInput = 16 << 20
	payload, err := io.ReadAll(io.LimitReader(reader, maxInput+1))
	if err != nil {
		return err
	}
	if len(payload) > maxInput {
		return errors.New("input exceeds 16 MiB")
	}
	var data input
	if err := json.Unmarshal(payload, &data); err != nil {
		return err
	}
	if data.Block == nil || data.Transaction == nil || data.Receipt == nil {
		return errors.New("block, transaction and receipt objects are required")
	}
	if data.Transaction.Hash == "" || !strings.EqualFold(data.Transaction.Hash, data.Receipt.TransactionHash) {
		return errors.New("transaction/receipt hashes must be present and match")
	}
	if data.Block.Number == "" || data.Receipt.BlockNumber == "" ||
		detector.HexOrDecimalInt(data.Block.Number) <= 0 ||
		detector.HexOrDecimalInt(data.Block.Number) != detector.HexOrDecimalInt(data.Receipt.BlockNumber) {
		return errors.New("block/receipt block numbers must be present and match")
	}
	if data.Receipt.Status != "0x0" && data.Receipt.Status != "0x1" {
		return errors.New("receipt status must be 0x0 or 0x1")
	}
	candidate := detector.IdentifyOnlineArb(*data.Block, *data.Transaction, *data.Receipt)
	report := struct {
		SchemaVersion     int                         `json:"schemaVersion"`
		IsCandidate       bool                        `json:"isCandidate"`
		ExecutionVerified bool                        `json:"executionVerified"`
		Warning           string                      `json:"warning"`
		Candidate         detector.OnlineArbCandidate `json:"candidate"`
	}{1, candidate.IsCandidate(*minimum), false, "Heuristic on supplied Ethereum data, not proof of arbitrage, net profit, protocol support or signal competition.", candidate}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
