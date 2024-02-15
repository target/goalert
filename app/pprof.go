package app

import (
	"net"
	"net/http"
	"net/http/pprof"

	"github.com/spf13/viper"
)

func initPprofServer() error {
	addr := viper.GetString("listen-pprof")
	if addr == "" {
		return nil
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()

	// Register pprof handlers (matches init() of net/http/pprof package)
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	srv := http.Server{
		Handler: mux,
	}
	go func() { _ = srv.Serve(l) }()
	return nil
}
