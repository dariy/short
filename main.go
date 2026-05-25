package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Links map[string]string `yaml:"links"`
}

type Server struct {
	mu       sync.RWMutex
	links    map[string]string
	cfgPath  string
}

func (s *Server) loadConfig() error {
	data, err := os.ReadFile(s.cfgPath)
	if err != nil {
		return err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}
	s.mu.Lock()
	s.links = cfg.Links
	s.mu.Unlock()
	log.Printf("loaded %d links from %s", len(cfg.Links), s.cfgPath)
	return nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hash := strings.TrimPrefix(r.URL.Path, "/")
	if hash == "" {
		http.NotFound(w, r)
		return
	}

	s.mu.RLock()
	target, ok := s.links[hash]
	s.mu.RUnlock()

	if !ok {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, target, http.StatusMovedPermanently)
}

func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "/config/links.yml"
	}
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	srv := &Server{cfgPath: cfgPath}
	if err := srv.loadConfig(); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Reload config on SIGHUP
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGHUP)
		for range ch {
			if err := srv.loadConfig(); err != nil {
				log.Printf("reload failed: %v", err)
			}
		}
	}()

	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
