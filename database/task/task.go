package task

import (
	"context"
	"time"

	"novelflow/database/mongodb"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type TaskStatus string

const (
	TaskPending   TaskStatus = "pending"
	TaskRunning   TaskStatus = "running"
	TaskCompleted TaskStatus = "completed"
	TaskFailed    TaskStatus = "failed"

	taskCollection = "tasks"
)

type Task struct {
	SessionID string     `bson:"_id"`
	UserID    uint       `bson:"user_id"`
	Status    TaskStatus `bson:"status"`
	Error     string     `bson:"error,omitempty"`
	CreatedAt time.Time  `bson:"created_at"`
	UpdatedAt time.Time  `bson:"updated_at"`
}

func collection(mdb *mongodb.MongoClient) *mongo.Collection {
	return mdb.Database("novelflow").Collection(taskCollection)
}

func CreateTask(ctx context.Context, mdb *mongodb.MongoClient, sessionID string, userID uint) (*Task, error) {
	now := time.Now()
	t := &Task{
		SessionID: sessionID,
		UserID:    userID,
		Status:    TaskPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err := collection(mdb).InsertOne(ctx, t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func UpdateTaskStatus(ctx context.Context, mdb *mongodb.MongoClient, sessionID string, status TaskStatus, errMsg string) error {
	set := bson.D{
		{Key: "status", Value: status},
		{Key: "updated_at", Value: time.Now()},
	}
	if errMsg != "" {
		set = append(set, bson.E{Key: "error", Value: errMsg})
	}
	_, err := collection(mdb).UpdateOne(
		ctx,
		bson.D{{Key: "_id", Value: sessionID}},
		bson.D{{Key: "$set", Value: set}},
	)
	return err
}

func GetTask(ctx context.Context, mdb *mongodb.MongoClient, sessionID string) (*Task, error) {
	var t Task
	err := collection(mdb).FindOne(ctx, bson.D{{Key: "_id", Value: sessionID}}).Decode(&t)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func ListUserTasks(ctx context.Context, mdb *mongodb.MongoClient, userID uint) ([]Task, error) {
	cursor, err := collection(mdb).Find(
		ctx,
		bson.D{{Key: "user_id", Value: userID}},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tasks []Task
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}
