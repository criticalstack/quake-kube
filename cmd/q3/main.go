package main

import (
	"log"

	"github.com/spf13/cobra"

	q3cmd "github.com/criticalstack/quake-kube/cmd/q3/app/cmd"
	q3content "github.com/criticalstack/quake-kube/cmd/q3/app/content"
	q3proxy "github.com/criticalstack/quake-kube/cmd/q3/app/proxy"
	q3server "github.com/criticalstack/quake-kube/cmd/q3/app/server"
)

var global struct {
	Verbosity int
}

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "q3",
		Short: "",
	}
	cmd.AddCommand(
		q3cmd.NewCommand(),
		q3content.NewCommand(),
		q3proxy.NewCommand(),
		q3server.NewCommand(),
	)

	cmd.PersistentFlags().CountVarP(&global.Verbosity, "verbose", "v", "log output verbosity")
	return cmd
}

func main() {
	if err := NewRootCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}
