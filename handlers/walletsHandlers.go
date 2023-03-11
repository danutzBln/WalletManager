package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"WalletManager/wallet"

	"github.com/asdine/storm"
	"gopkg.in/mgo.v2/bson"
)

// Encode a request body to a wallet struct
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

// Patch a wallet with different currency elements
func walletsPatchOne(w http.ResponseWriter, r *http.Request, id bson.ObjectId) {
	fetchedWallet, err := wallet.One(id)
	if err != nil {
		if err == storm.ErrNotFound {
			postError(w, http.StatusNotFound)
			return
		}
		postError(w, http.StatusInternalServerError)
		return
	}
	err = bodyToWallet(r, fetchedWallet)
	if err != nil {
		postError(w, http.StatusBadRequest)
		return
	}
	fetchedWallet.ID = id
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

type ReservationUpdate struct {
	Name   string `json:"name"`
	Amount *int   `json:"amount"`
}

// Reserve an amount of a given currency from a wallet
func walletsReserveCurrency(w http.ResponseWriter, r *http.Request, id bson.ObjectId) {
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
	var update ReservationUpdate
	json.Unmarshal(body, &update)

	missing := true

	// Update the wallet's currencies based on the update data
	for i, currency := range fetchedWallet.Currencies {
		if currency.Name == update.Name {
			missing = false
			if update.Amount != nil && fetchedWallet.Currencies[i].Available-*update.Amount >= 0 {
				newReservation := wallet.Reservation{
					Amount:          *update.Amount,
					ReservationTime: time.Now(),
				}
				fetchedWallet.Currencies[i].Reservations = append(fetchedWallet.Currencies[i].Reservations, newReservation)
				fetchedWallet.Currencies[i].Reserved += *update.Amount
				fetchedWallet.Currencies[i].Available = fetchedWallet.Currencies[i].Balance - fetchedWallet.Currencies[i].Reserved
			} else {
				err := fmt.Errorf("not enough funds available to spend for currency %s - %d", fetchedWallet.Currencies[i].Name, fetchedWallet.Currencies[i].Available)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			break
		}
	}

	if missing {
		err := fmt.Errorf("wallet does not contain currency %s", update.Name)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Save the updated wallet to the database if enough funds are available
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
