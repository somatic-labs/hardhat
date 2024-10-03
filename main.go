package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	BatchSize = 100000000 // Increase as needed for load testing
)

func main() {
	// Load the config
	config := Config{}
	if _, err := toml.DecodeFile("nodes.toml", &config); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Read seed phrase
	mnemonic, err := os.ReadFile("seedphrase")
	if err != nil {
		log.Fatalf("Failed to read seed phrase: %v", err)
	}
	privKey, pubKey, acctAddress := getPrivKey(config, mnemonic)

	// Load nodes from config
	nodes := loadNodes()
	if len(nodes) == 0 {
		log.Fatal("No nodes available to send transactions")
	}
	fmt.Printf("Number of nodes: %d\n", len(nodes))

	// Get the correct chain ID
	chainID, err := getChainID(nodes[0])
	if err != nil {
		log.Fatalf("Failed to get chain ID: %v", err)
	}

	// Compile regex patterns for error messages
	reMismatch := regexp.MustCompile(`account sequence mismatch, expected (\d+), got (\d+): incorrect account sequence`)

	// Build msgParams map
	msgParams := map[string]interface{}{
		"amount":     config.MsgParams.Amount,
		"to_address": config.MsgParams.ToAddress,
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var successfulTxns, failedTxns int
	responseCodes := make(map[uint32]int)

	// Channel to control overall concurrency if needed
	// Here, we limit the total number of concurrent transactions
	txChan := make(chan struct{}, len(nodes)) // One slot per node

	// Function to send transactions per node
	for _, nodeURL := range nodes {
		wg.Add(1)
		go func(nodeURL string) {
			defer wg.Done()

			// Fetch the initial account number and sequence from this node
			initialSequence, accNum := getInitialSequence(acctAddress, config)
			sequence := initialSequence

			for i := 0; i < BatchSize; i++ {
				txChan <- struct{}{} // Acquire a slot

				currentSequence := sequence
				sequence++

				// Send the transaction
				resp, _, err := sendTransactionViaRPC(
					config,
					nodeURL,
					chainID,
					uint64(currentSequence),
					uint64(accNum),
					privKey,
					pubKey,
					acctAddress,
					config.MsgType,
					msgParams,
				)

				if err != nil {
					fmt.Printf("%s Node: %s, Error: %v\n", time.Now().Format("15:04:05"), nodeURL, err)

					// Adjust sequence on specific errors
					switch {
					case reMismatch.MatchString(err.Error()):
						// Extract the expected sequence number from the error message
						matches := reMismatch.FindStringSubmatch(err.Error())
						if len(matches) >= 2 {
							expectedSeq, parseErr := strconv.ParseUint(matches[1], 10, 64)
							if parseErr == nil {
								sequence = int64(expectedSeq)
								fmt.Printf("%s Node: %s, Set sequence to expected value %d due to mismatch\n",
									time.Now().Format("15:04:05"), nodeURL, sequence)
							} else {
								// Handle parsing error
								sequence--
							}
						} else {
							// If regex did not capture the expected number, decrement the sequence
							sequence--
						}
					default:
						// For other errors, you may choose to log or handle differently
						sequence++
					}

					mu.Lock()
					failedTxns++
					mu.Unlock()
				} else {
					fmt.Printf("%s Node: %s, Transaction succeeded, sequence: %d\n",
						time.Now().Format("15:04:05"), nodeURL, currentSequence)

					mu.Lock()
					successfulTxns++
					responseCodes[resp.Code]++
					mu.Unlock()
				}

				<-txChan // Release the slot
			}
		}(nodeURL)
	}

	wg.Wait()

	fmt.Println("Successful transactions:", successfulTxns)
	fmt.Println("Failed transactions:", failedTxns)
	totalTxns := successfulTxns + failedTxns
	fmt.Println("Response code breakdown:")
	for code, count := range responseCodes {
		percentage := float64(count) / float64(totalTxns) * 100
		fmt.Printf("Code %d: %d (%.2f%%)\n", code, count, percentage)
	}
}
