package app

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPRedirect(t *testing.T) {

	t.Run("no prefix", func(t *testing.T) {
		mux := httpRedirect("", "/old/path", "/new/path")(http.NewServeMux())
		srv := httptest.NewServer(mux)
		defer srv.Close()

		req, err := http.NewRequest("GET", srv.URL+"/old/path", nil)
		assert.Nil(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode, "Status Code")
		loc, err := resp.Location()
		assert.Nil(t, err)

		assert.Equal(t, srv.URL+"/new/path", loc.String(), "redirect URL")
	})

	t.Run("with prefix", func(t *testing.T) {
		mux := httpRedirect("/foobar", "/old/path", "/new/path")(http.NewServeMux())
		srv := httptest.NewServer(mux)
		defer srv.Close()

		req, err := http.NewRequest("GET", srv.URL+"/old/path", nil)
		assert.Nil(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode, "Status Code")
		loc, err := resp.Location()
		assert.Nil(t, err)

		assert.Equal(t, srv.URL+"/foobar/new/path", loc.String(), "redirect URL")
	})
}

func TestMuxRewrite(t *testing.T) {
	t.Run("simple rewrite", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/new/path", func(w http.ResponseWriter, req *http.Request) {
			io.WriteString(w, req.URL.String())
		})
		h := httpRewrite("", "/old/path", "/new/path")(mux)

		srv := httptest.NewServer(h)
		defer srv.Close()

		req, err := http.NewRequest("GET", srv.URL+"/old/path", nil)
		assert.Nil(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Status Code")
		data, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err)

		assert.Equal(t, "/new/path", string(data))
	})
	t.Run("query params", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/new/path", func(w http.ResponseWriter, req *http.Request) {
			io.WriteString(w, req.URL.String())
		})
		h := httpRewrite("", "/old/path", "/new/path?a=b")(mux)

		srv := httptest.NewServer(h)
		defer srv.Close()

		req, err := http.NewRequest("GET", srv.URL+"/old/path?c=d", nil)
		assert.Nil(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Status Code")
		data, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err)

		assert.Equal(t, "/new/path?a=b&c=d", string(data))
	})
	t.Run("simple rewrite (prefix)", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/foobar/new/path", func(w http.ResponseWriter, req *http.Request) {
			io.WriteString(w, req.URL.String())
		})
		h := httpRewrite("/foobar", "/old/path", "/new/path")(mux)

		srv := httptest.NewServer(h)
		defer srv.Close()

		req, err := http.NewRequest("GET", srv.URL+"/old/path", nil)
		assert.Nil(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Status Code")
		data, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err)

		assert.Equal(t, "/foobar/new/path", string(data))
	})
	t.Run("simple rewrite (prefix+route)", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/foobar/new/path", func(w http.ResponseWriter, req *http.Request) {
			io.WriteString(w, req.URL.String())
		})
		h := httpRewrite("/foobar", "/old/", "/new/")(mux)

		srv := httptest.NewServer(h)
		defer srv.Close()

		req, err := http.NewRequest("GET", srv.URL+"/old/path", nil)
		assert.Nil(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Status Code")
		data, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err)

		assert.Equal(t, "/foobar/new/path", string(data))
	})

}
