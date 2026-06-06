package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const rulesCollection = "rules"

var ErrRuleNotFound = errors.New("rule not found")

type Rule struct {
	ID        bson.ObjectID `json:"id"         bson:"_id,omitempty"`
	UserID    uint          `json:"user_id"    bson:"user_id"`
	Name      string        `json:"name"       bson:"name"`
	Content   string        `json:"content"    bson:"content"`
	IsEnabled bool          `json:"is_enabled" bson:"is_enabled"`
	CreatedAt time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time     `json:"updated_at" bson:"updated_at"`
}

type CreateRuleRequest struct {
	Name      string `json:"name"       binding:"required,max=100"`
	Content   string `json:"content"    binding:"required"`
	IsEnabled *bool  `json:"is_enabled"`
}

type UpdateRuleRequest struct {
	Name    string `json:"name"    binding:"omitempty,max=100"`
	Content string `json:"content" binding:"omitempty"`
}

type RuleResponse struct {
	ID        string    `json:"id"`
	UserID    uint      `json:"user_id"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	IsEnabled bool      `json:"is_enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (r *Rule) ToResponse() *RuleResponse {
	return &RuleResponse{
		ID:        r.ID.Hex(),
		UserID:    r.UserID,
		Name:      r.Name,
		Content:   r.Content,
		IsEnabled: r.IsEnabled,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

type RuleRepository struct {
	col *mongo.Collection
}

func NewRuleRepository(client *MongoClient) *RuleRepository {
	col := client.Database("novelflow").Collection(rulesCollection)
	return &RuleRepository{col: col}
}

func (r *RuleRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "is_enabled", Value: 1}},
	})
	return err
}

func (r *RuleRepository) Create(ctx context.Context, rule *Rule) error {
	rule.ID = bson.NewObjectID()
	now := time.Now()
	rule.CreatedAt = now
	rule.UpdatedAt = now
	_, err := r.col.InsertOne(ctx, rule)
	return err
}

func (r *RuleRepository) FindByID(ctx context.Context, id bson.ObjectID) (*Rule, error) {
	var rule Rule
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&rule)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrRuleNotFound
	}
	return &rule, err
}

func (r *RuleRepository) FindByUserID(ctx context.Context, userID uint) ([]Rule, error) {
	cursor, err := r.col.Find(ctx, bson.M{"user_id": userID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}))
	if err != nil {
		return nil, err
	}
	var rules []Rule
	if err := cursor.All(ctx, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *RuleRepository) FindEnabledByUserID(ctx context.Context, userID uint) ([]Rule, error) {
	cursor, err := r.col.Find(ctx, bson.M{"user_id": userID, "is_enabled": true},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}))
	if err != nil {
		return nil, err
	}
	var rules []Rule
	if err := cursor.All(ctx, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *RuleRepository) Update(ctx context.Context, id bson.ObjectID, userID uint, req *UpdateRuleRequest) (*Rule, error) {
	update := bson.M{"$set": bson.M{"updated_at": time.Now()}}
	if req.Name != "" {
		update["$set"].(bson.M)["name"] = req.Name
	}
	if req.Content != "" {
		update["$set"].(bson.M)["content"] = req.Content
	}
	filter := bson.M{"_id": id, "user_id": userID}
	res, err := r.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}
	if res.MatchedCount == 0 {
		// 区分 not found 与 forbidden
		if _, ferr := r.FindByID(ctx, id); errors.Is(ferr, ErrRuleNotFound) {
			return nil, ErrRuleNotFound
		}
		return nil, ErrRuleForbidden
	}
	return r.FindByID(ctx, id)
}

func (r *RuleRepository) ToggleEnabled(ctx context.Context, id bson.ObjectID, userID uint) (*Rule, error) {
	rule, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rule.UserID != userID {
		return nil, ErrRuleForbidden
	}
	update := bson.M{"$set": bson.M{"is_enabled": !rule.IsEnabled, "updated_at": time.Now()}}
	if _, err := r.col.UpdateOne(ctx, bson.M{"_id": id}, update); err != nil {
		return nil, err
	}
	rule.IsEnabled = !rule.IsEnabled
	return rule, nil
}

func (r *RuleRepository) Delete(ctx context.Context, id bson.ObjectID, userID uint) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id, "user_id": userID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		if _, ferr := r.FindByID(ctx, id); errors.Is(ferr, ErrRuleNotFound) {
			return ErrRuleNotFound
		}
		return ErrRuleForbidden
	}
	return nil
}

var ErrRuleForbidden = errors.New("rule does not belong to user")
