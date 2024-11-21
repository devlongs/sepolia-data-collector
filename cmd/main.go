package main

import (
	"log"
	"os"
	"strings"
	"time"

	ethereum "github.com/devlongs/sepolia-data-collector/internal/network"
	"github.com/devlongs/sepolia-data-collector/internal/storage"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	rpcUrls := strings.Split(os.Getenv("RPC_URLS"), ",")
	client, err := ethereum.NewClient(rpcUrls)
	if err != nil {
		log.Fatalf("Failed to initialize Ethereum client: %v", err)
	}

	db, err := storage.NewLevelDB(os.Getenv("LEVELDB_PATH"))
	if err != nil {
		log.Fatalf("Failed to initialize LevelDB: %v", err)
	}
	defer db.Close()

	// Start a goroutine to update latencies periodically
	go func() {
		for {
			client.LB.UpdateLatencies()
			time.Sleep(5 * time.Minute)
		}
	}()

	err = client.FetchAndStoreEvents(db, os.Getenv("CONTRACT_ADDRESS"), os.Getenv("TOPIC_HASH"))
	if err != nil {
		log.Fatalf("Failed to fetch and store events: %v", err)
	}

	log.Println("Event data successfully stored in LevelDB.")
}
