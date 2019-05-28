package app

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMuxRedirect(t *testing.T) {
	mux := http.NewServeMux()

	muxRedirect(mux, "/old/path", "/new/path")
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
}

func TestMuxRedirectPrefix(t *testing.T) {
	mux := http.NewServeMux()

	muxRedirectPrefix(mux, "/old/", "/new/")
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
}

func TestMuxRewrite(t *testing.T) {
	t.Run("simple rewrite", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/new/path", func(w http.ResponseWriter, req *http.Request) {
			io.WriteString(w, req.URL.String())
		})
		muxRewrite(mux, "/old/path", "/new/path")

		srv := httptest.NewServer(mux)
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
		muxRewrite(mux, "/old/path", "/new/path?a=b")

		srv := httptest.NewServer(mux)
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
}

func TestMuxRewritePrefix(t *testing.T) {
	t.Run("simple prefix", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/new/path", func(w http.ResponseWriter, req *http.Request) {
			io.WriteString(w, req.URL.String())
		})
		muxRewritePrefix(mux, "/old/", "/new/")

		srv := httptest.NewServer(mux)
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
		muxRewritePrefix(mux, "/old/", "/new/?c=d")

		srv := httptest.NewServer(mux)
		defer srv.Close()

		req, err := http.NewRequest("GET", srv.URL+"/old/path?a=b", nil)
		assert.Nil(t, err)

		resp, err := http.DefaultTransport.RoundTrip(req)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Status Code")
		data, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err)

		assert.Equal(t, "/new/path?a=b&c=d", string(data))
	})
}
