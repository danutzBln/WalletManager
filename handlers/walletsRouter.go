package handlers

import (
	"net/http"
	"strings"

	"gopkg.in/mgo.v2/bson"
)

// WalletsRouter handles the wallets route
func WalletsRouter(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")
	if path == "/wallets" {
		switch r.Method {
		case http.MethodGet:
			walletsGetAll(w, r)
			return
		case http.MethodPost:
			walletsPostOne(w, r)
			return
		default:
			postError(w, http.StatusMethodNotAllowed)
			return
		}
	}

	// Check if an id is present after the path
	path = strings.TrimPrefix(path, "/wallets/")
	if !bson.IsObjectIdHex(path) {
		postError(w, http.StatusNotFound)
		return
	}

	id := bson.ObjectIdHex(path)

	switch r.Method {
	case http.MethodGet:
		walletsGetOne(w, r, id)
		return
	case http.MethodPut:
		walletsPostOne(w, r)
		return
	case http.MethodPatch:
		walletsAddCurrency(w, r, id)
		return
	case http.MethodDelete:
		walletsDeleteOne(w, r, id)
		return
	default:
		postError(w, http.StatusMethodNotAllowed)
	}
}
