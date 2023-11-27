package framework

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/suave/sdk"
)

type Artifact struct {
	Abi *abi.ABI

	// Code is the code to deploy the contract
	Code []byte
}

func (f *Framework)  ReadArtifact(path string) (*Artifact, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("unable to get the current filename")
	}
	dirname := filepath.Dir(filename)

	data, err := os.ReadFile(filepath.Join(dirname, "../out", path))
	if err != nil {
		return nil, err
	}

	var artifact struct {
		Abi      *abi.ABI `json:"abi"`
		Bytecode struct {
			Object string `json:"object"`
		} `json:"bytecode"`
	}
	if err := json.Unmarshal(data, &artifact); err != nil {
		return nil, err
	}

	code, err := hex.DecodeString(artifact.Bytecode.Object[2:])
	if err != nil {
		return nil, err
	}

	art := &Artifact{
		Abi:  artifact.Abi,
		Code: code,
	}
	return art, nil
}

type PrivKey struct {
	Priv *ecdsa.PrivateKey
}

func (p *PrivKey) Address() common.Address {
	return crypto.PubkeyToAddress(p.Priv.PublicKey)
}

func (p *PrivKey) MarshalPrivKey() []byte {
	return crypto.FromECDSA(p.Priv)
}

func NewPrivKeyFromHex(hex string) *PrivKey {
	key, err := crypto.HexToECDSA(hex)
	if err != nil {
		panic(fmt.Sprintf("failed to parse private key: %v", err))
	}
	return &PrivKey{Priv: key}
}

func GeneratePrivKey() *PrivKey {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic(fmt.Sprintf("failed to generate private key: %v", err))
	}
	return &PrivKey{Priv: key}
}

type Contract struct {
	*sdk.Contract

	addr common.Address
	abi  *abi.ABI
	fr   *Framework
}

func (c *Contract) Call(methodName string) []interface{} {
	input, err := c.abi.Pack(methodName)
	if err != nil {
		panic(err)
	}

	callMsg := ethereum.CallMsg{
		To:   &c.addr,
		Data: input,
	}
	rpcClient := ethclient.NewClient(c.fr.rpc)
	output, err := rpcClient.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		panic(err)
	}

	results, err := c.abi.Methods[methodName].Outputs.Unpack(output)
	if err != nil {
		panic(err)
	}
	return results
}

func (c *Contract) SendTransaction(method string, args []interface{}, confidentialBytes []byte) *types.Receipt {
	txnResult, err := c.Contract.SendTransaction(method, args, confidentialBytes)
	if err != nil {
		panic(err)
	}
	receipt, err := txnResult.Wait()
	if err != nil {
		panic(err)
	}
	if receipt.Status == 0 {
		panic("bad")
	}
	return receipt
}

type Framework struct {
	config *Config
	rpc    *rpc.Client
	Clt    *sdk.Client
}

type Config struct {
	KettleRPC     string
	KettleAddr    common.Address
	FundedAccount *PrivKey
}

func DefaultConfig() *Config {
	isRigid := false

	// Check if "--rigid" is among the command-line arguments
	for _, arg := range os.Args[1:] {
		if arg == "--rigil" {
			isRigid = true
			break
		}
	}

	if isRigid {
		fmt.Printf("Using rigid mode\n")
		return &Config{
			KettleRPC:     "https://rpc.rigil.suave.flashbots.net",
			KettleAddr:    common.HexToAddress("03493869959c866713c33669ca118e774a30a0e5"),
			FundedAccount: NewPrivKeyFromHex("b2a626589787bac610d36b678f6c5878eb6ea39f7078df2f7560ce9ea5bd46ed"),
		}
	}

	fmt.Printf("Using localhost mode\n")
	return &Config{
		KettleRPC:     "http://localhost:8545",
		KettleAddr:    common.HexToAddress("b5feafbdd752ad52afb7e1bd2e40432a485bbb7f"),
		FundedAccount: NewPrivKeyFromHex("91ab9a7e53c220e6210460b65a7a3bb2ca181412a8a7b43ff336b3df1737ce12"),
	}
}

func New() *Framework {
	config := DefaultConfig()

	rpc, _ := rpc.Dial(config.KettleRPC)
	clt := sdk.NewClient(rpc, config.FundedAccount.Priv, config.KettleAddr)

	return &Framework{
		config: DefaultConfig(),
		rpc:    rpc,
		Clt:    clt,
	}
}

func (f *Framework) DeployContract(path string) *Contract {
	fmt.Printf("Deploying %s\n", path)
	artifact, err := f.ReadArtifact(path)
	if err != nil {
		panic(err)
	}

	// deploy contract
	txnResult, err := sdk.DeployContract(artifact.Code, f.Clt)
	if err != nil {
		panic(err)
	}

	receipt, err := txnResult.Wait()
	if err != nil {
		panic(err)
	}
	if receipt.Status == 0 {
		panic(fmt.Errorf("transaction failed"))
	}

	contract := sdk.GetContract(receipt.ContractAddress, artifact.Abi, f.Clt)
	return &Contract{addr: receipt.ContractAddress, fr: f, abi: artifact.Abi, Contract: contract}
}

func (c *Contract) Ref(acct *PrivKey) *Contract {
	cc := &Contract{
		addr:     c.addr,
		abi:      c.abi,
		fr:       c.fr,
		Contract: sdk.GetContract(c.addr, c.abi, c.fr.NewClient(acct)),
	}
	return cc
}

func (f *Framework) NewClient(acct *PrivKey) *sdk.Client {
	cc := DefaultConfig()
	rpc, _ := rpc.Dial(cc.KettleRPC)
	return sdk.NewClient(rpc, acct.Priv, cc.KettleAddr)
}

func (f *Framework) SignTx(priv *PrivKey, tx *types.LegacyTx) (*types.Transaction, error) {
	rpc, _ := rpc.Dial("https://rpc.rigil.suave.flashbots.net")

	cltAcct1 := sdk.NewClient(rpc, priv.Priv, common.Address{})
	signedTxn, err := cltAcct1.SignTxn(tx)
	if err != nil {
		return nil, err
	}
	return signedTxn, nil
}

var errFundAccount = fmt.Errorf("failed to fund account")

func (f *Framework) FundAccount(to common.Address, value *big.Int) error {
	txn := &types.LegacyTx{
		Value: value,
		To:    &to,
	}
	result, err := f.Clt.SendTransaction(txn)
	if err != nil {
		return err
	}
	_, err = result.Wait()
	if err != nil {
		return err
	}
	// check balance
	balance, err := f.Clt.RPC().BalanceAt(context.Background(), to, nil)
	if err != nil {
		return err
	}
	if balance.Cmp(value) != 0 {
		return errFundAccount
	}
	return nil
}
