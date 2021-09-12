package krakend

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	botdetector "github.com/badboyd/krakend-botdetector/mux"
	httpsecure "github.com/devopsfaith/krakend-httpsecure/mux"
	lua "github.com/devopsfaith/krakend-lua/router/mux"

	gorilla "github.com/gorilla/mux"

	"github.com/luraproject/lura/config"
	"github.com/luraproject/lura/logging"
	"github.com/luraproject/lura/router/mux"
)

type gMux struct {
	Router *gorilla.Router
}

func (g *gMux) Handle(pattern, method string, handler http.Handler) {
	newRoute := g.Router.HandleFunc(pattern, handler.ServeHTTP)
	if method != "" {
		newRoute.Methods(method)
	}
}

func (g *gMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.Router.ServeHTTP(mux.NewHTTPErrorInterceptor(w), r)
}

// NewEngine creates a new gin engine with some default values and a secure middleware
func NewMuxEngine(cfg config.ServiceConfig, logger logging.Logger, w io.Writer) *gMux {
	router := gorilla.NewRouter()
	router.StrictSlash(true)

	router.Use(LoggerWithConfig(LogConfig{
		Output: os.Stdout,
	}))

	router.Use(httpsecure.NewSecureMw(cfg.ExtraConfig).Handler)

	mw := []mux.HandlerMiddleware{}
	lua.RegisterMiddleware(logger, cfg.ExtraConfig, paramExtractor, mw)
	if len(mw) > 0 {
		router.Use(mw[0].Handler)
	}

	if bot := botdetector.NewMiddleware(cfg, logger); bot != nil {
		router.Use(bot.Handler)
	}

	return &gMux{
		Router: router,
	}
}

type muxEngineFactory struct{}

func (e muxEngineFactory) NewEngine(cfg config.ServiceConfig, l logging.Logger, w io.Writer) *gMux {
	return NewMuxEngine(cfg, l, w)
}

const (
	green   = "\033[97;42m"
	white   = "\033[90;47m"
	yellow  = "\033[90;43m"
	red     = "\033[97;41m"
	blue    = "\033[97;44m"
	magenta = "\033[97;45m"
	cyan    = "\033[97;46m"
	reset   = "\033[0m"
)

type LogConfig struct {
	Output io.Writer
}

type LogFormatterParams struct {
	Request *http.Request

	// TimeStamp shows the time after the server returns a response.
	TimeStamp time.Time
	// StatusCode is HTTP response code.
	StatusCode int
	// Latency is how much time the server cost to process a certain request.
	Latency time.Duration
	// Method is the HTTP method given to the request.
	Method string
	// Path is a path the client requests.
	Path string
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(s int) {
	w.status = s
	w.ResponseWriter.WriteHeader(s)
}

func (p *LogFormatterParams) StatusCodeColor() string {
	code := p.StatusCode

	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return green
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		return white
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return yellow
	default:
		return red
	}
}

func (p *LogFormatterParams) MethodColor() string {
	method := p.Method

	switch method {
	case http.MethodGet:
		return blue
	case http.MethodPost:
		return cyan
	case http.MethodPut:
		return yellow
	case http.MethodDelete:
		return red
	case http.MethodPatch:
		return green
	case http.MethodHead:
		return magenta
	case http.MethodOptions:
		return white
	default:
		return reset
	}
}

func (p *LogFormatterParams) ResetColor() string {
	return reset
}

func Logger(next http.Handler) http.Handler {
	return LoggerWithConfig(LogConfig{})(next)
}

func LoggerWithConfig(c LogConfig) gorilla.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		out := c.Output
		if out == nil {
			out = os.Stdout
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			path := r.URL.Path
			raw := r.URL.RawQuery

			sw := statusWriter{
				ResponseWriter: w,
				status:         200,
			}
			next.ServeHTTP(&sw, r)

			if raw != "" {
				path = fmt.Sprintf("%s?%s", path, raw)
			}

			stop := time.Now()
			p := LogFormatterParams{
				Request:    r,
				TimeStamp:  stop,
				Latency:    stop.Sub(start),
				Method:     r.Method,
				StatusCode: sw.status,
				Path:       path,
			}

			fmt.Fprintf(out, formatter(p))
		})
	}
}

func formatter(p LogFormatterParams) string {
	statusColor := p.StatusCodeColor()
	methodColor := p.MethodColor()
	resetColor := p.ResetColor()

	if p.Latency > time.Minute {
		p.Latency = p.Latency - p.Latency%time.Second
	}

	return fmt.Sprintf("[MUX] %v |%s %3d %s| %13v |%s %-7s %s %s\n",
		p.TimeStamp.Format("2006/01/02 - 15:04:05"),
		statusColor, p.StatusCode, resetColor,
		p.Latency,
		methodColor, p.Method, resetColor,
		p.Path,
	)
}
