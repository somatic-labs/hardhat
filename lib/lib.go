package lib

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/somatic-labs/hardhat/types"
)

var client = &http.Client{
	Timeout: 10 * time.Second, // Adjusted timeout to 10 seconds
	Transport: &http.Transport{
		MaxIdleConns:        100,              // Increased maximum idle connections
		MaxIdleConnsPerHost: 10,               // Increased maximum idle connections per host
		IdleConnTimeout:     90 * time.Second, // Increased idle connection timeout
		TLSHandshakeTimeout: 10 * time.Second, // Increased TLS handshake timeout
	},
}

func GetInitialSequence(address string, config types.Config) (int64, int64) {
	resp, err := HTTPGet(config.Nodes.API + "/cosmos/auth/v1beta1/accounts/" + address)
	if err != nil {
		log.Printf("Failed to get initial sequence: %v", err)
		return 0, 0
	}

	var accountRes types.AccountResult
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

func GetChainID(nodeURL string) (string, error) {
	resp, err := HTTPGet(nodeURL + "/status")
	if err != nil {
		log.Printf("Failed to get node status: %v", err)
		return "", err
	}

	var statusRes types.NodeStatusResponse
	err = json.Unmarshal(resp, &statusRes)
	if err != nil {
		log.Printf("Failed to unmarshal node status result: %v", err)
		return "", err
	}

	return statusRes.Result.NodeInfo.Network, nil
}

func HTTPGet(url string) ([]byte, error) {
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
func LoadNodes() []string {
	var config types.Config
	if _, err := toml.DecodeFile("nodes.toml", &config); err != nil {
		log.Fatalf("Failed to load nodes.toml: %v", err)
	}
	return config.Nodes.RPC
}

func GenerateRandomString(config types.Config) (string, error) {
	// Generate a random size between config.RandMin and config.RandMax
	sizeB, err := rand.Int(rand.Reader, big.NewInt(config.RandMax-config.RandMin+1))
	if err != nil {
		return "", err
	}
	sizeB = sizeB.Add(sizeB, big.NewInt(config.RandMin))

	// Calculate the number of bytes to generate (2 characters per byte in hex encoding)
	nBytes := int(sizeB.Int64()) / 2

	randomBytes := make([]byte, nBytes)
	_, err = rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(randomBytes), nil
}

func GenerateRandomStringOfLength(n int) (string, error) {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		b[i] = letters[num.Int64()]
	}
	return string(b), nil
}

func GenerateRandomAccount(prefix string) (sdk.AccAddress, error) {
	// Generate 20 random bytes
	randomBytes := make([]byte, 20)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// Create an AccAddress from the random bytes
	accAddress := sdk.AccAddress(randomBytes)

	return accAddress, nil
}
