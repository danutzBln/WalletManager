package handlers

import (
	"WalletManager/wallet"
	"net/http"
	"reflect"
	"testing"
)

func Test(t *testing.T) {
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
