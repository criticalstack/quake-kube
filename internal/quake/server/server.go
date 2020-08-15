package server

import (
	"context"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"sigs.k8s.io/yaml"

	"github.com/criticalstack/quake-kube/internal/util/exec"
)

type Server struct {
	Dir           string
	WatchInterval time.Duration
	ConfigFile    string
	Addr          string
}

func (s *Server) Start(ctx context.Context) error {
	if s.Addr == "" {
		s.Addr = "0.0.0.0:27960"
	}
	host, port, err := net.SplitHostPort(s.Addr)
	if err != nil {
		return err
	}
	args := []string{
		"+set", "dedicated", "1",
		"+set", "net_ip", host,
		"+set", "net_port", port,
		"+exec", "server.cfg",
	}
	cmd := exec.CommandContext(ctx, "ioq3ded", args...)
	cmd.Dir = s.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if s.ConfigFile == "" {
		cfg := Default()
		data, err := cfg.Marshal()
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(filepath.Join(s.Dir, "baseq3/server.cfg"), data, 0644); err != nil {
			return err
		}
		if err := cmd.Start(); err != nil {
			return err
		}
		return cmd.Wait()
	}

	if err := s.reload(); err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		if err := cmd.Wait(); err != nil {
			log.Println(err)
		}
	}()

	ch, err := s.watch(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ch:
			if err := s.reload(); err != nil {
				return err
			}
			if err := cmd.Restart(ctx); err != nil {
				return err
			}
			go func() {
				if err := cmd.Wait(); err != nil {
					log.Println(err)
				}
			}()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *Server) reload() error {
	data, err := ioutil.ReadFile(s.ConfigFile)
	if err != nil {
		return err
	}
	var cfg *Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}
	data, err = cfg.Marshal()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(s.Dir, "baseq3/server.cfg"), data, 0644)
}

func (s *Server) watch(ctx context.Context) (<-chan struct{}, error) {
	if s.WatchInterval == 0 {
		s.WatchInterval = 15 * time.Second
	}
	cur, err := os.Stat(s.ConfigFile)
	if err != nil {
		return nil, err
	}

	ch := make(chan struct{})

	go func() {
		ticker := time.NewTicker(s.WatchInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if fi, err := os.Stat(s.ConfigFile); err == nil {
					if fi.ModTime().After(cur.ModTime()) {
						ch <- struct{}{}
					}
					cur = fi
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}
