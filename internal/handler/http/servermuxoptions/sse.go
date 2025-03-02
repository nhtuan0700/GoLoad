package servermuxoptions

import (
	"errors"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type eventSourceMarshaler struct {
	runtime.JSONPb
}

func (m *eventSourceMarshaler) ContentType(v any) string {
	return "text/event-stream"
}

func (m *eventSourceMarshaler) Marshal(v any) ([]byte, error) {
	b, err := m.JSONPb.Marshal(v)
	if err != nil {
		return nil, err
	}

	return []byte(fmt.Sprintf("data: %s\n", string(b))), nil
}

func (m *eventSourceMarshaler) Unmarshal(data []byte, v any) error {
	return errors.New("text/event-stream unmarshaler is unsupported")
}

func WithSSE() runtime.ServeMuxOption {
	return runtime.WithMarshalerOption("text/event-stream", &eventSourceMarshaler{JSONPb: runtime.JSONPb{}})
}
