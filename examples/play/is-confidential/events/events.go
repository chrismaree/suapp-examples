package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Artifact struct {
	Abi *abi.ABI

	// Code is the code to deploy the contract
	Code []byte
}

func main() {
	log.Println("Starting up...")

	// Connect to the Ethereum client
	client, err := ethclient.Dial("https://rpc.rigil.suave.flashbots.net")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	log.Println("Connected to Ethereum client")

	// Specify the contract address
	contractAddress := common.HexToAddress("0xFec1d78a8A6cFB445e6F254014fdE17d3EBE83cb")
	log.Printf("Monitoring contract at address: %s\n", contractAddress.Hex())

	// Load your contract's ABI
	abiJSON, err := os.ReadFile("../../../../out/is-confidential.sol/IsConfidential.json")
	if err != nil {
		log.Fatalf("Failed to read ABI: %v", err)
	}

	var artifact struct {
		Abi      *abi.ABI `json:"abi"`
		Bytecode struct {
			Object string `json:"object"`
		} `json:"bytecode"`
	}
	if err := json.Unmarshal(abiJSON, &artifact); err != nil {
		log.Fatalf("Failed to read ABI: %v", err)
	}

	code, err := hex.DecodeString(artifact.Bytecode.Object[2:])
	if err != nil {
		log.Fatalf("Failed to read ABI: %v", err)
	}

	contractABI := &Artifact{
		Abi:  artifact.Abi,
		Code: code,
	}

	// Starting block number for querying events
	startBlock := big.NewInt(0) // Adjust this as needed

	// Polling interval
	pollInterval := time.Second * 10 // Adjust the polling interval as needed

	// Polling loop
	for {
		// Get the latest block number
		latestBlock, err := client.BlockNumber(context.Background())
		if err != nil {
			log.Fatalf("Failed to get latest block number: %v", err)
		}

		// Define a query to filter logs
		query := ethereum.FilterQuery{
			FromBlock: startBlock,
			ToBlock:   big.NewInt(int64(latestBlock)),
			Addresses: []common.Address{contractAddress},
		}

		// Query the logs
		logs, err := client.FilterLogs(context.Background(), query)
		if err != nil {
			log.Fatalf("Failed to query logs: %v", err)
		}

		// Process each log
		for _, vLog := range logs {
			fmt.Println("Raw Log:", vLog)

			// Parse each log with the contract ABI
			event, err := contractABI.Abi.Unpack("InFunction", vLog.Data)
			if err != nil {
				log.Println("Failed to unpack log data:", err)
				continue
			}

			fmt.Printf("Decoded event data: %v\n", event)
		}

		// Update the starting block for the next query
		startBlock = big.NewInt(int64(latestBlock) + 1)

		// Wait for the next poll
		time.Sleep(pollInterval)
	}
}
