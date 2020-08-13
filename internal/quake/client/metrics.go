package client

import (
	"context"
	"log"
	"time"

	quakenet "github.com/criticalstack/quake-kube/internal/quake/net"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	activeConns = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "quake_active_connections",
		Help: "The current number of active websocket proxy connections",
	})
	scores = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "quake_player_scores",
		Help: "Current scores by player, by map",
	}, []string{"player", "map"})
	pings = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "quake_player_pings",
		Help: "Current ping by player",
	}, []string{"player"})
)

func (s *Server) pollStatusMetrics(ctx context.Context, interval time.Duration) {
	tick := time.NewTicker(interval)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			status, err := quakenet.GetStatus(s.ServerAddr)
			if err != nil {
				log.Printf("metrics: get status failed %v", err)
				continue
			}
			activeConns.Set(float64(len(status.Players)))
			for _, p := range status.Players {
				scores.WithLabelValues(p.Name, status.Configuration["mapname"]).Set(float64(p.Score))
				pings.WithLabelValues(p.Name).Set(float64(p.Ping))
			}
		case <-ctx.Done():
			return
		}
	}
}
