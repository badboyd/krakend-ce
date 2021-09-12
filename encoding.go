package krakend

import (
	// muxxml "github.com/badboyd/krakend-xml/mux"
	rss "github.com/devopsfaith/krakend-rss"
	xml "github.com/devopsfaith/krakend-xml"
	// "github.com/luraproject/lura/router/mux"
)

// RegisterEncoders registers all the available encoders
func RegisterEncoders() {
	xml.Register()
	rss.Register()

	// mux.RegisterRender()
	// mux.RegisterRender(xml.Name, muxxml.Render)
}
