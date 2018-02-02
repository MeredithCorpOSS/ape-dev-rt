// Copyright 2014 Unknwon
// Copyright 2014 Torkel Ödegaard

package cmd

import (
	"fmt"
	"net/http"
	"path"

	"github.com/Unknwon/macaron"

	"github.com/grafana/grafana/pkg/api"
	"github.com/grafana/grafana/pkg/api/static"
	"github.com/grafana/grafana/pkg/log"
	"github.com/grafana/grafana/pkg/middleware"
	"github.com/grafana/grafana/pkg/plugins"
	"github.com/grafana/grafana/pkg/setting"
)

func newMacaron() *macaron.Macaron {
	macaron.Env = setting.Env
	m := macaron.New()

	m.Use(middleware.Logger())
	m.Use(macaron.Recovery())

	if setting.EnableGzip {
		m.Use(middleware.Gziper())
	}

	for _, route := range plugins.StaticRoutes {
		pluginRoute := path.Join("/public/plugins/", route.Url)
		log.Info("Plugin: Adding static route %s -> %s", pluginRoute, route.Path)
		mapStatic(m, route.Path, "", pluginRoute)
	}

	mapStatic(m, setting.StaticRootPath, "", "public")
	mapStatic(m, setting.StaticRootPath, "app", "app")
	mapStatic(m, setting.StaticRootPath, "css", "css")
	mapStatic(m, setting.StaticRootPath, "img", "img")
	mapStatic(m, setting.StaticRootPath, "fonts", "fonts")
	mapStatic(m, setting.StaticRootPath, "robots.txt", "robots.txt")

	m.Use(macaron.Renderer(macaron.RenderOptions{
		Directory:  path.Join(setting.StaticRootPath, "views"),
		IndentJSON: macaron.Env != macaron.PROD,
		Delims:     macaron.Delims{Left: "[[", Right: "]]"},
	}))

	if setting.EnforceDomain {
		m.Use(middleware.ValidateHostHeader(setting.Domain))
	}

	m.Use(middleware.GetContextHandler())
	m.Use(middleware.Sessioner(&setting.SessionOptions))

	return m
}

func mapStatic(m *macaron.Macaron, rootDir string, dir string, prefix string) {
	headers := func(c *macaron.Context) {
		c.Resp.Header().Set("Cache-Control", "public, max-age=3600")
	}

	if setting.Env == setting.DEV {
		headers = func(c *macaron.Context) {
			c.Resp.Header().Set("Cache-Control", "max-age=0, must-revalidate, no-cache")
		}
	}

	m.Use(httpstatic.Static(
		path.Join(rootDir, dir),
		httpstatic.StaticOptions{
			SkipLogging: true,
			Prefix:      prefix,
			AddHeaders:  headers,
		},
	))
}

func StartServer() {

	var err error
	m := newMacaron()
	api.Register(m)

	listenAddr := fmt.Sprintf("%s:%s", setting.HttpAddr, setting.HttpPort)
	log.Info("Listen: %v://%s%s", setting.Protocol, listenAddr, setting.AppSubUrl)
	switch setting.Protocol {
	case setting.HTTP:
		err = http.ListenAndServe(listenAddr, m)
	case setting.HTTPS:
		err = http.ListenAndServeTLS(listenAddr, setting.CertFile, setting.KeyFile, m)
	default:
		log.Fatal(4, "Invalid protocol: %s", setting.Protocol)
	}

	if err != nil {
		log.Fatal(4, "Fail to start server: %v", err)
	}
}
