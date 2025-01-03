package llm

type MessageGenerator interface {
	// Generate generates the next message of the history.
	Generate(history []Message, tools []ToolFunction) (Message, error)
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

type ToolFunction struct {
	Name        string
	Description string
	Parameters  map[string]PropertyType
}

type PropertyType int

const (
	String PropertyType = iota
	Number
	Integer
	Boolean
	Array
)

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
