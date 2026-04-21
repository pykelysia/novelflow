package agents

import (
	"context"
	"fmt"
	"novelflow/database/mongodb"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type SessionPart struct {
	SID      string
	Title    string
	Messages []Message
}

type Session struct {
	SessionID   string
	SessionPart SessionPart
	mongoClient *mongodb.MongoClient
}

func (s *Session) Save() error {
	filter := bson.D{{
		Key:   "session_id",
		Value: s.SessionPart.SID,
	}}
	_, err := s.mongoClient.Database("novelflow").Collection("sessions").ReplaceOne(
		context.TODO(),
		filter,
		s.SessionPart,
		options.Replace().SetUpsert(true),
	)
	return err
}

func (s *Session) Append(msg Message) error {
	s.SessionPart.Messages = append(s.SessionPart.Messages, msg)
	return s.Save()
}

func NewSession(ctx context.Context, sid string, mdb *mongodb.MongoClient) (*Session, error) {
	s := &Session{
		SessionID:   sid,
		mongoClient: mdb,
	}
	var sp SessionPart
	if len(sid) != 36 {
		sp = createNewSessionPart()
	} else {
		result := SessionPart{}
		filter := bson.D{{
			Key:   "session_id",
			Value: sid,
		}}

		err := s.mongoClient.Database("novelflow").Collection("sessions").FindOne(ctx, filter).Decode(&result)
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no such session")
		}
		if err != nil {
			return nil, fmt.Errorf("error: %v", err)
		}
		sp = result
	}
	s.SessionPart = sp
	if err := s.Save(); err != nil {
		return nil, fmt.Errorf("failed to save session: %v", err)
	}
	return s, nil
}

func createNewSessionPart() SessionPart {
	sid := uuid.New().String()
	return SessionPart{
		SID: sid,
	}
}
