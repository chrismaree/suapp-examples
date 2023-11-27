package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/flashbots/suapp-examples/framework"
)

func main() {
	fr := framework.New()
	contract := fr.DeployContract("store-key.sol/StoreKey.json")
	// artifact, _ := fr.ReadArtifact("is-confidential.sol/IsConfidential.json")
	// contract := sdk.GetContract(contractAddress, artifact.Abi, fr.Clt)

	fmt.Println("Deployed! Contract address: ", contract.Address())

	privateKey := []byte("yourPrivateKeyHere")
	// privateKeyJson, _ := json.Marshal(privateKey)
	tx := contract.SendTransaction("storePrivateKey", []interface{}{}, privateKey)

	hintEvent := &HintEvent{}
	if err := hintEvent.Unpack(tx.Logs[0]); err != nil {
		panic(err)
	}

	fmt.Println("Hint event id", hintEvent.BidId)
	hintEventHintString := string(hintEvent.Hint)
	fmt.Println("Hint event hint (string)", hintEventHintString)

	// Call the emitStoredEvent function from the contract with the bidId
	txStoredEvent := contract.SendTransaction("emitStoredEvent", []interface{}{hintEvent.BidId}, nil)

	// Unpack the logs from the transaction
	storedEvent := &HintEvent{}
	if err := storedEvent.Unpack(txStoredEvent.Logs[0]); err != nil {
		panic(err)
	}

	// Print the stored event details
	fmt.Println("Stored event id", storedEvent.BidId)
	storedEventHintString := string(storedEvent.Hint)
	fmt.Println("Stored event hint (string)", storedEventHintString)

}

type HintEvent struct {
	BidId [16]byte
	Hint  []byte
}

func (h *HintEvent) Unpack(log *types.Log) error {
	unpacked, err := hintEventABI.Inputs.Unpack(log.Data)
	if err != nil {
		return err
	}
	h.BidId = unpacked[0].([16]byte)
	h.Hint = unpacked[1].([]byte)
	return nil
}

var hintEventABI abi.Event

func init() {
	fr := framework.New()
	artifact, _ := fr.ReadArtifact("store-key.sol/StoreKey.json")
	hintEventABI = artifact.Abi.Events["HintEvent"]
}
