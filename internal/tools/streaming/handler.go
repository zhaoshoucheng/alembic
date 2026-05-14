package streaming

type DeltaProgress string

const (
	DeltaInit              DeltaProgress = "init"
	DeltaStreaming          DeltaProgress = "delta"
	DeltaArgumentsFinished DeltaProgress = "argumentsFinished"
	DeltaRollback          DeltaProgress = "rollback"
	DeltaDone              DeltaProgress = "done"
)

type ToolDeltaEvent struct {
	TargetID      string
	ToolName      string
	ActionID      string
	DeltaProgress DeltaProgress
	Brief         string
	Detail        map[string]any
}

type StreamEventPusher interface {
	PushDeltaEvent(event ToolDeltaEvent) error
}

type FinishedToolInfo struct {
	ToolID   string
	EventID  string
	ToolName string
	Params   map[string]any
}

type StreamHandler interface {
	ToolName() string
	OnToolStart(toolID string, pusher StreamEventPusher) error
	OnParamsDelta(currentParams map[string]any, pusher StreamEventPusher) error
	OnToolFinish(finalParams map[string]any, pusher StreamEventPusher) (*FinishedToolInfo, error)
	OnToolAbort(reason string, pusher StreamEventPusher) error
}
