package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/somatic-labs/meteorite/broadcast"
	"github.com/somatic-labs/meteorite/lib"
	"github.com/somatic-labs/meteorite/types"
)

const (
	BatchSize       = 100000000
	TimeoutDuration = 50 * time.Millisecond
)

func main() {
	config := types.Config{}
	if _, err := toml.DecodeFile("nodes.toml", &config); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	mnemonic, err := os.ReadFile("seedphrase")
	if err != nil {
		log.Fatalf("Failed to read seed phrase: %v", err)
	}
	privKey, pubKey, acctAddress := lib.GetPrivKey(config, mnemonic)

	nodes := lib.LoadNodes()
	if len(nodes) == 0 {
		log.Fatal("No nodes available to send transactions")
	}
	nodeURL := nodes[0] // Use only the first node

	if nodeURL == "" {
		log.Fatal("Node URL is empty. Please verify the nodes configuration.")
	}
	chainID, err := lib.GetChainID(nodeURL)
	if err != nil {
		log.Fatalf("Failed to get chain ID: %v", err)
	}

	msgParams := config.MsgParams

	// Get the account info
	_, accNum := lib.GetAccountInfo(acctAddress, config)
	if err != nil {
		log.Fatalf("Failed to get account info: %v", err)
	}

	sequence := uint64(1) // Start from sequence number 1

	// Create a TransactionParams struct
	txParams := types.TransactionParams{
		Config:      config,
		NodeURL:     nodeURL,
		ChainID:     chainID,
		Sequence:    sequence,
		AccNum:      accNum,
		PrivKey:     privKey,
		PubKey:      pubKey,
		AcctAddress: acctAddress,
		MsgType:     config.MsgType,
		MsgParams:   msgParams,
	}

	// Call the broadcast loop
	successfulTxns, failedTxns, responseCodes, _ := broadcastLoop(txParams, BatchSize)

	// After the loop
	fmt.Println("Successful transactions:", successfulTxns)
	fmt.Println("Failed transactions:", failedTxns)
	totalTxns := successfulTxns + failedTxns
	fmt.Println("Response code breakdown:")
	for code, count := range responseCodes {
		percentage := float64(count) / float64(totalTxns) * 100
		fmt.Printf("Code %d: %d (%.2f%%)\n", code, count, percentage)
	}
}

// broadcastLoop handles the main transaction broadcasting logic
func broadcastLoop(
	txParams types.TransactionParams,
	batchSize int,
) (successfulTxns, failedTxns int, responseCodes map[uint32]int, updatedSequence uint64) {
	successfulTxns = 0
	failedTxns = 0
	responseCodes = make(map[uint32]int)
	sequence := txParams.Sequence

	for i := 0; i < batchSize; i++ {
		currentSequence := sequence

		fmt.Println("FROM LOOP, currentSequence", currentSequence)
		fmt.Println("FROM LOOP, accNum", txParams.AccNum)
		fmt.Println("FROM LOOP, chainID", txParams.ChainID)

		start := time.Now()
		resp, _, err := broadcast.SendTransactionViaRPC(
			txParams,
			currentSequence,
		)
		elapsed := time.Since(start)

		fmt.Println("FROM MAIN, err", err)
		fmt.Println("FROM MAIN, resp", resp.Code)

		if err == nil {
			fmt.Printf("%s Transaction succeeded, sequence: %d, time: %v\n",
				time.Now().Format("15:04:05"), currentSequence, elapsed)
			successfulTxns++
			responseCodes[resp.Code]++
			sequence++ // Increment sequence for next transaction
			continue
		}

		fmt.Printf("%s Error: %v\n", time.Now().Format("15:04:05.000"), err)
		fmt.Println("FROM MAIN, resp.Code", resp.Code)

		if resp.Code == 32 {
			// Extract the expected sequence number from the error message
			expectedSeq, parseErr := extractExpectedSequence(err.Error())
			if parseErr != nil {
				fmt.Printf("%s Failed to parse expected sequence: %v\n", time.Now().Format("15:04:05.000"), parseErr)
				failedTxns++
				continue
			}

			sequence = expectedSeq
			fmt.Printf("%s Set sequence to expected value %d due to mismatch\n",
				time.Now().Format("15:04:05"), sequence)

			// Re-send the transaction with the correct sequence
			start = time.Now()
			resp, _, err = broadcast.SendTransactionViaRPC(
				txParams,
				sequence,
			)
			elapsed = time.Since(start)

			if err != nil {
				fmt.Printf("%s Error after adjusting sequence: %v\n", time.Now().Format("15:04:05.000"), err)
				failedTxns++
				continue
			}

			fmt.Printf("%s Transaction succeeded after adjusting sequence, sequence: %d, time: %v\n",
				time.Now().Format("15:04:05"), sequence, elapsed)
			successfulTxns++
			responseCodes[resp.Code]++
			sequence++ // Increment sequence for next transaction
			continue
		}
		failedTxns++

	}
	updatedSequence = sequence
	return successfulTxns, failedTxns, responseCodes, updatedSequence
}

// Function to extract the expected sequence number from the error message
func extractExpectedSequence(errMsg string) (uint64, error) {
	// Parse the error message to extract the expected sequence number
	// Example error message:
	// "account sequence mismatch, expected 42, got 41: incorrect account sequence"
	index := strings.Index(errMsg, "expected ")
	if index == -1 {
		return 0, errors.New("expected sequence not found in error message")
	}

	start := index + len("expected ")
	rest := errMsg[start:]
	parts := strings.SplitN(rest, ",", 2)
	if len(parts) < 1 {
		return 0, errors.New("failed to split expected sequence from error message")
	}

	expectedSeqStr := strings.TrimSpace(parts[0])
	expectedSeq, err := strconv.ParseUint(expectedSeqStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse expected sequence number: %v", err)
	}

	return expectedSeq, nil
}
