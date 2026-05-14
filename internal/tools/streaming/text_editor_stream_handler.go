package streaming

type TextEditorStreamHandler struct {
	BaseStreamHandler
}

func NewTextEditorStreamHandler(toolName string) *TextEditorStreamHandler {
	return &TextEditorStreamHandler{BaseStreamHandler: NewBaseStreamHandler(toolName)}
}

func (h *TextEditorStreamHandler) OnToolStart(toolID string, pusher StreamEventPusher) error {
	h.start(toolID)
	return pusher.PushDeltaEvent(ToolDeltaEvent{
		TargetID:      h.eventID,
		ToolName:      h.name,
		ActionID:      toolID,
		DeltaProgress: DeltaInit,
	})
}

func (h *TextEditorStreamHandler) OnParamsDelta(params map[string]any, pusher StreamEventPusher) error {
	h.mergeParams(params)
	brief, _ := params["brief"].(string)
	path, _ := params["path"].(string)
	content, _ := params["file_text"].(string)
	if content == "" {
		content, _ = params["new_str"].(string)
	}
	action := "view"
	if cmd, ok := params["command"].(string); ok {
		action = cmd
	}
	return pusher.PushDeltaEvent(ToolDeltaEvent{
		TargetID:      h.eventID,
		ToolName:      h.name,
		ActionID:      h.toolID,
		DeltaProgress: DeltaStreaming,
		Brief:         brief,
		Detail: map[string]any{
			"textEditor": map[string]any{
				"action":  action,
				"path":    path,
				"content": content,
			},
		},
	})
}

func (h *TextEditorStreamHandler) OnToolFinish(params map[string]any, pusher StreamEventPusher) (*FinishedToolInfo, error) {
	h.mergeParams(params)
	err := pusher.PushDeltaEvent(ToolDeltaEvent{
		TargetID:      h.eventID,
		ToolName:      h.name,
		ActionID:      h.toolID,
		DeltaProgress: DeltaArgumentsFinished,
	})
	if err != nil {
		return nil, err
	}
	info := h.finishedInfo(params)
	h.reset()
	return info, nil
}

func (h *TextEditorStreamHandler) OnToolAbort(reason string, pusher StreamEventPusher) error {
	err := pusher.PushDeltaEvent(ToolDeltaEvent{
		TargetID:      h.eventID,
		ToolName:      h.name,
		ActionID:      h.toolID,
		DeltaProgress: DeltaRollback,
		Brief:         reason,
	})
	h.reset()
	return err
}
