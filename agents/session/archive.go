package session

import (
	"context"
	"fmt"
	"time"

	"novelflow/database/mongodb"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const archiveCollection = "tool_result_archives"

type ToolResultArchive struct {
	ID        string    `bson:"_id"`
	SessionID string    `bson:"session_id"`
	ToolName  string    `bson:"tool_name"`
	ToolArgs  string    `bson:"tool_args"`
	Result    string    `bson:"result"`
	Success   bool      `bson:"success"`
	CreatedAt time.Time `bson:"created_at"`
}

func NewArchive(sessionID, toolName, toolArgs, result string, success bool) ToolResultArchive {
	return ToolResultArchive{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		ToolName:  toolName,
		ToolArgs:  toolArgs,
		Result:    result,
		Success:   success,
		CreatedAt: time.Now(),
	}
}

func SaveArchive(ctx context.Context, mdb *mongodb.MongoClient, arch ToolResultArchive) error {
	_, err := mdb.Database("novelflow").Collection(archiveCollection).InsertOne(ctx, arch)
	return err
}

func LoadArchive(ctx context.Context, mdb *mongodb.MongoClient, id string) (*ToolResultArchive, error) {
	var arch ToolResultArchive
	err := mdb.Database("novelflow").Collection(archiveCollection).
		FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&arch)
	if err != nil {
		return nil, fmt.Errorf("archive %s not found: %w", id, err)
	}
	return &arch, nil
}
