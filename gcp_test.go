package propagationgcp_test

import (
	"context"
	"github.com/amammay/propagationgcp"
	"net/http/httptest"
	"reflect"
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

func TestHTTPFormatExtract(t *testing.T) {

	req1 := httptest.NewRequest("GET", "http://example.com", nil)
	req1.Header.Set("X-Cloud-Trace-Context", "a0d3eee13de6a4bbcf291eb444b94f28/999;o=1")

	// Tests the extract funcation
	hf := propagationgcp.HTTPFormat{}
	ctx := hf.Extract(context.Background(), prop.HeaderCarrier(req1.Header))

	sc := trace.SpanContextFromContext(ctx)

	if diff := reflect.DeepEqual(sc.TraceID().String(), "a0d3eee13de6a4bbcf291eb444b94f28"); !diff {
		t.Errorf("failed to traceid test: %v", diff)
	}

	if diff := reflect.DeepEqual(sc.SpanID().String(), "00000000000003e7"); !diff {
		t.Errorf("failed to spanid test: %v", diff)
	}

	if sc.TraceFlags().IsSampled()  {
		t.Errorf("failed to trace flag test")
	}
}
