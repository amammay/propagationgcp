package propagationgcp_test

import (
	"context"
	"encoding/binary"
	"github.com/amammay/propagationgcp"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	"go.opentelemetry.io/otel/oteltest"
	"go.opentelemetry.io/otel/trace"

	prop "go.opentelemetry.io/otel/propagation"
)

func TestHTTPFormatInject(t *testing.T) {

	mockTracer := oteltest.DefaultTracer()
	ctx, _ := mockTracer.Start(context.Background(), "inject")

	req1 := httptest.NewRequest("GET", "http://example.com", nil)

	// Tests the inject funcation
	hf := propagationgcp.HTTPFormat{}
	hf.Inject(ctx, prop.HeaderCarrier(req1.Header))

	want := "00000000000000020000000000000000/2;o=0"
	got := req1.Header.Get("X-Cloud-Trace-Context")
	if diff := reflect.DeepEqual(got, want); !diff {
		t.Errorf("failed to inject test: %v", diff)
	}
}

func TestHTTPFormat_Extract(t *testing.T) {
	type want struct {
		traceID string
		spanID  string
		sampled bool
	}

	tests := []struct {
		traceHeader string
		want        want
	}{
		{
			traceHeader: "a0d3eee13de6a4bbcf291eb444b94f28/8528140779317015234;o=1",
			want: want{
				traceID: "a0d3eee13de6a4bbcf291eb444b94f28",
				spanID:  "8528140779317015234",
				sampled: true,
			},
		},
		{
			traceHeader: "a0d3eee13de6a4bbcf291eb444b94f28/8528140779317015234;o=0",
			want: want{
				traceID: "a0d3eee13de6a4bbcf291eb444b94f28",
				spanID:  "8528140779317015234",
				sampled: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.traceHeader, func(t *testing.T) {
			httpFormat := propagationgcp.HTTPFormat{}
			req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
			req.Header.Add("X-Cloud-Trace-Context", tt.traceHeader)
			ctx := httpFormat.Extract(context.Background(), prop.HeaderCarrier(req.Header))
			sc := trace.SpanContextFromContext(ctx)
			if diff := reflect.DeepEqual(sc.TraceID().String(), tt.want.traceID); !diff {
				t.Errorf("failed to traceid test: %v", diff)
			}

			s := sc.SpanID()
			data := binary.BigEndian.Uint64(s[:])

			if diff := reflect.DeepEqual(strconv.Itoa(int(data)), tt.want.spanID); !diff {
				t.Errorf("failed to spanid test: %v", diff)
			}

			if diff := reflect.DeepEqual(sc.IsSampled(), tt.want.sampled); !diff {
				t.Errorf("failed to trace flag test")
			}

		})
	}
}
