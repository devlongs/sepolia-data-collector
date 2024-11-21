package network

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/devlongs/sepolia-data-collector/internal/storage"
	models "github.com/devlongs/sepolia-data-collector/internal/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

type Client struct {
	LB *LoadBalancer
}

func NewClient(urls []string) (*Client, error) {
	lb, err := NewLoadBalancer(urls)
	if err != nil {
		return nil, fmt.Errorf("failed to create load balancer: %v", err)
	}
	return &Client{LB: lb}, nil
}

func (c *Client) getContractCreationBlock(contractAddress common.Address) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	client := c.LB.GetClient()
	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("Failed to get latest block: %v", err)
	}
	latestBlock := header.Number.Uint64()

	var creationBlock uint64
	low := uint64(0)
	high := latestBlock

	for low <= high {
		mid := (low + high) / 2

		code, err := client.CodeAt(ctx, contractAddress, big.NewInt(int64(mid)))
		if err != nil {
			return 0, fmt.Errorf("Failed to get code at block %d: %v", mid, err)
		}

		if len(code) > 0 {
			creationBlock = mid
			high = mid - 1
		} else {
			low = mid + 1
		}
	}

	if creationBlock == 0 {
		return 0, fmt.Errorf("Contract creation block not found, check the address again")
	}

	return creationBlock, nil
}

func (c *Client) FetchAndStoreEvents(db *storage.LevelDB, contractAddressStr, topicHashStr string) error {
	contractAddress := common.HexToAddress(contractAddressStr)
	topicHash := common.HexToHash(topicHashStr)

	startBlock, err := c.getContractCreationBlock(contractAddress)
	if err != nil {
		return fmt.Errorf("failed to get contract creation block: %v", err)
	}
	log.Printf("Contract was deployed at block number: %d\n", startBlock)

	client := c.LB.GetClient()
	latestBlockHeader, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("failed to fetch the latest block header: %v", err)
	}
	latestBlockNumber := latestBlockHeader.Number.Uint64()
	log.Printf("Latest block number: %d\n", latestBlockNumber)

	blockRangeSize := uint64(10000)
	index := 0

	for startBlock <= latestBlockNumber {
		endBlock := startBlock + blockRangeSize
		if endBlock > latestBlockNumber {
			endBlock = latestBlockNumber
		}

		log.Printf("Querying logs from block %d to %d...\n", startBlock, endBlock)

		query := ethereum.FilterQuery{
			Addresses: []common.Address{contractAddress},
			Topics:    [][]common.Hash{{topicHash}},
			FromBlock: big.NewInt(int64(startBlock)),
			ToBlock:   big.NewInt(int64(endBlock)),
		}

		client := c.LB.GetClient()
		logs, err := client.FilterLogs(context.Background(), query)
		if err != nil {
			return fmt.Errorf("failed to filter logs: %v", err)
		}

		log.Printf("Found %d logs in the range [%d, %d]\n", len(logs), startBlock, endBlock)

		for _, vLog := range logs {
			log.Printf("Processing log from block %d, transaction hash: %s\n", vLog.BlockNumber, vLog.TxHash.Hex())

			client := c.LB.GetClient()
			block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(vLog.BlockNumber)))
			if err != nil {
				return fmt.Errorf("failed to fetch block: %v", err)
			}

			eventData := &models.EventData{
				L1InfoRoot: common.BytesToHash(vLog.Data).Hex(),
				BlockTime:  block.Time(),
				ParentHash: block.ParentHash().Hex(),
			}

			if err := db.StoreEvent(index, int(vLog.Index), eventData); err != nil {
				return fmt.Errorf("failed to store event: %v", err)
			}

			index++
		}

		startBlock = endBlock + 1
	}

	return nil
}
