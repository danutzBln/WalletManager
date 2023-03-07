package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"WalletManager/handlers"
)

// var Wallets []Wallet

func generateUniqueWalletId() []byte {
	newUUID, err := exec.Command("uuidgen").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Generated UUID:")
	fmt.Printf("%s", newUUID)
	return newUUID
}

func main() {
	// Wallets = []Wallet{
	// 	Wallet{ID: string(generateUniqueWalletId()), Currencies: []Currency{Name: "USD"}},
	// }

	http.HandleFunc("/wallets", handlers.WalletsRouter)
	http.HandleFunc("/wallets/", handlers.WalletsRouter)
	http.HandleFunc("/", handlers.RootHandler)

	err := http.ListenAndServe("localhost:3000", nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

/*
### TODO ###
* Manage config
* Logging
* DynamoDB
* Auth?
* CRUD principle Create Read Update Delete (post, get, patch, delete)
* Currencies should be an array
TODO
When creating new wallet, add Total Amount in POST call. By default reserved is 0 and available is total amount
When creating new wallet, check if balance is present in currency. if not, return error or 0 total amount
Create new RESERVE endpoint. Give id and reserve amount as params, get the wallet, check if amounts match(total - reserved >= 0) and return 201 if yes or some error if not.error should have a message
*/
