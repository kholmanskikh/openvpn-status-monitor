package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/kholmanskikh/openvpn-status-monitor/internal/monitor"
	"github.com/kholmanskikh/openvpn-status-monitor/internal/webapp"
)

func main() {
	statusFilePath := flag.String("status-file", "", "path to the OpenVPN status file")
	interval := flag.Duration("interval", 5*time.Second, "read interval")
	listenAddr := flag.String("listen", ":8080",
		"listen address as in go http.ListenAndServe")
	flag.Parse()

	if *statusFilePath == "" {
		log.Fatal("No OpenVPN status file path specified")
	}

	app := webapp.NewApp(*listenAddr)
	go monitor.Monitor(*statusFilePath, *interval, app.UpdateChannel())

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}
