package propagationgcp

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"strconv"
	"strings"
)

const (
	httpHeaderMaxSize = 200
	httpHeader        = `X-Cloud-Trace-Context`
)

var _ propagation.TextMapPropagator = HTTPFormat{}

// HTTPFormat implements propagation.HTTPFormat to propagate
// traces in HTTP headers for Google Cloud Platform and Stackdriver Trace.
type HTTPFormat struct{}

func (f HTTPFormat) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	sc := trace.SpanFromContext(ctx).SpanContext()
	if !sc.TraceID().IsValid() || !sc.SpanID().IsValid() {
		return
	}
	spanID := sc.SpanID()
	sid := binary.BigEndian.Uint64(spanID[:])
	header := fmt.Sprintf("%s/%d;o=%d", sc.TraceID().String(), sid, sc.TraceFlags())
	carrier.Set(httpHeader, header)
}

func (f HTTPFormat) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	if h := carrier.Get(httpHeader); h != "" {
		sc, err := extract(h)
		if err == nil && sc.IsValid() {
			return trace.ContextWithRemoteSpanContext(ctx, sc)
		}
	}
	return ctx
}

// extract is using functionality from https://github.com/glassonion1/logz/blob/main/propagation/http_format.go
func extract(h string) (trace.SpanContext, error) {
	sc := trace.SpanContext{}

	if h == "" || len(h) > httpHeaderMaxSize {
		return trace.SpanContext{}, fmt.Errorf("header over max size")
	}

	// Parse the trace id field.
	slash := strings.Index(h, `/`)
	if slash == -1 {
		return sc, errors.New("failed to parse value")
	}
	tid, h := h[:slash], h[slash+1:]

	traceID, err := trace.TraceIDFromHex(tid)
	if err != nil {
		return sc, fmt.Errorf("failed to parse value: %w", err)
	}

	sc = sc.WithTraceID(traceID)

	// Parse the span id field.
	spanstr := h
	semicolon := strings.Index(h, `;`)
	if semicolon != -1 {
		spanstr, h = h[:semicolon], h[semicolon+1:]
	}
	sid, err := strconv.ParseUint(spanstr, 10, 64)
	if err != nil {
		return sc, fmt.Errorf("failed to parse value: %w", err)
	}
	spanID := sc.SpanID()
	binary.BigEndian.PutUint64(spanID[:], sid)
	sc = sc.WithSpanID(spanID)

	// Parse the options field, options field is optional.
	if !strings.HasPrefix(h, "o=") {
		return sc, errors.New("failed to parse value")
	}

	o, err := strconv.ParseUint(h[2:], 10, 64)
	if err != nil {
		return sc, fmt.Errorf("failed to parse value: %w", err)
	}

	// 1 = to sample
	if o == 1 {
		sc = sc.WithTraceFlags(trace.FlagsSampled)
	}
	return sc, nil
}

func (f HTTPFormat) Fields() []string {
	return []string{httpHeader}
}
