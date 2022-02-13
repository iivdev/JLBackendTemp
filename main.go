package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	err := ReadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}
	initDBs()
	http.HandleFunc("/auth/register", registerHandler)
	http.HandleFunc("/auth/confirm", confirmHandler)
	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/auth/checkAuth", checkAuthHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
