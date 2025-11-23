package db

import (
	"context"
	"sync"
)

// Manager maintains singleton connectors per database name
type Manager struct {
	mu         sync.Mutex
	connectors map[string]DBConnector
}

var globalMgr = &Manager{connectors: map[string]DBConnector{}}

// GetMongoConnector returns a singleton MongoConnector for the given dbName.
// If not connected, caller should call Connect.
func GetMongoConnector(dbName string) *MongoConnector {
	globalMgr.mu.Lock()
	defer globalMgr.mu.Unlock()
	if v, ok := globalMgr.connectors[dbName]; ok {
		if mc, ok2 := v.(*MongoConnector); ok2 {
			return mc
		}
	}
	mc := NewMongoConnector(dbName)
	globalMgr.connectors[dbName] = mc
	return mc
}

// CloseAll disconnects all connectors
func CloseAll(ctx context.Context) error {
	globalMgr.mu.Lock()
	defer globalMgr.mu.Unlock()
	for _, c := range globalMgr.connectors {
		_ = c.Close(ctx)
	}
	globalMgr.connectors = map[string]DBConnector{}
	return nil
}
