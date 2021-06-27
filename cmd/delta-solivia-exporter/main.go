package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/go-kit/kit/log/level"
	"github.com/kyleplant/delta-solivia-exporter/pkg/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/version"
)

var (
	metricsAddr = flag.String("metrics.addr", ":9134", "host:port for delta solivia exporter")
	metricsPath = flag.String("metrics.path", "/metrics", "URL path for surfacing collected metrics")
)

func init() {
	prometheus.MustRegister(version.NewCollector("delta_solivia_exporter"))
}

func main() {
	var (
		opts = exporter.SerialOpts{}
	)

	promlogConfig := &promlog.Config{}
	logger := promlog.New(promlogConfig)

	// table := crc16.MakeTable(crc16.CRC16_MODBUS)

	// fmt.Printf("%v\n", table)

	// crc := crc16.Checksum([]byte("\x13\x03"), table)
	// fmt.Printf("CRC-16 MAXIM: %X\n", crc)

	// // using the standard library hash.Hash interface
	// h := crc16.New(table)
	// h.Write([]byte("\x13\x03"))
	// fmt.Printf("CRC-16 MAXIM: %X\n", h.Sum16())

	level.Info(logger).Log("msg", "Starting delta_solivia_exporter", "version", version.Info())
	level.Info(logger).Log("build_context", version.BuildContext())

	exporter, err := exporter.New(opts, logger)
	if err != nil {
		level.Error(logger).Log("msg", "Error creating the exporter", "err", err)
		os.Exit(1)
	}

	prometheus.MustRegister(exporter)

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
