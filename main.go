package main

import (
	"context"
	"fmt"
	"os"
	"wemaps/internal/adapters/http"
	"wemaps/internal/infrastructure/repository"

	"cmp"
)

func main() {
	port := cmp.Or(os.Getenv("PORT"), "80")

	repo := repository.NewMongoDBRepository()
	if repo == nil {
		fmt.Println("Error: No se pudo inicializar el repositorio MongoDB")
		os.Exit(1)
	}

	defer func() {
		if err := repo.Close(context.Background()); err != nil {
			fmt.Printf("Error desconectando de MongoDB: %v\n", err)
		}
	}()

	httpServer := http.NewServer(repo)
	err := httpServer.Start(port)
	if err != nil {
		fmt.Printf("Error iniciando servidor HTTP: %v\n", err)
		os.Exit(1)
	}
}
