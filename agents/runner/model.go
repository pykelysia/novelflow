package runner

import (
	"context"

	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/spf13/viper"
)

func getChatModel(ctx context.Context) (model.BaseChatModel, error) {
	isOpenAI := viper.GetString("llm.model_type")
	baseurl := viper.GetString("llm.base_url")
	modelname := viper.GetString("llm.model_name")
	apiKey := viper.GetString("llm.api_key")
	if isOpenAI == "openai" {
		cm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
			Model:   modelname,
			BaseURL: baseurl,
			APIKey:  apiKey,
		})
		if err != nil {
			return nil, err
		}
		return cm, nil
	} else {
		cm, err := claude.NewChatModel(ctx, &claude.Config{
			Model:   modelname,
			BaseURL: &baseurl,
			APIKey:  apiKey,
		})
		if err != nil {
			return nil, err
		}
		return cm, nil
	}
}

func getLiteChatModel(ctx context.Context) (model.BaseChatModel, error) {

	isOpenAI := viper.GetString("lite_llm.model_type")
	baseurl := viper.GetString("lite_llm.base_url")
	modelname := viper.GetString("lite_llm.model_name")
	apiKey := viper.GetString("lite_llm.api_key")
	if isOpenAI == "openai" {
		cm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
			Model:   modelname,
			BaseURL: baseurl,
			APIKey:  apiKey,
		})
		if err != nil {
			return nil, err
		}
		return cm, nil
	} else {
		cm, err := claude.NewChatModel(ctx, &claude.Config{
			Model:   modelname,
			BaseURL: &baseurl,
			APIKey:  apiKey,
		})
		if err != nil {
			return nil, err
		}
		return cm, nil
	}
}
