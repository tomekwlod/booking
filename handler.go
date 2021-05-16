package main

import (
	"log"
	"net/http"

	_ "github.com/jackc/pgx/stdlib"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func dbtestHandler(w http.ResponseWriter, r *http.Request) {

	err := pconn.Ping()

	if err != nil {

		log.Printf("Error while pinging DB, %v", err)
		writeError(w, errInternalServer)
	}

	w.Write([]byte("ok"))
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {

	type jsonRes struct {
		Username string `json:"username"`
		Code     string `json:"code"`
		ID       int    `json:"id"`
	}

	j := &jsonRes{"Tomek27", "39s94jr99c", 1}

	writeJSON(w, j, http.StatusOK)
}
