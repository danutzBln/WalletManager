package wallet

import (
	"errors"
	"time"

	"github.com/asdine/storm"
	"gopkg.in/mgo.v2/bson"
)

type Reservation struct {
	Amount          int       `json:"balance"`
	ReservationTime time.Time `json:"reservationtime"`
}

type Currency struct {
	Name         string        `json:"name"`
	Balance      int           `json:"balance"`
	Reservations []Reservation `json:"reservations"`
	Reserved     int           `json:"reserved"`
	Available    int           `json:"available"`
}

// Wallet holds data for
type Wallet struct {
	ID         bson.ObjectId `json:"id" storm:"id"`
	Currencies []Currency    `json:"currencies"`
}

const (
	dbPath = "wallets.db"
)

// errors
var ErrRecordInvalid = errors.New("record is invalid")

// All retrieves all wallets from database
func All() ([]Wallet, error) {
	db, err := storm.Open(dbPath)

	if err != nil {
		return nil, err
	}

	defer db.Close()
	wallets := []Wallet{}
	err = db.All(&wallets)

	if err != nil {
		return nil, err
	}
	return wallets, nil
}

// One returns a single wallet record from the database
func One(id bson.ObjectId) (*Wallet, error) {
	db, err := storm.Open(dbPath)

	if err != nil {
		return nil, err
	}

	defer db.Close()

	wallet := new(Wallet)
	err = db.One("ID", id, wallet)

	if err != nil {
		return nil, err
	}

	return wallet, nil
}

// Delete removes a given wallet record from the database
func Delete(id bson.ObjectId) error {
	db, err := storm.Open(dbPath)
	if err != nil {
		return err
	}

	defer db.Close()

	wallet := new(Wallet)
	err = db.One("ID", id, wallet)

	if err != nil {
		return err
	}

	return db.DeleteStruct(wallet)
}

// Save updates or creates a given record in the database
func (wallet *Wallet) Save() error {
	if err := wallet.validateWalletFields(); err != nil {
		return err
	}
	db, err := storm.Open(dbPath)
	if err != nil {
		return err
	}

	defer db.Close()

	return db.Save(wallet)
}

// Validate if entered wallet record contains valid data
func (wallet *Wallet) validateWalletFields() error {
	if wallet.ID.Hex() == "" || len(wallet.Currencies) == 0 {
		return ErrRecordInvalid
	}

	for _, c := range wallet.Currencies {
		if c.Name == "" {
			return ErrRecordInvalid
		}
	}

	return nil
}
