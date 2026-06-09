package agent

import (
	"context"
	"fmt"
	"novelflow/database/mongodb"
	"strings"
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
	UserID    uint      `bson:"user_id,omitempty"`
	Title     string    `bson:"title,omitempty"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// MessageStore abstracts session message persistence, allowing the middleware
// layer to remain independent of the concrete storage backend.
type MessageStore interface {
	Append(msg Message) error
	Load() []adk.Message
}

type Session struct {
	SessionPart SessionPart
	mongoClient *mongodb.MongoClient
}

func NewSession(ctx context.Context, sid string, userID uint, mdb *mongodb.MongoClient) (*Session, error) {
	s := &Session{
		mongoClient: mdb,
	}
	if len(sid) != 36 {
		sp := createNewSessionPart()
		sp.UserID = userID
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
			sp.UserID = userID
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

func (s *Session) Load() (msgs []adk.Message) {
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

	toolCallIdx := 0
	for _, m := range messages {
		switch m.Type {
		case ContentType:
			msgs = append(msgs, &schema.Message{
				Role:    m.Role,
				Content: m.Content,
			})
		case ToolResultType:
			if m.Role == schema.Assistant && m.Content != "" {
				parts := strings.SplitN(m.Content, "\n", 2)
				toolName := parts[0]
				args := ""
				if len(parts) > 1 {
					args = parts[1]
				}
				toolCallID := fmt.Sprintf("call_%d", toolCallIdx)
				toolCallIdx++
				msgs = append(msgs, &schema.Message{
					Role: schema.Assistant,
					ToolCalls: []schema.ToolCall{
						{
							ID:   toolCallID,
							Type: "function",
							Function: schema.FunctionCall{
								Name:      toolName,
								Arguments: args,
							},
						},
					},
				})
			} else if m.Role == schema.Tool && m.ToolResult != "" {
				toolCallID := ""
				if len(msgs) > 0 {
					lastMsg := msgs[len(msgs)-1]
					if len(lastMsg.ToolCalls) > 0 {
						toolCallID = lastMsg.ToolCalls[0].ID
					}
				}
				msgs = append(msgs, &schema.Message{
					Role:       schema.Tool,
					Content:    m.ToolResult,
					ToolCallID: toolCallID,
					ToolName:   m.Content,
				})
			}
		}
	}
	return msgs
}
