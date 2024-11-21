package network

import (
	"context"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

type RPCEndpoint struct {
	URL      string
	Client   *ethclient.Client
	Latency  time.Duration
	Weight   int
	LastUsed time.Time
}

type LoadBalancer struct {
	endpoints []*RPCEndpoint
	mu        sync.Mutex
}

func NewLoadBalancer(urls []string) (*LoadBalancer, error) {
	lb := &LoadBalancer{}
	for _, url := range urls {
		client, err := ethclient.Dial(url)
		if err != nil {
			return nil, err
		}
		lb.endpoints = append(lb.endpoints, &RPCEndpoint{
			URL:    url,
			Client: client,
			Weight: 1,
		})
	}
	return lb, nil
}

func (lb *LoadBalancer) UpdateLatencies() {
	var wg sync.WaitGroup
	for _, endpoint := range lb.endpoints {
		wg.Add(1)
		go func(e *RPCEndpoint) {
			defer wg.Done()
			start := time.Now()
			_, err := e.Client.BlockNumber(context.Background())
			if err != nil {
				e.Latency = time.Hour // Set a high latency for failed requests
			} else {
				e.Latency = time.Since(start)
			}
		}(endpoint)
	}
	wg.Wait()

	lb.mu.Lock()
	defer lb.mu.Unlock()

	// Sort endpoints by latency
	sort.Slice(lb.endpoints, func(i, j int) bool {
		return lb.endpoints[i].Latency < lb.endpoints[j].Latency
	})

	// Update weights based on latency
	totalWeight := 0
	for i, e := range lb.endpoints {
		e.Weight = len(lb.endpoints) - i
		totalWeight += e.Weight
	}

	// Normalize weights
	for _, e := range lb.endpoints {
		e.Weight = (e.Weight * 100) / totalWeight
	}
}

func (lb *LoadBalancer) GetClient() *ethclient.Client {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	now := time.Now()
	var selectedEndpoint *RPCEndpoint
	maxScore := -1.0

	for _, e := range lb.endpoints {
		timeFactor := now.Sub(e.LastUsed).Seconds() / 60 // Time factor increases every minute
		score := float64(e.Weight) * (1 + timeFactor)
		if score > maxScore {
			maxScore = score
			selectedEndpoint = e
		}
	}

	selectedEndpoint.LastUsed = now
	return selectedEndpoint.Client
}

func (lb *LoadBalancer) FetchBlocksRange(startBlock, endBlock uint64) ([]*big.Int, error) {
	var blocks []*big.Int
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := startBlock; i <= endBlock; i++ {
		wg.Add(1)
		go func(blockNum uint64) {
			defer wg.Done()
			client := lb.GetClient()
			block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(blockNum)))
			if err != nil {
				// Handle error (maybe retry with a different client)
				return
			}
			mu.Lock()
			blocks = append(blocks, block.Number())
			mu.Unlock()
		}(i)
	}

	wg.Wait()
	return blocks, nil
}
