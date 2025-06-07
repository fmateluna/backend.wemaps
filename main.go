package main

import (
	"cmp"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"wemaps/internal/adapters/http"
	"wemaps/internal/infrastructure/repository"
)

type TLSConfig struct {
	CertFile string `json:"certFile"`
	KeyFile  string `json:"keyFile"`
}

func main() {
	var httpsConfigPath string
	flag.StringVar(&httpsConfigPath, "https", "", "Ruta al archivo JSON con configuración TLS (cert y key)")
	flag.Parse()

	port := cmp.Or(os.Getenv("PORT"), "80")
	var certFile, keyFile string

	// Si se pasa el flag -https, leer el archivo JSON
	if httpsConfigPath != "" {
		file, err := os.Open(httpsConfigPath)
		if err != nil {
			fmt.Printf("Error abriendo archivo de configuración HTTPS: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		var tlsConfig TLSConfig
		if err := json.NewDecoder(file).Decode(&tlsConfig); err != nil {
			fmt.Printf("Error decodificando archivo JSON: %v\n", err)
			os.Exit(1)
		}

		certFile = tlsConfig.CertFile
		keyFile = tlsConfig.KeyFile
	}

	reporPortal, errorPostgress := repository.NewPostgresDBRepository()
	if errorPostgress != nil {
		fmt.Println("Error: No se pudo inicializar el repositorio Postgress")

	}

	repoAddress, errorMongo := repository.NewMongoDBRepository()
	if errorMongo != nil {
		fmt.Println("Error: No se pudo inicializar el repositorio MongoDB")
		os.Exit(1)
	}
	defer func() {
		if err := repoAddress.Close(context.Background()); err != nil {
			fmt.Printf("Error desconectando de MongoDB: %v\n", err)
		}
	}()

	httpServer := http.NewServer(repoAddress, reporPortal)

	if err := httpServer.StartServer(port, certFile, keyFile); err != nil {
		fmt.Printf("Error iniciando servidor: %v\n", err)
		os.Exit(1)
	}
}
