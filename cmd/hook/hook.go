package hook

import "github.com/spf13/cobra"

// HookCmd represents the hook command
var HookCmd = &cobra.Command{Use: "hook"}

func init() {
	HookCmd.AddCommand(postReceiveCmd)
}
