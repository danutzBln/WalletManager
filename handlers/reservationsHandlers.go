package handlers

import (
	"WalletManager/wallet"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/asdine/storm"
	"gopkg.in/mgo.v2/bson"
)

type OngoingReservations struct {
	Name         string               `json:"name"`
	Reservations []wallet.Reservation `json:"reservations"`
}

// Get all reservations from a wallet
func reservationsGetAll(w http.ResponseWriter, r *http.Request, id bson.ObjectId) {
	fetchedWallet, err := wallet.One(id)
	if err != nil {
		if err == storm.ErrNotFound {
			postError(w, http.StatusNotFound)
			return
		}
		postError(w, http.StatusInternalServerError)
		return
	}

	var reservations []OngoingReservations

	// Get the currency's reservations from the wallet
	for i := range fetchedWallet.Currencies {
		currentReservations := OngoingReservations{
			Name:         fetchedWallet.Currencies[i].Name,
			Reservations: fetchedWallet.Currencies[i].Reservations,
		}
		reservations = append(reservations, currentReservations)
	}

	postBodyResponse(w, http.StatusOK, jsonResponse{"ongoing": reservations})
}

type ReservationUpdate struct {
	Name   string `json:"name"`
	Amount *int   `json:"amount"`
}

// Reserve an amount of a given currency from a wallet
func reservationsCreateOne(w http.ResponseWriter, r *http.Request, id bson.ObjectId) {
	fetchedWallet, err := wallet.One(id)
	if err != nil {
		if err == storm.ErrNotFound {
			postError(w, http.StatusNotFound)
			return
		}
		postError(w, http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(r.Body)
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
