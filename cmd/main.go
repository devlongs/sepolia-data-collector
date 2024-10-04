package main

import (
	"log"
	"os"

	ethereum "github.com/devlongs/sepolia-data-collector/internal/network"
	"github.com/devlongs/sepolia-data-collector/internal/storage"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	client, err := ethereum.NewClient(os.Getenv("RPC_URL"))
	if err != nil {
		log.Fatalf("Failed to initialize Ethereum client: %v", err)
	}
	defer client.Close()

	db, err := storage.NewLevelDB(os.Getenv("LEVELDB_PATH"))
	if err != nil {
		log.Fatalf("Failed to initialize LevelDB: %v", err)
	}
	defer db.Close()

	err = client.FetchAndStoreEvents(db, os.Getenv("CONTRACT_ADDRESS"), os.Getenv("TOPIC_HASH"))
	if err != nil {
		log.Fatalf("Failed to fetch and store events: %v", err)
	}

	log.Println("Event data successfully stored in LevelDB.")
}
