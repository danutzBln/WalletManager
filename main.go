package main

import (
	"fmt"
	"net/http"
	"os"

	"WalletManager/handlers"
)

func main() {
	http.HandleFunc("/wallets", handlers.WalletsRouter)
	http.HandleFunc("/wallets/", handlers.WalletsRouter)
	http.HandleFunc("/wallets/reserve", handlers.WalletsRouter)
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
*/
