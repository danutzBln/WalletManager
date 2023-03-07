package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"WalletManager/wallet"

	"github.com/asdine/storm"
	"gopkg.in/mgo.v2/bson"
)

func bodyToWallet(r *http.Request, wallet *wallet.Wallet) error {
	if r.Body == nil {
		return errors.New("request body is empty")
	}

	if wallet == nil {
		return errors.New("a wallet is required")
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, wallet)
}

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
	newWallet.Currency.Available = newWallet.Currency.Balance
	newWallet.Currency.Reserved = 0

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

// Reserve an amount of a given currency from a wallet
func walletsReserveCurrency(w http.ResponseWriter, r *http.Request, id bson.ObjectId) {
	postBodyResponse(w, http.StatusOK, jsonResponse{"id": id})
}
