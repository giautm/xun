package htmx

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

var (
	client http.Client
)

func TestMain(m *testing.M) {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		if strings.HasPrefix(addr, "abc.com") {
			return net.Dial("tcp", strings.TrimPrefix(addr, "abc.com"))
		}
		return net.Dial("tcp", addr)
	}
	client = http.Client{
		Transport: tr,
	}
	os.Exit(m.Run())
}

func TestJsonViewer(t *testing.T) {

	m := http.NewServeMux()
	srv := httptest.NewServer(m)
	defer srv.Close()

	app := New(WithMux(m))
	defer app.Close()

	app.Get("/", func(c *Context) error {
		return c.View(map[string]any{"method": "GET", "num": 1})
	})

	app.Post("/", func(c *Context) error {
		return c.View(map[string]any{"method": "POST", "num": 2})
	})

	app.Put("/", func(c *Context) error {
		return c.View(map[string]any{"method": "PUT", "num": 3})
	})

	app.Delete("/", func(c *Context) error {
		return c.View(map[string]any{"method": "DELETE", "num": 4})
	})

	app.HandleFunc("GET /func", func(c *Context) error {
		return c.View(map[string]any{"method": "HandleFunc", "num": 5})
	})

	go app.Start()

	data := &struct {
		Method string `json:"method"`
		Num    int    `json:"num"`
	}{}

	req, err := http.NewRequest("GET", srv.URL+"/", nil)
	req.Header.Set("Accept", "application/json")
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)

	err = json.NewDecoder(resp.Body).Decode(data)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, "GET", data.Method)
	require.Equal(t, 1, data.Num)

	req, err = http.NewRequest("POST", srv.URL+"/", nil)
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)

	err = json.NewDecoder(resp.Body).Decode(&data)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, "POST", data.Method)
	require.Equal(t, 2, data.Num)

	req, err = http.NewRequest("PUT", srv.URL+"/", nil)
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)

	err = json.NewDecoder(resp.Body).Decode(&data)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, "PUT", data.Method)
	require.Equal(t, 3, data.Num)

	req, err = http.NewRequest("DELETE", srv.URL+"/", nil)
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)

	err = json.NewDecoder(resp.Body).Decode(&data)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, "DELETE", data.Method)
	require.Equal(t, 4, data.Num)

	req, err = http.NewRequest("GET", srv.URL+"/func", nil)
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)

	err = json.NewDecoder(resp.Body).Decode(&data)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, "HandleFunc", data.Method)
	require.Equal(t, 5, data.Num)

}

func TestStaticViewEngine(t *testing.T) {
	fsys := fstest.MapFS{
		"public/index.html": &fstest.MapFile{
			Data: []byte(`<!DOCTYPE html>
<html>	
	<head>
		<meta charset="utf-8">
		<title>Index</title>
	</head>
	<body>
		<div hx-get="/" hx-swap="innerHTML"></div>
	</body>
</html>`),
		},
		"public/home.html": &fstest.MapFile{
			Data: []byte(`<!DOCTYPE html>
<html>	
	<head>
		<meta charset="utf-8">
		<title>Home</title>
	</head>
	<body>
		<div hx-get="/home" hx-swap="innerHTML"></div>
	</body>
</html>`),
		},
		"public/admin/index.html": &fstest.MapFile{
			Data: []byte(`<!DOCTYPE html>
<html>	
	<head>
		<meta charset="utf-8">
		<title>admin/index</title>
	</head>
	<body>
		<div hx-get="/admin" hx-swap="innerHTML"></div>
	</body>
</html>`),
		},
		"public/assets/skin.css": &fstest.MapFile{
			Data: []byte(`body {
			background: red;
		}`),
		},
		"public/assets/empty.js": &fstest.MapFile{
			Data: []byte(``),
		},
		"public/assets/nil.js": &fstest.MapFile{
			Data: nil,
		},
		// test pattern with host condition
		"public/@abc.com/index.html": &fstest.MapFile{
			Data: []byte(`<!DOCTYPE html>
<html>	
	<head>
		<meta charset="utf-8">
		<title>abc.com/Index</title>
	</head>
	<body>
		<div hx-get="/" hx-swap="innerHTML"></div>
	</body>
</html>`),
		},
		"public/@abc.com/home.html": &fstest.MapFile{
			Data: []byte(`<!DOCTYPE html>
<html>	
	<head>
		<meta charset="utf-8">
		<title>abc.com/home</title>
	</head>
	<body>
		<div hx-get="/" hx-swap="innerHTML"></div>
	</body>
</html>`),
		},
		"public/@abc.com/admin/index.html": &fstest.MapFile{
			Data: []byte(`<!DOCTYPE html>
<html>	
	<head>
		<meta charset="utf-8">
		<title>abc.com/admin</title>
	</head>
	<body>
		<div hx-get="/" hx-swap="innerHTML"></div>
	</body>
</html>`),
		},
	}

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	app := New(WithMux(mux), WithFsys(fsys))

	app.Start()
	defer app.Close()

	req, err := http.NewRequest("GET", srv.URL+"/", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)

	buf, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, fsys["public/index.html"].Data, buf)

	req, err = http.NewRequest("GET", srv.URL+"/home.html", nil)
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)

	buf, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, fsys["public/home.html"].Data, buf)

	req, err = http.NewRequest("GET", srv.URL+"/assets/skin.css", nil)
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)

	buf, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, fsys["public/assets/skin.css"].Data, buf)

	req, err = http.NewRequest("GET", srv.URL+"/assets/empty.js", nil)
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)

	buf, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, 0, len(buf))

	req, err = http.NewRequest("GET", srv.URL+"/assets/nil.js", nil)
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)

	buf, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, 0, len(buf))

	host := strings.ReplaceAll(srv.URL, "127.0.0.1", "abc.com")

	req, err = http.NewRequest("GET", host+"/", nil)
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)

	buf, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, fsys["public/@abc.com/index.html"].Data, buf)

	req, err = http.NewRequest("GET", host+"/home.html", nil)
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)

	buf, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, fsys["public/@abc.com/home.html"].Data, buf)

	req, err = http.NewRequest("GET", host+"/admin", nil)
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)

	buf, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, fsys["public/@abc.com/admin/index.html"].Data, buf)

	req, err = http.NewRequest("GET", host+"/admin/", nil)
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)

	buf, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, fsys["public/@abc.com/admin/index.html"].Data, buf)

}

