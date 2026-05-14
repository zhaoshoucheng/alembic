package streaming

type TerminalStreamHandler struct {
	BaseStreamHandler
}

func NewTerminalStreamHandler(toolName string) *TerminalStreamHandler {
	return &TerminalStreamHandler{BaseStreamHandler: NewBaseStreamHandler(toolName)}
}

func (h *TerminalStreamHandler) OnToolStart(toolID string, pusher StreamEventPusher) error {
	h.start(toolID)
	return pusher.PushDeltaEvent(ToolDeltaEvent{
		TargetID:      h.eventID,
		ToolName:      h.name,
		ActionID:      toolID,
		DeltaProgress: DeltaInit,
	})
}

func (h *TerminalStreamHandler) OnParamsDelta(params map[string]any, pusher StreamEventPusher) error {
	h.mergeParams(params)
	brief, _ := params["brief"].(string)
	command, _ := params["command"].(string)
	sessionID, _ := params["session_id"].(string)
	return pusher.PushDeltaEvent(ToolDeltaEvent{
		TargetID:      h.eventID,
		ToolName:      h.name,
		ActionID:      h.toolID,
		DeltaProgress: DeltaStreaming,
		Brief:         brief,
		Detail: map[string]any{
			"terminal": map[string]any{
				"session_id": sessionID,
				"command":    command,
			},
		},
	})
}

func (h *TerminalStreamHandler) OnToolFinish(params map[string]any, pusher StreamEventPusher) (*FinishedToolInfo, error) {
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

func (h *TerminalStreamHandler) OnToolAbort(reason string, pusher StreamEventPusher) error {
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
