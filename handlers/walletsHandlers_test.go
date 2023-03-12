package handlers

import (
	"WalletManager/wallet"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"testing"
	"time"

	"gopkg.in/mgo.v2/bson"
)

func Test(t *testing.T) {
	valid := &wallet.Wallet{
		ID: bson.NewObjectId(),
		Currencies: []wallet.Currency{
			{
				Name:      "USD",
				Balance:   100,
				Reserved:  0,
				Available: 90,
				Reservations: []wallet.Reservation{
					{
						Amount:          10,
						ReservationTime: time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		},
	}
	js, err := json.Marshal(valid)
	if err != nil {
		t.Errorf("Error marshalling a valid wallet: %s", err)
		t.FailNow()
	}

	ts := []struct {
		txt string
		r   *http.Request
		w   *wallet.Wallet
		err bool
		exp *wallet.Wallet
	}{
		{
			txt: "nil request",
			err: true,
		},
		{
			txt: "empty request body",
			r:   &http.Request{},
			err: true,
		},
		{
			txt: "empty wallet",
			r: &http.Request{
				Body: io.NopCloser(bytes.NewBufferString("{}")),
			},
			err: true,
		},
		{
			txt: "malformed data",
			r: &http.Request{
				Body: io.NopCloser(bytes.NewBufferString(`{"id": 999}`)),
			},
			w:   &wallet.Wallet{},
			err: true,
		},
		{
			txt: "valid request",
			r: &http.Request{
				Body: io.NopCloser(bytes.NewBuffer(js)),
			},
			w:   &wallet.Wallet{},
			exp: valid,
		},
	}

	for _, tc := range ts {
		t.Log(tc.txt)
		err := bodyToWallet(tc.r, tc.w)
		if tc.err {
			if err == nil {
				t.Error("Expected error, got none")
			}
			continue
		}
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
			continue
		}
		if !reflect.DeepEqual(tc.w, tc.exp) {
			t.Error("Unmarshalled data is different:")
			t.Error(tc.w)
			t.Error(tc.exp)
		}
	}
}
