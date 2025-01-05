package llm

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var ErrNotStopDoneReason = errors.New("reason of done is not \"stop\"")

type MessageGenerator interface {
	// Generate generates the next message of the history.
	Generate(ctx context.Context, history []Message, tools []ToolFunction) (Message, error)
}

type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

type Message struct {
	Role      Role
	Content   string
	ToolCalls []ToolCallFunction
}

type Role int

const (
	System Role = iota
	User
	Assistant
	Tool
)

func (role Role) String() string {
	s, err := role.ToString()
	if err != nil {
		return "unknown"
	}

	return s
}

func (role Role) ToString() (string, error) {
	switch role {
	case System:
		return "system", nil
	case User:
		return "user", nil
	case Assistant:
		return "assistant", nil
	case Tool:
		return "tool", nil
	default:
		return "", fmt.Errorf("unknown role %d", role)
	}
}

func RoleFromString(s string) (Role, error) {
	switch strings.ToLower(s) {
	case "system":
		return System, nil
	case "user":
		return User, nil
	case "assistant":
		return Assistant, nil
	case "tool":
		return Tool, nil
	default:
		return 0, fmt.Errorf("unknown role %q", s)
	}
}

type ToolFunction struct {
	Name        string
	Description string
	Parameters  map[string]FunctionProperty
}

type PropertyType int

const (
	String PropertyType = iota
	Number
	Integer
	Boolean
	Array
)

func (prop PropertyType) ToString() (string, error) {
	switch prop {
	case String:
		return "string", nil
	case Number:
		return "number", nil
	case Integer:
		return "integer", nil
	case Boolean:
		return "boolean", nil
	case Array:
		return "array", nil
	default:
		return "", fmt.Errorf("unknown property type %d", prop)
	}
}

type FunctionProperty struct {
	Type          PropertyType
	ArrayItemType PropertyType
	Description   string
	Enum          []string
	Required      bool
}

type ToolCallFunction struct {
	Name      string
	Arguments map[string]any
}
