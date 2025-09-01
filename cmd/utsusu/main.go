package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/zebdo/utsusu/internal/chans"
	"github.com/zebdo/utsusu/internal/core"
	"github.com/zebdo/utsusu/internal/server"
	"github.com/zebdo/utsusu/internal/storage"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Port                 string `yaml:"port"`
	SQLitePath           string `yaml:"sqlite_path"`
	AdminToken           string `yaml:"admin_token"`
	UserAgent            string `yaml:"user_agent"`
	DefaultWatchInterval string `yaml:"default_watch_interval"`
}

func main() {
	cfgPath := flag.String("config", "config.yaml", "config file")
	flag.Parse()

	cfg := &Config{Port: "7200", DefaultWatchInterval: "30s"}
	if b, err := ioutil.ReadFile(*cfgPath); err == nil {
		_ = yaml.Unmarshal(b, cfg)
	}

	if env := os.Getenv("GOCHAN_PORT"); env != "" {
		cfg.Port = env
	}
	if env := os.Getenv("GOCHAN_SQLITE_PATH"); env != "" {
		cfg.SQLitePath = env
	}
	if env := os.Getenv("GOCHAN_ADMIN_TOKEN"); env != "" {
		cfg.AdminToken = env
	}
	if env := os.Getenv("GOCHAN_USER_AGENT"); env != "" {
		cfg.UserAgent = env
	}

	var store storage.Storage
	if cfg.SQLitePath != "" {
		st, err := storage.NewSQLite(cfg.SQLitePath)
		if err != nil {
			log.Fatalf("sqlite init: %v", err)
		}
		store = st
		log.Printf("using sqlite at %s", cfg.SQLitePath)
	} else {
		store = storage.NewMemory()
		log.Printf("using in-memory store")
	}

	sources := map[string]chans.ChanSource{
		"demo":  chans.NewDemoSource(),
		"4chan": chans.NewFourChan(chans.FourChanConfig{UserAgent: cfg.UserAgent}),
	}

	arch := core.NewArchiver(store, sources)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go arch.Run(ctx)

	srv := server.New(store, sources, arch, cfg.AdminToken)
	addr := ":" + cfg.Port
	log.Printf("starting server on %s", addr)
	if err := srv.Run(addr); err != nil {
		log.Fatal(err)
	}
}