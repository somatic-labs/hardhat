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
	BatchSize  = 100
	MaxWorkers = 10000
)

func main() {
	// Load the config
	config := Config{}
	if _, err := toml.DecodeFile("nodes.toml", &config); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create a map to store the sequence number for each node
	sequenceMap := make(map[string]int64)
	var sequenceMu sync.Mutex // Mutex to protect the sequenceMap

	// Tracking variables
	var successfulTxns int
	var failedTxns int
	var mu sync.Mutex
	// Declare a map to hold response codes and their counts
	responseCodes := make(map[uint32]int)

	// Read seed phrase
	mnemonic, err := os.ReadFile("seedphrase")
	if err != nil {
		log.Fatalf("Failed to read seed phrase: %v", err)
	}
	privkey, pubKey, acctaddress := getPrivKey(config, mnemonic)

	// Load nodes from config
	successfulNodes := loadNodes()
	fmt.Printf("Number of nodes: %d\n", len(successfulNodes))

	// Get the correct chain ID
	chainID, err := getChainID(successfulNodes[0])
	if err != nil {
		log.Fatalf("Failed to get chain ID: %v", err)
	}

	var wg sync.WaitGroup

	// Compile the regex outside the loop
	reMismatch := regexp.MustCompile("account sequence mismatch")
	reExpected := regexp.MustCompile(`expected (\d+)`)

	// Get the account number (accNum) once
	_, accNum := getInitialSequence(acctaddress, config)

	// Build msgParams map
	msgParams := map[string]interface{}{
		"amount":   config.MsgParams.Amount,
		"receiver": config.MsgParams.Receiver,
		// Add other parameters from MsgParams as needed
	}

	transactionCh := make(chan string, BatchSize) // Create a buffered channel for transactions

	for _, nodeURL := range successfulNodes {
		wg.Add(1)
		go func(nodeURL string) {
			defer wg.Done()

			// Initialize the sequence number for the node to zero
			sequenceMu.Lock()
			sequenceMap[nodeURL] = 0
			sequenceMu.Unlock()

			for {
				select {
				case <-transactionCh:
					sequenceMu.Lock()
					currentSequence := sequenceMap[nodeURL]
					sequenceMu.Unlock()

					resp, _, err := sendTransactionViaRPC(
						config,
						nodeURL,
						chainID,
						uint64(currentSequence),
						uint64(accNum),
						privkey,
						pubKey,
						acctaddress,
						config.MsgType,
						msgParams,
					)
					if err != nil {
						mu.Lock()
						failedTxns++
						mu.Unlock()
						fmt.Printf("%s Node: %s, Error: %v\n", time.Now().Format("15:04:05"), nodeURL, err)
					} else {
						mu.Lock()
						successfulTxns++
						mu.Unlock()
						if resp != nil {
							// Increment the count for this response code
							mu.Lock()
							responseCodes[resp.Code]++
							mu.Unlock()
						}

						match := reMismatch.MatchString(resp.Log)
						if match {
							matches := reExpected.FindStringSubmatch(resp.Log)
							if len(matches) > 1 {
								newSequence, err := strconv.ParseInt(matches[1], 10, 64)
								if err != nil {
									log.Fatalf("Failed to convert sequence to integer: %v", err)
								}
								// Update the per-node sequence to the expected value
								sequenceMu.Lock()
								sequenceMap[nodeURL] = newSequence
								sequenceMu.Unlock()
								fmt.Printf("%s Node: %s, we had an account sequence mismatch, adjusting to %d\n",
									time.Now().Format("15:04:05"), nodeURL, newSequence)
							}
						} else {
							// Increment the per-node sequence number if there was no mismatch
							sequenceMu.Lock()
							sequenceMap[nodeURL]++
							sequenceMu.Unlock()
							fmt.Printf("%s Node: %s, sequence: %d\n", time.Now().Format("15:04:05"), nodeURL, sequenceMap[nodeURL])
						}
					}
				}
			}
		}(nodeURL)
	}

	// Send transactions to the worker goroutines
	for i := 0; i < len(successfulNodes)*BatchSize; i++ {
		transactionCh <- fmt.Sprintf("Transaction %d", i)
	}

	close(transactionCh) // Close the transaction channel when all transactions are sent

	wg.Wait()

	fmt.Println("Successful transactions: ", successfulTxns)
	fmt.Println("Failed transactions: ", failedTxns)
	totalTxns := successfulTxns + failedTxns
	fmt.Println("Response code breakdown:")
	for code, count := range responseCodes {
		percentage := float64(count) / float64(totalTxns) * 100
		fmt.Printf("Code %d: %d (%.2f%%)\n", code, count, percentage)
	}
}
