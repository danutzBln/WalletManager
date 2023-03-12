package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"WalletManager/wallet"

	"github.com/asdine/storm"
	"gopkg.in/mgo.v2/bson"
)

// Encode a request body to a wallet struct
func bodyToWallet(r *http.Request, wallet *wallet.Wallet) error {
	if r == nil {
		return errors.New("a request is required")
	}
	if r.Body == nil {
		return errors.New("request body is empty")
	}
	if wallet == nil {
		return errors.New("a wallet is required")
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, wallet)
}

// Get one walllet from the database
func walletsGetOne(w http.ResponseWriter, r *http.Request, id bson.ObjectId) {
	fetchedWallet, err := wallet.One(id)
	if err != nil {
		if err == storm.ErrNotFound {
			postError(w, http.StatusNotFound)
			return
		}
		postError(w, http.StatusInternalServerError)
		return
	}
	postBodyResponse(w, http.StatusOK, jsonResponse{"wallet": fetchedWallet})
}

// Get all wallets from the database
func walletsGetAll(w http.ResponseWriter, r *http.Request) {
	wallets, err := wallet.All()
	if err != nil {
		postError(w, http.StatusInternalServerError)
	}

	postBodyResponse(w, http.StatusOK, jsonResponse{"wallets": wallets})
}

// Create new wallet
func walletsPostOne(w http.ResponseWriter, r *http.Request) {
	newWallet := new(wallet.Wallet)
	err := bodyToWallet(r, newWallet)

	if err != nil {
		fmt.Println(err.Error())
		postError(w, http.StatusBadRequest)
		return
	}

	newWallet.ID = bson.NewObjectId()
	for i := range newWallet.Currencies {
		newWallet.Currencies[i].Available = newWallet.Currencies[i].Balance
		newWallet.Currencies[i].Reserved = 0
	}

	err = newWallet.Save()
	if err != nil {
		if err == wallet.ErrRecordInvalid {
			fmt.Println(err.Error())
			postError(w, http.StatusBadRequest)
		} else {
			postError(w, http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Location", "/wallets/"+newWallet.ID.Hex())
	w.WriteHeader(http.StatusCreated)
}

type NewCurrencyElement struct {
	Name   string `json:"name"`
	Amount *int   `json:"amount"`
}

// Patch a wallet with different currency elements
// TODO if name exists, increment balance and available amount with update.amount
// if not, add new currency
func walletsAddCurrency(w http.ResponseWriter, r *http.Request, id bson.ObjectId) {
	fetchedWallet, err := wallet.One(id)
	if err != nil {
		if err == storm.ErrNotFound {
			postError(w, http.StatusNotFound)
			return
		}
		postError(w, http.StatusInternalServerError)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var update NewCurrencyElement
	json.Unmarshal(body, &update)

	missing := true

	// Update the wallet's currencies based on the update data
	for i, currency := range fetchedWallet.Currencies {
		if currency.Name == update.Name {
			missing = false
			fetchedWallet.Currencies[i].Balance += *update.Amount
			fetchedWallet.Currencies[i].Available += *update.Amount
			break
		}
	}

	// Add new currency if it does not already exist in the wallet
	if missing {
		newCurrency := wallet.Currency{Name: update.Name, Balance: *update.Amount, Reserved: 0, Available: *update.Amount}
		fetchedWallet.Currencies = append(fetchedWallet.Currencies, newCurrency)
	}

	// Save the updated wallet to the database with new currency element
	err = fetchedWallet.Save()
	if err != nil {
		if err == wallet.ErrRecordInvalid {
			postError(w, http.StatusBadRequest)
		} else {
			postError(w, http.StatusInternalServerError)
		}
		return
	}

	postBodyResponse(w, http.StatusOK, jsonResponse{"wallet": fetchedWallet})
}

// Delete one wallet by id
func walletsDeleteOne(w http.ResponseWriter, r *http.Request, id bson.ObjectId) {
	err := wallet.Delete(id)
	if err != nil {
		if err == storm.ErrNotFound {
			postError(w, http.StatusNotFound)
			return
		}
		postError(w, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
