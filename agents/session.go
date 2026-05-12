package agent

import (
	"context"
	"fmt"
	"novelflow/database/mongodb"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type SessionPart struct {
	SID       string    `bson:"_id"`
	Title     string    `bson:"title,omitempty"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

type Session struct {
	SessionID   string
	SessionPart SessionPart
	mongoClient *mongodb.MongoClient
}

func NewSession(ctx context.Context, sid string, mdb *mongodb.MongoClient) (*Session, error) {
	s := &Session{
		SessionID:   sid,
		mongoClient: mdb,
	}
	if len(sid) != 36 {
		sp := createNewSessionPart()
		s.SessionID = sp.SID
		s.SessionPart = sp
		if err := s.Save(); err != nil {
			return nil, fmt.Errorf("failed to save session: %v", err)
		}
	} else {
		result := SessionPart{}
		filter := bson.D{{Key: "_id", Value: sid}}
		err := s.mongoClient.Database("novelflow").Collection("sessions").FindOne(ctx, filter).Decode(&result)
		if err == mongo.ErrNoDocuments {
			sp := createNewSessionPart()
			s.SessionPart = sp
			if err := s.Save(); err != nil {
				return nil, fmt.Errorf("failed to save session: %v", err)
			}
		} else if err != nil {
			return nil, fmt.Errorf("error: %v", err)
		} else {
			s.SessionPart = result
		}
	}
	return s, nil
}

func createNewSessionPart() SessionPart {
	return SessionPart{
		SID:       uuid.New().String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (s *Session) Save() error {
	s.SessionPart.UpdatedAt = time.Now()
	filter := bson.D{{Key: "_id", Value: s.SessionPart.SID}}
	_, err := s.mongoClient.Database("novelflow").Collection("sessions").ReplaceOne(
		context.TODO(),
		filter,
		s.SessionPart,
		options.Replace().SetUpsert(true),
	)
	return err
}

func (s *Session) Append(msg Message) error {
	msg.SessionID = s.SessionPart.SID
	msg.CreatedAt = time.Now()
	_, err := s.mongoClient.Database("novelflow").Collection("messages").InsertOne(
		context.TODO(),
		msg,
	)
	return err
}

func (s *Session) Use() (msgs []adk.Message) {
	cursor, err := s.mongoClient.Database("novelflow").Collection("messages").Find(
		context.TODO(),
		bson.D{{Key: "session_id", Value: s.SessionPart.SID}},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}),
	)
	if err != nil {
		return nil
	}
	defer cursor.Close(context.TODO())

	var messages []Message
	if err := cursor.All(context.TODO(), &messages); err != nil {
		return nil
	}

	for _, m := range messages {
		if m.Type != ContentType {
			continue
		}
		msgs = append(msgs, &schema.Message{
			Role:    m.Role,
			Content: m.Content,
		})
	}
	return msgs
}
