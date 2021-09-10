package krakend

import (
	"io"

	httpsecure "github.com/devopsfaith/krakend-httpsecure/mux"
	"github.com/gorilla/mux"
	"github.com/luraproject/lura/config"
	"github.com/luraproject/lura/logging"
)

// NewEngine creates a new gin engine with some default values and a secure middleware
func NewMuxEngine(cfg config.ServiceConfig, logger logging.Logger, w io.Writer) *mux.Router {

	engine := mux.NewRouter()

	httpsecure.NewSecureMw(cfg.ExtraConfig)

	// lua.RegisterMiddleware(logger, cfg, nil, nil)


	return engine
}

type muxEngineFactory struct{}

func (e muxEngineFactory) NewEngine(cfg config.ServiceConfig, l logging.Logger, w io.Writer) *mux.Router {
	return NewMuxEngine(cfg, l, w)
}
