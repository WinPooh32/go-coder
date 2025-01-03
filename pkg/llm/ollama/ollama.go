package ollama

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/WinPooh32/go-coder/pkg/llm"
	"github.com/ollama/ollama/api"
)

type LLM struct {
	model   string
	options options
	client  *api.Client
}

func New(serverURL *url.URL, model string, opts ...Option) *LLM {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	var httpClient *http.Client

	if o.httpClient != nil {
		httpClient = o.httpClient
	}

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	client := api.NewClient(serverURL, httpClient)

	return &LLM{
		model:   model,
		options: o,
		client:  client,
	}
}

func (ollm *LLM) Generate(ctx context.Context, history []llm.Message, tools []llm.ToolFunction) (llm.Message, error) {
	stream := false

	opts, err := ollm.options.ollamaOptions.AsMapParams()
	if err != nil {
		return llm.Message{}, fmt.Errorf("convert ollamaOptions as map params: %w", err)
	}

	reqHistory, err := convertHistoryToRequest(history)
	if err != nil {
		return llm.Message{}, fmt.Errorf("convert history to request: %w", err)
	}

	reqTools, err := convertToolsToRequest(tools)
	if err != nil {
		return llm.Message{}, fmt.Errorf("convert tools to request: %w", err)
	}

	req := &api.ChatRequest{
		Model:     ollm.model,
		Messages:  reqHistory,
		Stream:    &stream,
		Format:    ollm.options.format,
		KeepAlive: ollm.options.keepAlive,
		Tools:     reqTools,
		Options:   opts,
	}

	var msg llm.Message

	if err := ollm.client.Chat(ctx, req, func(resp api.ChatResponse) error {
		if resp.DoneReason != "stop" {
			return fmt.Errorf("%w: reason %q", llm.ErrNotStopDoneReason, resp.DoneReason)
		}

		msg, err = convertResponseToMessage(resp.Message)
		if err != nil {
			return fmt.Errorf("parse message from response: %w", err)
		}

		return nil
	}); err != nil {
		return llm.Message{}, fmt.Errorf("ollama client: chat: %w", err)
	}

	return msg, nil
}

func convertHistoryToRequest(history []llm.Message) ([]api.Message, error) {
	var apiMessages []api.Message

	for _, msg := range history {
		role, err := msg.Role.ToString()
		if err != nil {
			return nil, fmt.Errorf("role to string: %w", err)
		}

		apiMsg := api.Message{
			Role:      role,
			Content:   msg.Content,
			Images:    nil,
			ToolCalls: nil,
		}

		if len(msg.ToolCalls) > 0 {
			apiToolCalls := make([]api.ToolCall, len(msg.ToolCalls))
			for i, toolCall := range msg.ToolCalls {
				apiToolCalls[i] = api.ToolCall{
					Function: api.ToolCallFunction{
						Index:     i,
						Name:      toolCall.Name,
						Arguments: convertMapToToolCallFunctionArguments(toolCall.Arguments),
					},
				}
			}

			apiMsg.ToolCalls = apiToolCalls
		}

		apiMessages = append(apiMessages, apiMsg)
	}

	return apiMessages, nil
}

func convertMapToToolCallFunctionArguments(arguments map[string]any) api.ToolCallFunctionArguments {
	apiArgs := make(api.ToolCallFunctionArguments, len(arguments))
	for k, v := range arguments {
		apiArgs[k] = v
	}

	return apiArgs
}

type apiParameters = struct {
	Type       string                 `json:"type"`
	Required   []string               `json:"required"`
	Properties map[string]apiProperty `json:"properties"`
}

type apiProperty = struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

func convertToolsToRequest(tools []llm.ToolFunction) ([]api.Tool, error) {
	var apiTools []api.Tool

	for _, tool := range tools {
		props, err := convertFunctionProperties(tool.Parameters)
		if err != nil {
			return nil, fmt.Errorf("convert function properties: %w", err)
		}

		apiTool := api.Tool{
			Type: "function",
			Function: api.ToolFunction{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters: apiParameters{
					Type:       "object",
					Required:   getRequiredFields(tool.Parameters),
					Properties: props,
				},
			},
		}

		apiTools = append(apiTools, apiTool)
	}

	return apiTools, nil
}

func getRequiredFields(parameters map[string]llm.FunctionProperty) []string {
	var requiredFields []string

	for paramName, param := range parameters {
		if param.Required {
			requiredFields = append(requiredFields, paramName)
		}
	}

	return requiredFields
}

func convertFunctionProperties(parameters map[string]llm.FunctionProperty) (map[string]apiProperty, error) {
	properties := make(map[string]apiProperty)

	for paramName, param := range parameters {
		typ, err := param.Type.ToString()
		if err != nil {
			return nil, fmt.Errorf("convert property type: %w", err)
		}

		apiProperty := apiProperty{
			Type:        typ,
			Description: param.Description,
			Enum:        param.Enum,
		}

		properties[paramName] = apiProperty
	}

	return properties, nil
}

func convertResponseToMessage(msg api.Message) (llm.Message, error) {
	role, err := llm.RoleFromString(msg.Role)
	if err != nil {
		return llm.Message{}, fmt.Errorf("parse role from string response: %w", err)
	}

	var toolCalls []llm.ToolCallFunction

	for _, tc := range msg.ToolCalls {
		name := tc.Function.Name
		arguments := tc.Function.Arguments
		toolCalls = append(toolCalls, llm.ToolCallFunction{
			Name:      name,
			Arguments: arguments,
		})
	}

	llmMsg := llm.Message{
		Role:      role,
		Content:   msg.Content,
		ToolCalls: toolCalls,
	}

	return llmMsg, nil
}
