package server

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	quakeclient "github.com/criticalstack/quake-kube/internal/quake/client"
	"github.com/criticalstack/quake-kube/internal/quake/content"
	quakeserver "github.com/criticalstack/quake-kube/internal/quake/server"
	netutil "github.com/criticalstack/quake-kube/internal/util/net"
	httputil "github.com/criticalstack/quake-kube/internal/util/net/http"
	"github.com/criticalstack/quake-kube/public"
)

var opts struct {
	ClientAddr    string
	ServerAddr    string
	ContentServer string
	AcceptEula    bool
	AssetsDir     string
	ConfigFile    string
	WatchInterval time.Duration
}

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "server",
		Short:        "q3 server",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.ClientAddr == "" {
				hostIPv4, err := netutil.DetectHostIPv4()
				if err != nil {
					return err
				}
				opts.ClientAddr = fmt.Sprintf("%s:8080", hostIPv4)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			csurl, err := url.Parse(opts.ContentServer)
			if err != nil {
				return err
			}
			if !opts.AcceptEula {
				fmt.Println(quakeserver.Q3DemoEULA)
				return errors.New("You must agree to the EULA to continue")
			}
			if err := httputil.GetUntil(opts.ContentServer, ctx.Done()); err != nil {
				return err
			}

			// TODO(chrism): only download what is in map config
			if err := content.CopyAssets(csurl, opts.AssetsDir); err != nil {
				return err
			}

			go func() {
				s := quakeserver.Server{
					Dir:           opts.AssetsDir,
					WatchInterval: opts.WatchInterval,
					ConfigFile:    opts.ConfigFile,
					Addr:          opts.ServerAddr,
				}
				if err := s.Start(ctx); err != nil {
					panic(err)
				}
			}()

			e, err := quakeclient.NewRouter(&quakeclient.Config{
				ContentServerURL: opts.ContentServer,
				ServerAddr:       opts.ServerAddr,
				Files:            public.Files,
			})
			if err != nil {
				return err
			}
			s := &quakeclient.Server{
				Addr:       opts.ClientAddr,
				Handler:    e,
				ServerAddr: opts.ServerAddr,
			}
			fmt.Printf("Starting server %s\n", opts.ClientAddr)
			return s.ListenAndServe()
		},
	}
	cmd.Flags().StringVarP(&opts.ConfigFile, "config", "c", "", "server configuration file")
	cmd.Flags().StringVar(&opts.ContentServer, "content-server", "http://content.quakejs.com", "content server url")
	cmd.Flags().BoolVar(&opts.AcceptEula, "agree-eula", false, "agree to the Quake 3 demo EULA")
	cmd.Flags().StringVar(&opts.AssetsDir, "assets-dir", "assets", "location for game files")
	cmd.Flags().StringVar(&opts.ClientAddr, "client-addr", "", "client address <host>:<port>")
	cmd.Flags().StringVar(&opts.ServerAddr, "server-addr", "0.0.0.0:27960", "dedicated server <host>:<port>")
	cmd.Flags().DurationVar(&opts.WatchInterval, "watch-interval", 15*time.Second, "dedicated server <host>:<port>")
	return cmd
}
