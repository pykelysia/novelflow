package service

import (
	"context"
	"fmt"

	"novelflow/backend/internal/servicecontext"
	"novelflow/cache"
	"novelflow/database/mongodb"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type RuleService struct{}

func NewRuleService() *RuleService {
	return &RuleService{}
}

func (s *RuleService) CreateRule(svc *servicecontext.ServiceContext, userID uint, req *mongodb.CreateRuleRequest) (*mongodb.RuleResponse, error) {
	enabled := true
	if req.IsEnabled != nil {
		enabled = *req.IsEnabled
	}
	rule := &mongodb.Rule{
		UserID:    userID,
		Name:      req.Name,
		Content:   req.Content,
		IsEnabled: enabled,
	}
	if err := svc.RuleRepo.Create(context.Background(), rule); err != nil {
		return nil, fmt.Errorf("create rule: %w", err)
	}
	invalidateRulesCache(context.Background(), svc.RedisClient, userID)
	return rule.ToResponse(), nil
}

func (s *RuleService) ListRules(svc *servicecontext.ServiceContext, userID uint) ([]*mongodb.RuleResponse, error) {
	rules, err := svc.RuleRepo.FindByUserID(context.Background(), userID)
	if err != nil {
		return nil, fmt.Errorf("list rules: %w", err)
	}
	resp := make([]*mongodb.RuleResponse, len(rules))
	for i := range rules {
		resp[i] = rules[i].ToResponse()
	}
	return resp, nil
}

func (s *RuleService) GetRule(svc *servicecontext.ServiceContext, userID uint, idHex string) (*mongodb.RuleResponse, error) {
	id, err := bson.ObjectIDFromHex(idHex)
	if err != nil {
		return nil, ErrRuleNotFound
	}
	rule, err := svc.RuleRepo.FindByID(context.Background(), id)
	if err != nil {
		return nil, mapRuleErr(err)
	}
	if rule.UserID != userID {
		return nil, ErrRuleForbidden
	}
	return rule.ToResponse(), nil
}

func (s *RuleService) UpdateRule(svc *servicecontext.ServiceContext, userID uint, idHex string, req *mongodb.UpdateRuleRequest) (*mongodb.RuleResponse, error) {
	id, err := bson.ObjectIDFromHex(idHex)
	if err != nil {
		return nil, ErrRuleNotFound
	}
	rule, err := svc.RuleRepo.Update(context.Background(), id, userID, req)
	if err != nil {
		return nil, mapRuleErr(err)
	}
	invalidateRulesCache(context.Background(), svc.RedisClient, userID)
	return rule.ToResponse(), nil
}

func (s *RuleService) ToggleRule(svc *servicecontext.ServiceContext, userID uint, idHex string) (*mongodb.RuleResponse, error) {
	id, err := bson.ObjectIDFromHex(idHex)
	if err != nil {
		return nil, ErrRuleNotFound
	}
	rule, err := svc.RuleRepo.ToggleEnabled(context.Background(), id, userID)
	if err != nil {
		return nil, mapRuleErr(err)
	}
	invalidateRulesCache(context.Background(), svc.RedisClient, userID)
	return rule.ToResponse(), nil
}

func (s *RuleService) DeleteRule(svc *servicecontext.ServiceContext, userID uint, idHex string) error {
	id, err := bson.ObjectIDFromHex(idHex)
	if err != nil {
		return ErrRuleNotFound
	}
	if err := svc.RuleRepo.Delete(context.Background(), id, userID); err != nil {
		return mapRuleErr(err)
	}
	invalidateRulesCache(context.Background(), svc.RedisClient, userID)
	return nil
}

func (s *RuleService) GetEnabledRules(svc *servicecontext.ServiceContext, userID uint) ([]mongodb.Rule, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s%d", cache.UserRulesKeyPrefix, userID)

	var rules []mongodb.Rule
	if hit, err := svc.RedisClient.GetJSON(ctx, key, &rules); err == nil && hit {
		return rules, nil
	}

	rules, err := svc.RuleRepo.FindEnabledByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get enabled rules: %w", err)
	}
	_ = svc.RedisClient.SetJSON(ctx, key, rules, cache.RulesCacheTTL)
	return rules, nil
}

func invalidateRulesCache(ctx context.Context, rc *cache.Client, userID uint) {
	_ = rc.Del(ctx, fmt.Sprintf("%s%d", cache.UserRulesKeyPrefix, userID))
}

func mapRuleErr(err error) error {
	switch err {
	case mongodb.ErrRuleNotFound:
		return ErrRuleNotFound
	case mongodb.ErrRuleForbidden:
		return ErrRuleForbidden
	}
	return err
}
