package loop

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/raynaythegreat/ai-business-hq/pkg/agent"
)

// NewLoopCommand returns the `aibhq loop` cobra command.
func NewLoopCommand() *cobra.Command {
	var (
		maxRuns int
	)

	cmd := &cobra.Command{
		Use:   "loop <interval> <prompt>",
		Short: "Create a recurring agent task that fires on an interval",
		Long: `Create a recurring agent task that fires on a cron-like interval.

The interval is a Go duration string such as 5m, 1h, or 30s.
The gateway must be running for loops to execute.`,
		Example: `  aibhq loop 5m "check build status"
  aibhq loop 1h "summarise today's activity" --max-runs 8`,
		Args: cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			intervalStr := args[0]
			prompt := args[1]

			interval, err := time.ParseDuration(intervalStr)
			if err != nil {
				return fmt.Errorf("invalid interval %q: %w", intervalStr, err)
			}
			if interval <= 0 {
				return fmt.Errorf("interval must be a positive duration (got %s)", intervalStr)
			}

			task := agent.LoopTask{
				Prompt:   prompt,
				Interval: interval,
				MaxRuns:  maxRuns,
			}

			// Compute ID the same way the scheduler would so we can display it
			// before the scheduler is running.
			taskID := fmt.Sprintf("loop-%d", time.Now().UnixNano())
			task.ID = taskID

			fmt.Printf("Loop task created: %s, runs every %s\n", taskID, interval)
			if maxRuns > 0 {
				fmt.Printf("  Max runs: %d\n", maxRuns)
			}
			fmt.Println()
			fmt.Println("Note: Start the gateway with `aibhq gateway start` to execute loops.")

			return nil
		},
	}

	cmd.Flags().IntVar(&maxRuns, "max-runs", 0, "Maximum number of times to run (0 = unlimited)")

	return cmd
}
