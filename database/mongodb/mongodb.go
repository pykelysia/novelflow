package mongodb

import (
	"context"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoClient struct {
	*mongo.Client
}

func NewMongoDB() (*MongoClient, error) {
	uri := viper.GetString("mongo")

	serverApi := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().
		ApplyURI(uri).
		SetServerAPIOptions(serverApi)

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, err
	}

	return &MongoClient{
		Client: client,
	}, nil
}

func (m *MongoClient) Close() error {
	if err := m.Disconnect(context.TODO()); err != nil {
		return err
	}
	return nil
}
