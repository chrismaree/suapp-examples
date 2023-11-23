package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/suave/sdk"
	"github.com/flashbots/suapp-examples/framework"
)

func main() {
	fr := framework.New()
	// contract := fr.DeployContract("is-confidential.sol/IsConfidential.json")
	artifact, _ := fr.ReadArtifact("is-confidential.sol/IsConfidential.json")
	contractAddress := common.HexToAddress("0xFec1d78a8A6cFB445e6F254014fdE17d3EBE83cb")

	contract := sdk.GetContract(contractAddress, artifact.Abi, fr.Clt)

	tx, error := contract.SendTransaction("example", nil, nil)

	fmt.Printf("%#v\n", tx)

	if error != nil {
		fmt.Printf("%#v\n", error)
	}

}