func TestJsonStatus(t *testing.T) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	app := New(WithMux(mux), WithViewer(&JsonViewer{}))

	app.Start()
	defer app.Close()

	app.Get("/400", func(c *Context) error {
		c.WriteStatus(http.StatusBadRequest)
		return ErrCancelled
	})

	app.Get("/401", func(c *Context) error {
		c.WriteStatus(http.StatusUnauthorized)
		return nil
	})
	app.Get("/403", func(c *Context) error {
		c.WriteStatus(http.StatusForbidden)
		return nil

	})

	app.Get("/404", func(c *Context) error {
		c.WriteStatus(http.StatusNotFound)
		return nil
	})

	app.Get("/500", func(c *Context) error {
		c.WriteStatus(http.StatusInternalServerError)
		return nil
	})

	req, err := http.NewRequest("GET", srv.URL+"/400", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()

	req, err = http.NewRequest("GET", srv.URL+"/401", nil)
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	req, err = http.NewRequest("GET", srv.URL+"/403", nil)
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()

	req, err = http.NewRequest("GET", srv.URL+"/404", nil)
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()

	req, err = http.NewRequest("GET", srv.URL+"/500", nil)
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	resp.Body.Close()

}

func TestHtmlViewEngine(t *testing.T) {

	fsys := fstest.MapFS{
		"components/footer.html": {Data: []byte("<div>footer</div>")},
		"components/header.html": {Data: []byte("<div>header</div>")},
		"layouts/main.html": {Data: []byte(`<html><body>{{ block "components/header" . }} {{end}}
<h1>main</h1>
{{ block "content" . }} {{end}}
{{ block "components/footer" . }} {{end}}
</body></html>`)},
		"layouts/admin.html": {Data: []byte(`<html><body>{{ block "components/header" . }} {{end}}
<h1>admin</h1>
{{ block "content" . }} {{end}}
{{ block "components/footer" . }} {{end}}
</body></html>`)},
		"views/user.html": {Data: []byte(`<html><body>{{ block "components/header" . }} {{end}}
<h1>user</h1>
{{ block "components/footer" . }} {{end}}
</body></html>`)},

		"pages/index.html": {Data: []byte(`<!--layout:main-->
{{ define "content" }}<div>index</div>{{ end }}`)},
		"pages/admin/index.html": {Data: []byte(`<!--layout:admin-->
{{ define "content" }}<div>admin/index</div>{{ end }}`)},

		"pages/about.html": {Data: []byte(`<html><body>{{ block "components/header" . }} {{end}}
<h1>about</h1>
{{ block "components/footer" . }} {{end}}
</body></html>`)},

		"pages/@abc.com/index.html": {Data: []byte(`<!--layout:main-->
{{ define "content" }}<div>abc.com/index</div>{{ end }}`)},
		"pages/@abc.com/admin/index.html": {Data: []byte(`<!--layout:admin-->
{{ define "content" }}<div>abc.com/admin/index</div>{{ end }}`)},
	}

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	app := New(WithMux(mux), WithFsys(fsys))

	app.Start()
	defer app.Close()

	req, err := http.NewRequest("GET", srv.URL+"/index", nil)
	req.Header.Set("Accept", "text/html, */*")
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)

	buf, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, `<html><body><div>header</div>
<h1>main</h1>
<div>index</div>
<div>footer</div>
</body></html>`, string(buf))

	req, err = http.NewRequest("GET", srv.URL+"/admin/index", nil)
	req.Header.Set("Accept", "text/html, */*")
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)

	buf, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, `<html><body><div>header</div>
<h1>admin</h1>
<div>admin/index</div>
<div>footer</div>
</body></html>`, string(buf))

	req, err = http.NewRequest("GET", srv.URL+"/about", nil)
	req.Header.Set("Accept", "text/html, */*")
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)

	buf, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, `<html><body><div>header</div>
<h1>about</h1>
<div>footer</div>
</body></html>`, string(buf))

	req, err = http.NewRequest("GET", srv.URL+"/user", nil)
	req.Header.Set("Accept", "text/html, */*")
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()

	host := strings.ReplaceAll(srv.URL, "127.0.0.1", "abc.com")

	req, err = http.NewRequest("GET", host+"/index", nil)
	req.Header.Set("Accept", "text/html, */*")
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)

	buf, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, `<html><body><div>header</div>
<h1>main</h1>
<div>abc.com/index</div>
<div>footer</div>
</body></html>`, string(buf))

	req, err = http.NewRequest("GET", host+"/admin/index", nil)
	req.Header.Set("Accept", "text/html, */*")
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)

	buf, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, `<html><body><div>header</div>
<h1>admin</h1>
<div>abc.com/admin/index</div>
<div>footer</div>
</body></html>`, string(buf))

}

