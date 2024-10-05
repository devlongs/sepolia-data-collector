package storage

import (
	"fmt"
	"os"
	"testing"

	models "github.com/devlongs/sepolia-data-collector/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
)

func TestLevelDBStorage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "leveldb_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	db, err := NewLevelDB(tempDir)
	if err != nil {
		t.Fatalf("Failed to create LevelDB: %v", err)
	}
	defer db.Close()

	testCases := []struct {
		index     int
		logIndex  int
		eventData *models.EventData
	}{
		{
			index:    0,
			logIndex: 0,
			eventData: &models.EventData{
				L1InfoRoot: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				BlockTime:  1234567890,
				ParentHash: "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			},
		},
		{
			index:    1,
			logIndex: 2,
			eventData: &models.EventData{
				L1InfoRoot: "0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321",
				BlockTime:  9876543210,
				ParentHash: "0x0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
		},
	}

	for _, tc := range testCases {
		err := db.StoreEvent(tc.index, tc.logIndex, tc.eventData)
		assert.NoError(t, err, "Failed to store event")
	}

	for _, tc := range testCases {
		retrievedData, err := db.RetrieveEvent(tc.index, tc.logIndex)
		assert.NoError(t, err, "Failed to retrieve event")
		assert.Equal(t, tc.eventData.L1InfoRoot, retrievedData.L1InfoRoot, "Retrieved L1InfoRoot does not match")
		assert.Equal(t, tc.eventData.BlockTime, retrievedData.BlockTime, "Retrieved BlockTime does not match")
		assert.Equal(t, tc.eventData.ParentHash, retrievedData.ParentHash, "Retrieved ParentHash does not match")
	}

	_, err = db.RetrieveEvent(999, 999)
	assert.Error(t, err, "Expected error when retrieving non-existent data")
	assert.Equal(t, leveldb.ErrNotFound, err, "Expected leveldb.ErrNotFound")

	newEventData := &models.EventData{
		L1InfoRoot: "0x9999999999999999999999999999999999999999999999999999999999999999",
		BlockTime:  5555555555,
		ParentHash: "0x8888888888888888888888888888888888888888888888888888888888888888",
	}
	err = db.StoreEvent(0, 0, newEventData)
	assert.NoError(t, err, "Failed to overwrite existing event")

	retrievedData, err := db.RetrieveEvent(0, 0)
	assert.NoError(t, err, "Failed to retrieve overwritten event")
	assert.Equal(t, newEventData.L1InfoRoot, retrievedData.L1InfoRoot, "Retrieved L1InfoRoot does not match overwritten data")
	assert.Equal(t, newEventData.BlockTime, retrievedData.BlockTime, "Retrieved BlockTime does not match overwritten data")
	assert.Equal(t, newEventData.ParentHash, retrievedData.ParentHash, "Retrieved ParentHash does not match overwritten data")

	for i := 0; i < 1000; i++ {
		eventData := &models.EventData{
			L1InfoRoot: fmt.Sprintf("0x%064d", i),
			BlockTime:  uint64(i),
			ParentHash: fmt.Sprintf("0x%064d", i*2),
		}
		err := db.StoreEvent(i, 0, eventData)
		assert.NoError(t, err, "Failed to store event in bulk test")
	}

	for i := 0; i < 1000; i++ {
		retrievedData, err := db.RetrieveEvent(i, 0)
		assert.NoError(t, err, "Failed to retrieve event in bulk test")
		assert.Equal(t, fmt.Sprintf("0x%064d", i), retrievedData.L1InfoRoot, "Retrieved L1InfoRoot does not match in bulk test")
		assert.Equal(t, uint64(i), retrievedData.BlockTime, "Retrieved BlockTime does not match in bulk test")
		assert.Equal(t, fmt.Sprintf("0x%064d", i*2), retrievedData.ParentHash, "Retrieved ParentHash does not match in bulk test")
	}

	eventData1 := &models.EventData{
		L1InfoRoot: "0xAAAA",
		BlockTime:  1,
		ParentHash: "0xBBBB",
	}

	eventData2 := &models.EventData{
		L1InfoRoot: "0xCCCC",
		BlockTime:  2,
		ParentHash: "0xDDDD",
	}

	err = db.StoreEvent(5, 0, eventData1)
	assert.NoError(t, err, "Failed to store event with index 5, logIndex 0")

	err = db.StoreEvent(5, 1, eventData2)
	assert.NoError(t, err, "Failed to store event with index 5, logIndex 1")

	retrievedData, err = db.RetrieveEvent(5, 0)
	assert.NoError(t, err, "Failed to retrieve event with index 5, logIndex 0")
	assert.Equal(t, "0xAAAA", retrievedData.L1InfoRoot)

	retrievedData, err = db.RetrieveEvent(5, 1)
	assert.NoError(t, err, "Failed to retrieve event with index 5, logIndex 1")
	assert.Equal(t, "0xCCCC", retrievedData.L1InfoRoot)
}
