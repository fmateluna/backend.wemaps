package main

import (
	"fmt"
	"os"

	"wemaps/internal/adapters/http"

	"cmp"
)

func main() {
	port := cmp.Or(os.Getenv("PORT"), "80")

	// Iniciar el servidor HTTP
	httpServer := http.NewServer()
	err := httpServer.Start(port)
	if err != nil {
		fmt.Printf("Error iniciando servidor HTTP: %v\n", err)
		os.Exit(1)
	}
}
