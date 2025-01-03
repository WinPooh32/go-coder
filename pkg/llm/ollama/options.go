package ollama

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ollama/ollama/api"
)

type options struct {
	ollamaOptions Options
	httpClient    *http.Client
	keepAlive     *api.Duration
	format        json.RawMessage
}

type Option func(*options)

func WithOllamaOptions(ollamaOptions Options) Option {
	return func(opts *options) {
		opts.ollamaOptions = ollamaOptions
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(opts *options) {
		opts.httpClient = client
	}
}

func WithKeepAlive(keepAlive *api.Duration) Option {
	return func(opts *options) {
		opts.keepAlive = keepAlive
	}
}

func WithFormat(format json.RawMessage) Option {
	return func(opts *options) {
		opts.format = format
	}
}

type Runner struct {
	NumCtx             int     `json:"num_ctx,omitempty"`
	NumBatch           int     `json:"num_batch,omitempty"`
	NumGQA             int     `json:"num_gqa,omitempty"`
	NumGPU             int     `json:"num_gpu,omitempty"`
	MainGPU            int     `json:"main_gpu,omitempty"`
	NumThread          int     `json:"num_thread,omitempty"`
	RopeFrequencyBase  float32 `json:"rope_frequency_base,omitempty"`
	RopeFrequencyScale float32 `json:"rope_frequency_scale,omitempty"`
	LogitsAll          bool    `json:"logits_all,omitempty"`
	VocabOnly          bool    `json:"vocab_only,omitempty"`
	UseMMap            bool    `json:"use_mmap,omitempty"`
	UseMLock           bool    `json:"use_mlock,omitempty"`
	EmbeddingOnly      bool    `json:"embedding_only,omitempty"`
	UseNUMA            bool    `json:"numa,omitempty"`
	F16KV              bool    `json:"f16_kv,omitempty"`
	LowVRAM            bool    `json:"low_vram,omitempty"`
}

type Options struct {
	Stop []string `json:"stop,omitempty"`
	Runner
	RepeatLastN      int     `json:"repeat_last_n,omitempty"`
	Seed             int     `json:"seed,omitempty"`
	TopK             int     `json:"top_k,omitempty"`
	NumKeep          int     `json:"num_keep,omitempty"`
	Mirostat         int     `json:"mirostat,omitempty"`
	NumPredict       int     `json:"num_predict,omitempty"`
	Temperature      float32 `json:"temperature"`
	TypicalP         float32 `json:"typical_p,omitempty"`
	RepeatPenalty    float32 `json:"repeat_penalty,omitempty"`
	PresencePenalty  float32 `json:"presence_penalty,omitempty"`
	FrequencyPenalty float32 `json:"frequency_penalty,omitempty"`
	TFSZ             float32 `json:"tfs_z,omitempty"`
	MirostatTau      float32 `json:"mirostat_tau,omitempty"`
	MirostatEta      float32 `json:"mirostat_eta,omitempty"`
	TopP             float32 `json:"top_p,omitempty"`
	PenalizeNewline  bool    `json:"penalize_newline,omitempty"`
}

func (o *Options) AsMapParams() (params map[string]any, err error) {
	b, err := json.Marshal(o)
	if err != nil {
		return nil, fmt.Errorf("marshal to json: %w", err)
	}

	if err := json.Unmarshal(b, &params); err != nil {
		return nil, fmt.Errorf("unmarshal from json: %w", err)
	}

	return params, nil
}
