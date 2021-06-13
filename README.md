
## Installation

```
$ go get -u github.com/amammay/propagationgcp
```

## Usage

```go
package main

import (
	"context"
	"github.com/amammay/propagationgcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func initTracing() {
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
		propagationgcp.HTTPFormat{},
	)
	otel.SetTextMapPropagator(propagator)
}

func someMethod(ctx context.Context) {
	sc := trace.SpanContextFromContext(ctx)
	traceID := sc.TraceID().String()
	spanID := sc.SpanID().String()
	isSampled := sc.IsSampled()
	_ = traceID
	_ = spanID
	_ = isSampled
}

```
