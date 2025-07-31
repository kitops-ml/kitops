package completions

import (
	"strings"

	"github.com/spf13/cobra"
)

// WithStaticArgCompletions adds static argument completions to a command
func WithStaticArgCompletions(cmd *cobra.Command, completions []string, maxCompletionsCount int) {
	cmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > (maxCompletionsCount - 1) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var suggestions []string
		for _, c := range completions {
			if strings.HasPrefix(c, toComplete) {
				suggestions = append(suggestions, c)
			}
		}
		return suggestions, cobra.ShellCompDirectiveNoFileComp
	}
}
