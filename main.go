package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/NexClipper/ds-switch/pkg/config"
	"github.com/NexClipper/ds-switch/pkg/ds"
	"github.com/NexClipper/ds-switch/pkg/monitor"
)

func main() {
	configPath := flag.String("config", "./conf/ds.yml", "path to ds-switch's config file")
	flag.Parse()

	cfg, err := config.New(*configPath)
	if err != nil {
		panic(err)
	}

	log.Println(cfg)

	datasource := ds.New(cfg)
	if datasource == nil {
		return
	}

	// test
	// if err := datasource.SetDefaultDatasource("Prometheus"); err != nil {
	// 	panic(err)
	// }

	p := monitor.New(fmt.Sprintf("%s%s", cfg.Prometheus.End_Point, cfg.Prometheus.Monitor_API.Method),
		cfg.DS_Switch.Monitor_Interval,
		cfg.DS_Switch.Evaluate_Interval,
		datasource)
	p.Run()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	select {
	case <-quit:
		os.Exit(1)
	}
}
