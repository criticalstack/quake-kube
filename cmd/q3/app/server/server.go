package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

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
	Maps          string
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
			if opts.ServerAddr == "" {
				hostIPv4, err := netutil.DetectHostIPv4()
				if err != nil {
					return err
				}
				opts.ServerAddr = fmt.Sprintf("%s:27960", hostIPv4)
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
			if err := content.CopyAssets(csurl, opts.AssetsDir); err != nil {
				return err
			}
			if err := writeDefaultServerConfig(filepath.Join(opts.AssetsDir, "baseq3/server.cfg")); err != nil {
				return err
			}
			if opts.ConfigFile != "" {
				data, err := ioutil.ReadFile(opts.ConfigFile)
				if err != nil {
					return err
				}
				if err := ioutil.WriteFile(filepath.Join(opts.AssetsDir, "baseq3/server.cfg"), data, 0644); err != nil {
					return err
				}
			}

			if err := writeDefaultMapConfig(filepath.Join(opts.AssetsDir, "baseq3/maps.cfg")); err != nil {
				return err
			}
			if opts.Maps != "" {
				data, err := ioutil.ReadFile(opts.Maps)
				if err != nil {
					return err
				}
				var maps quakeserver.Maps
				if err := yaml.Unmarshal(data, &maps); err != nil {
					return err
				}
				data, err = maps.Marshal()
				if err != nil {
					return err
				}
				if err := ioutil.WriteFile(filepath.Join(opts.AssetsDir, "baseq3/maps.cfg"), data, 0644); err != nil {
					return err
				}
			}

			go func() {
				if err := quakeserver.Start(ctx, opts.AssetsDir); err != nil {
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
	cmd.Flags().StringVar(&opts.ServerAddr, "server-addr", "", "dedicated server <host>:<port>")
	cmd.Flags().StringVar(&opts.Maps, "maps", "", "map rotation")
	return cmd
}

func writeDefaultMapConfig(path string) error {
	maps := quakeserver.Maps{
		{Name: "q3dm7", Type: quakeserver.FreeForAll},
		{Name: "q3dm17", Type: quakeserver.FreeForAll},
	}
	data, err := maps.Marshal()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}

func writeDefaultServerConfig(path string) error {
	data, err := quakeserver.Default().Marshal()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}
