package storage

import (
	"fmt"

	models "github.com/devlongs/sepolia-data-collector/internal/types"
	"github.com/syndtr/goleveldb/leveldb"
	"google.golang.org/protobuf/proto"
)

type LevelDB struct {
	*leveldb.DB
}

func NewLevelDB(path string) (*LevelDB, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open LevelDB: %v", err)
	}
	return &LevelDB{db}, nil
}

func (db *LevelDB) StoreEvent(index, logIndex int, eventData *models.EventData) error {
	key := fmt.Sprintf("%d-%d", index, logIndex)
	data, err := proto.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("failed to serialize event data: %v", err)
	}

	return db.Put([]byte(key), data, nil)
}

func (db *LevelDB) RetrieveEvent(index, logIndex int) (*models.EventData, error) {
	key := fmt.Sprintf("%d-%d", index, logIndex)
	data, err := db.Get([]byte(key), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, err
		}
		return nil, fmt.Errorf("failed to retrieve data: %v", err)
	}

	eventData := &models.EventData{}
	err = proto.Unmarshal(data, eventData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %v", err)
	}

	return eventData, nil
}
