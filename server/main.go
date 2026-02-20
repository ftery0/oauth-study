package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ftery0/ouath/server/config"
	"github.com/ftery0/ouath/server/router"
)

func main() {
	cfg := config.Load()
	mux := router.New()

	addr := ":" + cfg.Port
	fmt.Printf("ouath server running on %s\n", cfg.Issuer)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
