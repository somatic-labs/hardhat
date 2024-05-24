package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
)

func getInitialSequence(address string, config Config) (int64, int64) {
	resp, err := httpGet(config.Nodes.API + "/cosmos/auth/v1beta1/accounts/" + address)
	if err != nil {
		log.Printf("Failed to get initial sequence: %v", err)
		return 0, 0
	}

	var accountRes AccountResult
	err = json.Unmarshal(resp, &accountRes)
	if err != nil {
		log.Printf("Failed to unmarshal account result: %v", err)
		return 0, 0
	}

	seqint, err := strconv.ParseInt(accountRes.Account.Sequence, 10, 64)
	if err != nil {
		log.Printf("Failed to convert sequence to int: %v", err)
		return 0, 0
	}

	accnum, err := strconv.ParseInt(accountRes.Account.AccountNumber, 10, 64)
	if err != nil {
		log.Printf("Failed to convert account number to int: %v", err)
		return 0, 0
	}

	return seqint, accnum
}

func getChainID(nodeURL string) (string, error) {
	resp, err := httpGet(nodeURL + "/status")
	if err != nil {
		log.Printf("Failed to get node status: %v", err)
		return "", err
	}

	var statusRes NodeStatusResponse
	err = json.Unmarshal(resp, &statusRes)
	if err != nil {
		log.Printf("Failed to unmarshal node status result: %v", err)
		return "", err
	}

	return statusRes.Result.NodeInfo.Network, nil
}

func httpGet(url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		netErr, ok := err.(net.Error)
		if ok && netErr.Timeout() {
			log.Printf("Request to %s timed out, continuing...", url)
			return nil, nil
		}
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// This function will load our nodes from nodes.toml.
func loadNodes() []string {
	var config Config
	if _, err := toml.DecodeFile("nodes.toml", &config); err != nil {
		log.Fatalf("Failed to load nodes.toml: %v", err)
	}
	return config.Nodes.RPC
}
