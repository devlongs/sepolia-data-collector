# Sepolia Data Collector 

## Overview
The Sepolia Data Collector is a Go-based application designed to efficiently download and store data from the Sepolia Ethereum testnet. It focuses on retrieving specific event logs from a designated smart contract and storing the associated L1 info root, block time, and parent hash data.

## Features
- Fetch logs from the Sepolia testnet using Ethereum smart contract events.
- Store event data efficiently using LevelDB.
- Support batch log fetching to minimize network overhead.
- Easily configurable through environment variables.
- Robust error handling and logging for seamless blockchain interaction.

## Prerequisites

- Go 1.18 or higher
- Access to a Sepolia Ethereum node (e.g., via alchemy or infura)
- Protocol buffer compiler (protoc) to regenerate protocol buffer files, if needed.


## Installation

1. Clone the repository:
```bash
clone https://github.com/devlongs/sepolia-data-collector.git
cd sepolia-data-collector
```

2. Install dependencies:
```bash
mod tidy
```

## Configuration
Create a .env file in the root directory with the following contents:
```bash
RPC_URL=https://sepolia.infura.io/v3/YOUR-PROJECT-ID
CONTRACT_ADDRESS=0x761d53b47334bee6612c0bd1467fb881435375b2
TOPIC_HASH=0x3e54d0825ed78523037d00a81759237eb436ce774bd546993ee67a1b67b6e766
BLOCK_RANGE_SIZE=10000
LEVELDB_PATH=events_database
```

Replace YOUR-PROJECT-ID with your actual Infura project ID or use a different Sepolia endpoint if preferred.


## Usage
To run the data collector:
```bash
go run cmd/main.go
```
This will start the process of fetching event logs from the specified contract on Sepolia and storing them in the LevelDB database.

## Testing
To run the test suite:
```bash
go test -v ./...
```
This will run all tests, including storage tests and any other tests you've added to the project.

## Project Structure

- cmd/main.go: Entry point of the application
- internal/network/: Contains Ethereum client implementation
- internal/storage/: Implements LevelDB storage functionality
- internal/types/: Defines data models used in the project
- proto/: Contains Protocol Buffer definitions (if used)