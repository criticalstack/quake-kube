package proxy

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	quakeclient "github.com/criticalstack/quake-kube/internal/quake/client"
	netutil "github.com/criticalstack/quake-kube/internal/util/net"
)

var opts struct {
	ClientAddr    string
	ServerAddr    string
	ContentServer string
}

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "proxy",
		Short:        "q3 websocket/udp proxy",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.ClientAddr == "" {
				hostIPv4, err := netutil.DetectHostIPv4()
				if err != nil {
					return err
				}
				opts.ClientAddr = fmt.Sprintf("%s:8080", hostIPv4)
			}
			p, err := quakeclient.NewProxy(opts.ServerAddr)
			if err != nil {
				return err
			}
			s := http.Server{
				Addr:    opts.ClientAddr,
				Handler: p,
			}
			return s.ListenAndServe()
		},
	}
	cmd.Flags().StringVarP(&opts.ClientAddr, "client-addr", "c", "", "client address <host>:<port>")
	cmd.Flags().StringVarP(&opts.ServerAddr, "server-addr", "s", "", "dedicated server <host>:<port>")
	return cmd
}
