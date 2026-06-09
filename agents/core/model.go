package core

import (
	"context"
	"errors"

	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/spf13/viper"
)

func GetChatModel(ctx context.Context) (model.BaseChatModel, error) {
	isOpenAI := viper.GetString("llm.model_type")
	baseurl := viper.GetString("llm.base_url")
	modelname := viper.GetString("llm.model_name")
	apiKey := viper.GetString("llm.api_key")
	max_tokens := viper.GetInt("llm.max_tokens")

	return getModel(ctx, isOpenAI, baseurl, modelname, apiKey, max_tokens)
}

func GetLiteChatModel(ctx context.Context) (model.BaseChatModel, error) {
	isOpenAI := viper.GetString("lite_llm.model_type")
	baseurl := viper.GetString("lite_llm.base_url")
	modelname := viper.GetString("lite_llm.model_name")
	apiKey := viper.GetString("lite_llm.api_key")
	max_tokens := viper.GetInt("lite_llm.max_tokens")

	return getModel(ctx, isOpenAI, baseurl, modelname, apiKey, max_tokens)
}

func getModel(ctx context.Context, flag, baseurl, modelname, apikey string, max_tokens int) (model.BaseChatModel, error) {
	switch flag {
	case "openai":
		cm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
			Model:     modelname,
			BaseURL:   baseurl,
			APIKey:    apikey,
			MaxTokens: &max_tokens,
		})
		if err != nil {
			return nil, err
		}
		return cm, nil
	case "anthropic":
		cm, err := claude.NewChatModel(ctx, &claude.Config{
			Model:     modelname,
			BaseURL:   &baseurl,
			APIKey:    apikey,
			MaxTokens: max_tokens,
		})
		if err != nil {
			return nil, err
		}
		return cm, nil
	case "deepseek":
		cm, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
			Model:     modelname,
			BaseURL:   baseurl,
			APIKey:    apikey,
			MaxTokens: max_tokens,
		})
		if err != nil {
			return nil, err
		}
		return cm, nil
	default:
		return nil, errors.New("please use right model type.")
	}
}