func TestMixedViewers(t *testing.T) {
	fsys := fstest.MapFS{
		"views/user.html":  {Data: []byte(`user`)},
		"pages/index.html": {Data: []byte(`index`)},
		"pages/list.html":  {Data: []byte(`list`)},
	}

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	app := New(WithMux(mux), WithFsys(fsys))

	app.Get("/", func(c *Context) error {
		if c.Request().URL.Path == "/" {
			return c.View(nil, "index")
		}

		c.WriteStatus(http.StatusNotFound)
		return ErrCancelled
	})

	app.Get("/list", func(c *Context) error {
		return c.View(map[string]any{
			"name": "list",
			"num":  2,
		})
	})

	app.Start()
	defer app.Close()

	req, err := http.NewRequest("GET", srv.URL+"/", nil)
	req.Header.Set("Accept", "text/html, */*")
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)

	buf, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, fsys["pages/index.html"].Data, buf)

	req, err = http.NewRequest("GET", srv.URL+"/index", nil)
	req.Header.Set("Accept", "text/html, */*")
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)

	buf, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, fsys["pages/index.html"].Data, buf)

	req, err = http.NewRequest("GET", srv.URL+"/404", nil)
	req.Header.Set("Accept", "text/html, */*")
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()

	req, err = http.NewRequest("GET", srv.URL+"/list", nil)
	req.Header.Set("Accept", "text/html, */*")
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)

	buf, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, fsys["pages/list.html"].Data, buf)

	req, err = http.NewRequest("GET", srv.URL+"/list", nil)
	req.Header.Set("Accept", "application/json")
	require.NoError(t, err)
	resp, err = client.Do(req)
	require.NoError(t, err)

	buf, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	data := &struct {
		Name string `json:"name"`
		Num  int    `json:"num"`
	}{}

	err = json.Unmarshal(buf, data)

	require.NoError(t, err)

	require.Equal(t, data.Name, "list")
	require.Equal(t, data.Num, 2)

}

func TestApp(t *testing.T) {
	app := New(WithMux(http.NewServeMux()),
		WithFsys(os.DirFS(".")))

	app.Get("/hello", func(c *Context) error {
		//c.View(map[string]string{"name": "World"})

		return nil
	})

	admin := app.Group("/admin")

	admin.Use(func(next HandleFunc) HandleFunc {
		return func(c *Context) error {
			if c.routing.Options.String(NavigationAccess) != "admin:*" {
				c.WriteStatus(http.StatusForbidden)
				return ErrCancelled
			}

			return next(c)
		}

	})

	admin.Get("/", func(c *Context) error {
		return c.View(nil)

	}, WithNavigation("admin", "fa fa-home", "admin:*"))

	admin.Post("/form", func(c *Context) error {
		data, err := BindJSON[TestData](c.Request())

		if err != nil {
			c.WriteStatus(http.StatusBadRequest)
			return ErrCancelled
		}

		if !data.Validate(c.AcceptLanguage()...) {
			c.WriteStatus(http.StatusBadRequest)
			return c.View(data)
		}

		return c.View(data)
	})

	admin.Get("/search", func(c *Context) error {
		data, err := BindQuery[TestData](c.Request())

		if err != nil {
			c.WriteStatus(http.StatusBadRequest)
			return ErrCancelled
		}

		if !data.Validate(c.AcceptLanguage()...) {
			c.WriteStatus(http.StatusBadRequest)
			return c.View(data)
		}

		return c.View(data)
	})

	admin.Post("/form", func(c *Context) error {
		data, err := BindForm[TestData](c.Request())

		if err != nil {
			c.WriteStatus(http.StatusBadRequest)
			return ErrCancelled
		}

		if !data.Validate(c.AcceptLanguage()...) {
			c.WriteStatus(http.StatusBadRequest)
			return c.View(data)
		}

		return c.View(data)
	})
}

type TestData struct {
}
