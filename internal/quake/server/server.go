package server

import (
	"context"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"sigs.k8s.io/yaml"

	quakenet "github.com/criticalstack/quake-kube/internal/quake/net"
	"github.com/criticalstack/quake-kube/internal/util/exec"
)

var (
	actrvePlayers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "quake_active_players",
		Help: "The current number of active players",
	})

	scores = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "quake_player_scores",
		Help: "Current scores by player, by map",
	}, []string{"player", "map"})

	pings = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "quake_player_pings",
		Help: "Current ping by player",
	}, []string{"player"})

	configReloads = promauto.NewCounter(prometheus.CounterOpts{
		Name: "quake_config_reloads",
		Help: "Config file reload count",
	})
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
		"+set", "com_homepath", s.Dir,
		"+set", "com_basegame", "baseq3",
		"+set", "com_gamename", "Quake3Arena",
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

	go func() {
		addr := s.Addr
		if net.ParseIP(host).IsUnspecified() {
			addr = net.JoinHostPort("127.0.0.1", port)
		}
		tick := time.NewTicker(5 * time.Second)
		defer tick.Stop()

		for {
			select {
			case <-tick.C:
				status, err := quakenet.GetStatus(addr)
				if err != nil {
					log.Printf("metrics: get status failed %v", err)
					continue
				}
				actrvePlayers.Set(float64(len(status.Players)))
				for _, p := range status.Players {
					if mapname, ok := status.Configuration["mapname"]; ok {
						scores.WithLabelValues(p.Name, mapname).Set(float64(p.Score))
					}
					pings.WithLabelValues(p.Name).Set(float64(p.Ping))
				}
			case <-ctx.Done():
				return
			}
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
			configReloads.Inc()
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
	cfg := Default()
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
