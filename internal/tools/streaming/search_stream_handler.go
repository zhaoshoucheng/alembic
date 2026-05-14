package streaming

type SearchStreamHandler struct {
	BaseStreamHandler
}

func NewSearchStreamHandler(toolName string) *SearchStreamHandler {
	return &SearchStreamHandler{BaseStreamHandler: NewBaseStreamHandler(toolName)}
}

func (h *SearchStreamHandler) OnToolStart(toolID string, pusher StreamEventPusher) error {
	h.start(toolID)
	return pusher.PushDeltaEvent(ToolDeltaEvent{
		TargetID:      h.eventID,
		ToolName:      h.name,
		ActionID:      toolID,
		DeltaProgress: DeltaInit,
	})
}

func (h *SearchStreamHandler) OnParamsDelta(params map[string]any, pusher StreamEventPusher) error {
	h.mergeParams(params)
	brief, _ := params["brief"].(string)
	query, _ := params["query"].(string)
	return pusher.PushDeltaEvent(ToolDeltaEvent{
		TargetID:      h.eventID,
		ToolName:      h.name,
		ActionID:      h.toolID,
		DeltaProgress: DeltaStreaming,
		Brief:         brief,
		Detail: map[string]any{
			"search": map[string]any{
				"query": query,
			},
		},
	})
}

func (h *SearchStreamHandler) OnToolFinish(params map[string]any, pusher StreamEventPusher) (*FinishedToolInfo, error) {
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

func (h *SearchStreamHandler) OnToolAbort(reason string, pusher StreamEventPusher) error {
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
