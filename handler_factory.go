package krakend

import (
	"net/http"

	juju "github.com/badboyd/krakend-ratelimit/juju/router/mux"
	jose "github.com/devopsfaith/krakend-jose/mux"
	lua "github.com/devopsfaith/krakend-lua/router/mux"
	metrics "github.com/devopsfaith/krakend-metrics/mux"
	opencensus "github.com/devopsfaith/krakend-opencensus/router/mux"

	"github.com/luraproject/lura/logging"
	router "github.com/luraproject/lura/router/mux"
)

func paramExtractor(r *http.Request) map[string]string {
	return nil
}

// NewHandlerFactory returns a HandlerFactory with a rate-limit and a metrics collector middleware injected
func NewHandlerFactory(logger logging.Logger, metricCollector *metrics.Metrics, rejecter jose.RejecterFactory) router.HandlerFactory {
	handlerFactory := juju.HandlerFactory
	handlerFactory = lua.HandlerFactory(logger, handlerFactory, paramExtractor)
	handlerFactory = jose.HandlerFactory(handlerFactory, logger, rejecter)
	handlerFactory = metricCollector.NewHTTPHandlerFactory(handlerFactory)
	handlerFactory = opencensus.New(handlerFactory)
	return handlerFactory
}

type handlerFactory struct{}

func (h handlerFactory) NewHandlerFactory(l logging.Logger, m *metrics.Metrics, r jose.RejecterFactory) router.HandlerFactory {
	return NewHandlerFactory(l, m, r)
}
