package client

import (
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Config struct {
	ContentServerURL string
	ServerAddr       string

	Files http.FileSystem
}

func NewRouter(cfg *Config) (*echo.Echo, error) {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	f, err := cfg.Files.Open("index.html")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	templates, err := template.New("index").Parse(string(data))
	if err != nil {
		return nil, err
	}
	e.Renderer = &TemplateRenderer{templates}

	// default route
	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", map[string]string{
			"ServerAddr": cfg.ServerAddr,
		})
	})

	raddr, err := net.ResolveUDPAddr("udp", cfg.ServerAddr)
	if err != nil {
		return nil, err
	}

	e.GET("/info", func(c echo.Context) error {
		conn, err := net.ListenPacket("udp", "0.0.0.0:0")
		if err != nil {
			return err
		}
		defer conn.Close()

		buffer := make([]byte, 1024*1024)
		if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
			return err
		}
		n, err := conn.WriteTo([]byte("\xff\xff\xff\xffgetinfo xxx"), raddr)
		if err != nil {
			return err
		}
		n, _, err = conn.ReadFrom(buffer)
		if err != nil {
			return err
		}

		resp := buffer[:n]
		resp = bytes.TrimPrefix(resp, []byte("\xff\xff\xff\xffinfoResponse\n\\"))
		resp = bytes.TrimSuffix(resp, []byte("\\xxx"))

		parts := bytes.Split(resp, []byte("\\"))

		m := make(map[string]string)
		for i := 0; i < len(parts)-1; i += 2 {
			m[string(parts[i])] = string(parts[i+1])
		}
		return c.JSON(http.StatusOK, m)
	})

	// static files
	e.GET("/*", echo.WrapHandler(http.FileServer(cfg.Files)))

	// Quake3 assets requests must be proxied to the content server. The host
	// header is manipulated to ensure that services like CloudFlare will not
	// reject requests based upon incorrect host header.
	csurl, err := url.Parse(cfg.ContentServerURL)
	if err != nil {
		return nil, err
	}
	g := e.Group("/assets")
	g.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{
		Balancer: middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{
			{URL: csurl},
		}),
		Transport: &HostHeaderTransport{RoundTripper: http.DefaultTransport, Host: csurl.Host},
	}))
	return e, nil
}

type HostHeaderTransport struct {
	http.RoundTripper
	Host string
}

func (t *HostHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Host = t.Host
	return t.RoundTripper.RoundTrip(req)
}

type TemplateRenderer struct {
	*template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.ExecuteTemplate(w, name, data)
}
