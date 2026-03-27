package web

import (
	"github.com/spf13/cobra"

	webconsole "github.com/raynaythegreat/ai-business-hq/web/backend"
	"github.com/raynaythegreat/ai-business-hq/web/backend/utils"
)

func NewWebCommand() *cobra.Command {
	var opts webconsole.Options

	cmd := &cobra.Command{
		Use:   "web [config.json]",
		Short: "Start the web console",
		Long:  "Start the OctAi web console at http://localhost:18800",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.ConfigPath = args[0]
			} else {
				opts.ConfigPath = utils.GetDefaultConfigPath()
			}
			opts.ExplicitPort = cmd.Flags().Changed("port")
			opts.ExplicitPublic = cmd.Flags().Changed("public")
			return webconsole.Run(opts)
		},
	}

	cmd.Flags().StringVar(&opts.Port, "port", "18800", "Port to listen on")
	cmd.Flags().BoolVar(&opts.Public, "public", false, "Listen on all interfaces (0.0.0.0)")
	cmd.Flags().BoolVar(&opts.NoBrowser, "no-browser", false, "Do not auto-open browser on startup")
	cmd.Flags().StringVar(&opts.Lang, "lang", "", "Language: en (English) or zh (Chinese)")
	cmd.Flags().BoolVar(&opts.Console, "console", false, "Console mode, no system tray GUI")

	return cmd
}
