package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func (fe *frontend) apiPortfolioHistory(w http.ResponseWriter, req *http.Request) {
	id, ok := mux.Vars(req)["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		// TODO return proper json error and handle in the client
		return
	}
	type resp [][2]interface{}

	var r resp
	vals, err := fe.DB.UserValuationHistory(req.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		// TODO return proper json error and handle in the client
		return
	}
	for _, v := range vals {
		d := v.Date.Unix() * 1000
		c, _ := v.Value.F().Float64()
		r = append(r, [2]interface{}{d, c})
	}
	json.NewEncoder(w).Encode(r) // TODO handle err
}
