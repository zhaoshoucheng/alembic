package streaming

import "github.com/google/uuid"

type BaseStreamHandler struct {
	name     string
	toolID   string
	eventID  string
	params   map[string]any
}

func NewBaseStreamHandler(name string) BaseStreamHandler {
	return BaseStreamHandler{name: name}
}

func (h *BaseStreamHandler) ToolName() string { return h.name }

func (h *BaseStreamHandler) start(toolID string) string {
	h.toolID = toolID
	h.eventID = uuid.NewString()[:8]
	h.params = nil
	return h.eventID
}

func (h *BaseStreamHandler) mergeParams(params map[string]any) {
	h.params = params
}

func (h *BaseStreamHandler) reset() {
	h.toolID = ""
	h.eventID = ""
	h.params = nil
}

func (h *BaseStreamHandler) finishedInfo(params map[string]any) *FinishedToolInfo {
	return &FinishedToolInfo{
		ToolID:   h.toolID,
		EventID:  h.eventID,
		ToolName: h.name,
		Params:   params,
	}
}
