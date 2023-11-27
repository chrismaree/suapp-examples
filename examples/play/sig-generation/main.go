package main

import (
	"encoding/hex"
	"fmt"
	"github.com/flashbots/suapp-examples/framework"
)

func main() {
	fr := framework.New()

	// Deploy the contract
	fmt.Println("Deploying the contract...")
	contract := fr.DeployContract("ofa-signature.sol/ConfidentialKeyStore.json")

	// Call the createTransaction function and log the output
	fmt.Println("Calling the createTransaction function...")
	txOutput := contract.Call("createTransaction")
	fmt.Printf("createTransaction output: %s\n", hex.EncodeToString(txOutput[0].([]byte)))

	// Set the private key
	fmt.Println("Setting the private key...")
	// privateKey := "0x964f7e4657e386b633a056ebbdd060d45cf64b6745e65f0ccef34f072da45e61"
	// contract.SendTransaction("setPrivateKey", []interface{}{privateKey}, []byte{})

	// Call the signEthTransaction function and log the output
	fmt.Println("Calling the signEthTransaction function...")
	signedTxOutput := contract.Call("signEthTransaction")
	fmt.Printf("signEthTransaction output: %s\n", hex.EncodeToString(signedTxOutput[0].([]byte)))
}
