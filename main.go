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
	http.HandleFunc("/auth/auth", authHandler)
	http.HandleFunc("/auth/refresh", refreshAuthHandler)
	http.HandleFunc("/auth/checkAuth", checkAuthHandler)
	http.HandleFunc("/ping", pingHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
