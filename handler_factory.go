package krakend

import (
	"net/http"
	"strings"

	botdetector "github.com/badboyd/krakend-botdetector/mux"
	juju "github.com/badboyd/krakend-ratelimit/juju/router/mux"
	jose "github.com/devopsfaith/krakend-jose"
	muxjose "github.com/devopsfaith/krakend-jose/mux"
	lua "github.com/devopsfaith/krakend-lua/router/mux"
	metrics "github.com/devopsfaith/krakend-metrics/mux"
	opencensus "github.com/devopsfaith/krakend-opencensus/router/mux"
	gorilla "github.com/gorilla/mux"
	"github.com/luraproject/lura/logging"

	mux "github.com/luraproject/lura/router/mux"
)

func paramExtractor(r *http.Request) map[string]string {
	params := map[string]string{}
	for key, value := range gorilla.Vars(r) {
		params[strings.Title(key)] = value
	}
	return params
}

// NewHandlerFactory returns a HandlerFactory with a rate-limit and a metrics collector middleware injected
func NewHandlerFactory(logger logging.Logger, metricCollector *metrics.Metrics, rejecter jose.RejecterFactory) mux.HandlerFactory {
	handlerFactory := mux.CustomEndpointHandler(mux.NewRequestBuilder(paramExtractor))
	handlerFactory = juju.NewRateLimiterMw(handlerFactory)
	handlerFactory = lua.HandlerFactory(logger, handlerFactory, paramExtractor)
	handlerFactory = muxjose.HandlerFactory(handlerFactory, paramExtractor, logger, rejecter)
	handlerFactory = metricCollector.NewHTTPHandlerFactory(handlerFactory)
	handlerFactory = opencensus.New(handlerFactory)
	handlerFactory = botdetector.New(handlerFactory, logger)
	return handlerFactory
}

type handlerFactory struct{}

func (h handlerFactory) NewHandlerFactory(l logging.Logger, m *metrics.Metrics, r jose.RejecterFactory) mux.HandlerFactory {
	return NewHandlerFactory(l, m, r)
}
