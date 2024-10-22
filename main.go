package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/somatic-labs/hardhat/broadcast"
	"github.com/somatic-labs/hardhat/lib"
	"github.com/somatic-labs/hardhat/types"
)

const (
	BatchSize       = 100000000
	MaxRetries      = 1
	TimeoutDuration = 2 * time.Second
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

	successfulTxns, failedTxns := 0, 0
	responseCodes := make(map[uint32]int)

	initialSequence, accNum := lib.GetInitialSequence(acctAddress, config)
	sequence := initialSequence

	for i := 0; i < BatchSize; i++ {
		currentSequence := sequence
		sequence++

		start := time.Now()
		resp, _, err := sendTransactionWithRetry(
			config,
			nodeURL,
			chainID,
			uint64(currentSequence),
			uint64(accNum),
			privKey, // Remove .(cryptotypes.PrivKey)
			pubKey,  // Remove .(cryptotypes.PubKey)
			acctAddress,
			config.MsgType,
			msgParams,
		)
		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf("%s Error: %v\n", time.Now().Format("15:04:05"), err)

			if strings.Contains(err.Error(), "account sequence mismatch") {
				parts := strings.Split(err.Error(), "expected ")
				if len(parts) > 1 {
					expectedSeqParts := strings.Split(parts[1], ",")
					if len(expectedSeqParts) > 0 {
						expectedSeq, parseErr := strconv.ParseInt(expectedSeqParts[0], 10, 64)
						if parseErr == nil {
							sequence = expectedSeq
							fmt.Printf("%s Set sequence to expected value %d due to mismatch\n",
								time.Now().Format("15:04:05"), sequence)

							// Re-send the transaction with the correct sequence
							currentSequence = sequence
							sequence++
							resp, _, err := sendTransactionWithRetry(
								config,
								nodeURL,
								chainID,
								uint64(currentSequence),
								uint64(accNum),
								privKey, // Remove .(cryptotypes.PrivKey)
								pubKey,  // Remove .(cryptotypes.PubKey)
								acctAddress,
								config.MsgType,
								msgParams,
							)
							elapsed = time.Since(start)

							if err != nil {
								fmt.Printf("%s Error after adjusting sequence: %v\n", time.Now().Format("15:04:05"), err)
								failedTxns++
							} else {
								fmt.Printf("%s Transaction succeeded after adjusting sequence, sequence: %d, time: %v\n",
									time.Now().Format("15:04:05"), currentSequence, elapsed)
								successfulTxns++
								responseCodes[resp.Code]++
							}

							// Continue to the next iteration
							continue
						}
					}
				}
			}

			failedTxns++
		} else {
			fmt.Printf("%s Transaction succeeded, sequence: %d, time: %v\n",
				time.Now().Format("15:04:05"), currentSequence, elapsed)
			successfulTxns++
			responseCodes[resp.Code]++
		}
	}

	fmt.Println("Successful transactions:", successfulTxns)
	fmt.Println("Failed transactions:", failedTxns)
	totalTxns := successfulTxns + failedTxns
	fmt.Println("Response code breakdown:")
	for code, count := range responseCodes {
		percentage := float64(count) / float64(totalTxns) * 100
		fmt.Printf("Code %d: %d (%.2f%%)\n", code, count, percentage)
	}
}

func sendTransactionWithRetry(config types.Config, nodeURL, chainID string, sequence, accNum uint64, privKey cryptotypes.PrivKey, pubKey cryptotypes.PubKey, acctAddress, msgType string, msgParams types.MsgParams) (*coretypes.ResultBroadcastTx, string, error) {
	var lastErr error
	startTime := time.Now()
	for retry := 0; retry < MaxRetries; retry++ {
		attemptStart := time.Now()

		// Create a context with a timeout for each attempt
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		respChan := make(chan *coretypes.ResultBroadcastTx)
		errChan := make(chan error)

		go func() {
			resp, _, err := broadcast.SendTransactionViaRPC(
				config,
				nodeURL,
				chainID,
				sequence,
				accNum,
				privKey,
				pubKey,
				acctAddress,
				msgType,
				msgParams,
			)
			if err != nil {
				errChan <- err
			} else {
				respChan <- resp
			}
		}()

		select {
		case resp := <-respChan:
			return resp, "", nil
		case err := <-errChan:
			lastErr = err
		case <-ctx.Done():
			lastErr = ctx.Err()
		}

		attemptDuration := time.Since(attemptStart)
		fmt.Printf("%s Retry %d failed after %v: %v\n", time.Now().Format("15:04:05"), retry, attemptDuration, lastErr)

		if time.Since(startTime) > 2*time.Second {
			return nil, "", fmt.Errorf("total retry time exceeded 1 second")
		}

		time.Sleep(TimeoutDuration)
	}

	totalDuration := time.Since(startTime)
	return nil, "", fmt.Errorf("failed after %d retries in %v: %v", MaxRetries, totalDuration, lastErr)
}
