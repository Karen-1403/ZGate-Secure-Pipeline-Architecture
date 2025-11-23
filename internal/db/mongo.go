package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoConnector implements DBConnector for MongoDB
type MongoConnector struct {
	client *mongo.Client
	dbName string
}

func NewMongoConnector(dbName string) *MongoConnector {
	return &MongoConnector{dbName: dbName}
}

func (m *MongoConnector) Connect(ctx context.Context, uri string) error {
	if m.client != nil {
		return nil
	}
	opts := options.Client().ApplyURI(uri)
	client, err := mongo.NewClient(opts)
	if err != nil {
		return err
	}
	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := client.Connect(ctx2); err != nil {
		return err
	}
	// Optionally ping
	if err := client.Ping(ctx2, nil); err != nil {
		return err
	}
	m.client = client
	return nil
}

func (m *MongoConnector) ExecuteQuery(ctx context.Context, req QueryRequest) (interface{}, error) {
	if m.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	coll := m.client.Database(m.dbName).Collection(req.Collection)
	switch req.Action {
	case "find":
		// Convert filter to bson.M
		filter := bson.M{}
		if req.Filter != nil {
			filter = req.Filter
		}
		cursor, err := coll.Find(ctx, filter)
		if err != nil {
			return nil, err
		}
		var results []bson.M
		if err := cursor.All(ctx, &results); err != nil {
			return nil, err
		}
		return results, nil
	default:
		return nil, fmt.Errorf("unsupported action: %s", req.Action)
	}
}

func (m *MongoConnector) Close(ctx context.Context) error {
	if m.client == nil {
		return nil
	}
	return m.client.Disconnect(ctx)
}
