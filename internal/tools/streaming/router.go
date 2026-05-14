package streaming

import "fmt"

type DeltaToolCall struct {
	ID       string
	Name     string
	ArgChunk string
}

type streamState struct {
	toolID   string
	toolName string
	handler  StreamHandler
}

type StreamingRouter struct {
	handlers      map[string]StreamHandler
	parser        *JsonStreamParser
	current       *streamState
	FinishedTools []FinishedToolInfo
}

func NewStreamingRouter() *StreamingRouter {
	return &StreamingRouter{
		handlers: make(map[string]StreamHandler),
		parser:   NewJsonStreamParser(),
	}
}

func (r *StreamingRouter) Register(handler StreamHandler) *StreamingRouter {
	r.handlers[handler.ToolName()] = handler
	return r
}

func (r *StreamingRouter) ProcessChunk(call DeltaToolCall, pusher StreamEventPusher) error {
	if call.Name != "" && (r.current == nil || call.ID != r.current.toolID) {
		if err := r.finishCurrent(pusher); err != nil {
			return err
		}
		handler, ok := r.handlers[call.Name]
		if !ok {
			r.current = nil
			return nil
		}
		r.parser.Reset()
		r.current = &streamState{
			toolID:   call.ID,
			toolName: call.Name,
			handler:  handler,
		}
		if err := handler.OnToolStart(call.ID, pusher); err != nil {
			return err
		}
	}

	if call.ArgChunk != "" && r.current != nil {
		result := r.parser.Parse(call.ArgChunk)
		if result.Changed && result.Value != nil {
			if err := r.current.handler.OnParamsDelta(result.Value, pusher); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *StreamingRouter) FinishCurrentTool(pusher StreamEventPusher) error {
	return r.finishCurrent(pusher)
}

func (r *StreamingRouter) Abort(reason string, pusher StreamEventPusher) error {
	if r.current == nil {
		return nil
	}
	err := r.current.handler.OnToolAbort(reason, pusher)
	r.current = nil
	r.parser.Reset()
	return err
}

func (r *StreamingRouter) finishCurrent(pusher StreamEventPusher) error {
	if r.current == nil {
		return nil
	}
	finalParams := r.parser.CurrentValue()
	if finalParams == nil {
		finalParams = map[string]any{}
	}
	info, err := r.current.handler.OnToolFinish(finalParams, pusher)
	if err != nil {
		return fmt.Errorf("finish tool %s: %w", r.current.toolName, err)
	}
	if info != nil {
		r.FinishedTools = append(r.FinishedTools, *info)
	}
	r.current = nil
	r.parser.Reset()
	return nil
}

func (r *StreamingRouter) Reset() {
	r.current = nil
	r.parser.Reset()
	r.FinishedTools = nil
}
