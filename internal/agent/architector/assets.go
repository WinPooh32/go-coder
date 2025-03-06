package architector

import (
	"embed"
	"encoding/json"
)

var (
	//go:embed _assets/analyze*.tpl
	analyzePrompts embed.FS

	//go:embed _assets/analyze_task_schema.json
	analyzeTaskSchema json.RawMessage
)
