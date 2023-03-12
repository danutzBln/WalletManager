package handlers

import (
	"net/http"
	"strings"

	"gopkg.in/mgo.v2/bson"
)

// ReservationsRouter handles the reserve route
func ReservationsRouter(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/reserve/")

	// Check if an id is present after the path
	path = strings.TrimPrefix(path, "/reserve/")

	if !bson.IsObjectIdHex(path) {
		postError(w, http.StatusNotFound)
		return
	}

	id := bson.ObjectIdHex(path)

	switch r.Method {
	case http.MethodGet:
		reservationsGetAll(w, r, id)
		return
	case http.MethodPut:
		reservationsCreateOne(w, r, id)
		return
	default:
		postError(w, http.StatusMethodNotAllowed)
	}
}
