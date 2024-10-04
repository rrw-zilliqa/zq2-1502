package main

import (
	"encoding/json"
	"fmt"
	"github.com/Zilliqa/gozilliqa-sdk/v3/account"
	"github.com/Zilliqa/gozilliqa-sdk/v3/keytools"
	"github.com/Zilliqa/gozilliqa-sdk/v3/provider"
	"github.com/Zilliqa/gozilliqa-sdk/v3/transaction"
	"github.com/Zilliqa/gozilliqa-sdk/v3/util"
	"os"
	"strconv"
)

func main() {
	arg := os.Args[1]
	chainId, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("Chain id is not an integer: %v\n", err)
		return
	}
	privateKey := "0000000000000000000000000000000000000000000000000000000000000002"
	fmt.Printf("Private Key: %s\n\n", privateKey)
	wallet := account.NewWallet()
	wallet.AddByPrivateKey(privateKey)
	fmt.Println("Wallet initialized; private key added.\n")
	publicKeyBytes := keytools.GetPublicKeyFromPrivateKey(util.DecodeHex(privateKey), true)
	publicKey := util.EncodeHex(publicKeyBytes)
	fmt.Printf("pubkey: %s\n\n", publicKey)

	address := keytools.GetAddressFromPublic(publicKeyBytes)
	fmt.Printf("Address: %s\n\n", address)

	// Init provider
	// (host is for z2 network node 0)
	host := arg
	provider := provider.NewProvider(host)
	fmt.Printf("Provider initialized with host: %s\n\n", host)

	// Get balance and nonce
	balAndNonce, err := provider.GetBalance(address)
	if err != nil {
		fmt.Printf("Error fetching balance and nonce: %v\n", err)
		return
	}
	fmt.Printf("Balance and Nonce: \n")
	prettyPrintJSON(balAndNonce)
	fmt.Println()

	// Get minimum gas price
	gasPrice, err := provider.GetMinimumGasPrice()
	if err != nil {
		fmt.Printf("Error fetching minimum gas price: %v\n", err)
		return
	}
	fmt.Printf("Minimum gas price: %s\n\n", gasPrice)

	// Txn params
	msgVersion := 1
	version := util.Pack(chainId, msgVersion)
	versionStr := strconv.FormatInt(int64(version), 10)
	fmt.Printf("Transaction Version (packed): %d\n", version)
	fmt.Printf("Transaction Version (string): %s\n\n", versionStr)

	tx := &transaction.Transaction{
		Version:      versionStr,
		Nonce:        strconv.FormatInt(balAndNonce.Nonce+1, 10),
		ToAddr:       "4BAF5faDA8e5Db92C3d3242618c5B47133AE003C",
		Amount:       "10000000",
		GasPrice:     gasPrice,
		GasLimit:     "50000",
		Code:         "",
		Data:         "",
		Priority:     false,
		SenderPubKey: publicKey,
	}
	fmt.Println("Transaction Object:")
	prettyPrintJSON(tx)
	fmt.Println()

	// Sign the transaction
	err = wallet.Sign(tx, *provider)
	if err != nil {
		fmt.Printf("Error signing transaction: %v\n", err)
		return
	}
	fmt.Println("Transaction signed successfully.\n")
	fmt.Println("Signed Transaction:")
	prettyPrintJSON(tx)
	fmt.Println()

	// Send the transaction
	rsp, err := provider.CreateTransaction(tx.ToTransactionPayload())
	if err != nil {
		fmt.Printf("Error creating transaction: %v\n", err)
		return
	}

	// Check for errors in the response
	if rsp.Error != nil {
		fmt.Printf("Transaction error: %v\n", rsp.Error.Message)
		return
	}
	fmt.Println("Transaction sent successfully")
	fmt.Println("Transaction Response:")
	prettyPrintJSON(rsp)
	fmt.Println()

	// Extract the transaction hash
	if rsp.Result == nil {
		fmt.Println("Error: transaction result is nil")
		return
	}
	resMap := rsp.Result.(map[string]interface{})
	hash, ok := resMap["TranID"].(string)
	if !ok {
		fmt.Println("Error: unable to retrieve txn hash")
		return
	}
	fmt.Printf("Transaction Hash: %s\n\n", hash)

	// Confirm the transaction
	tx.Confirm(hash, 1000, 3, provider)
	fmt.Printf("Transaction confirmation status: %d\n", tx.Status)
}

func prettyPrintJSON(v interface{}) {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}
