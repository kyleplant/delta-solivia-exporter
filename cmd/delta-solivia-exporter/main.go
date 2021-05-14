package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	//"github.com/kyleplant/delta-solivia-exporter/pkg/exporter"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
)

var (
	metricsAddr = flag.String("metrics.addr", ":9134", "host:port for delta solivia exporter")
	metricsPath = flag.String("metrics.path", "/metrics", "URL path for surfacing collected metrics")
)

func main() {

	promlogConfig := &promlog.Config{}
	logger := promlog.New(promlogConfig)
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Delta Solivia Exporter</title></head>
             <body>
             <h1>Delta Solivia Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})

	http.HandleFunc("/-/healthy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	http.HandleFunc("/-/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	level.Info(logger).Log("msg", "Listening on address", "address", *metricsAddr)
	if err := http.ListenAndServe(*metricsAddr, nil); err != nil {
		level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
