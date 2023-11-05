package template

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type CmdTemplate struct {
	Title    string
	Commands []*cobra.Command
}

type cmdGroup []CmdTemplate

func CreatCmdGroup(cmdTemplates ...CmdTemplate) []CmdTemplate {
	cmdGroups := []CmdTemplate{}
	cmdGroups = append(cmdGroups, cmdTemplates...)
	return cmdGroups
}

func HelpFunc(cmd *cobra.Command, cmdGrp cmdGroup) {
	out := cmd.OutOrStdout()
	indentFirstLint := false
	fmt.Fprintf(out, "%s\n\n", cmd.Long)
	fmt.Fprintf(out, "Usage:\n  %s\n\n", cmd.UseLine())

	for _, group := range cmdGrp {
		fmt.Fprintln(out, group.Title)
		for idx, c := range group.Commands {
			if c.Runnable() {
				if strings.TrimSpace(c.Use) == "" {
					continue
				}

				if idx == 0 && !indentFirstLint {
					fmt.Fprintf(out, "  %s\t\t%s\n", c.Name(), c.Short)
					indentFirstLint = true
					continue
				}
				fmt.Fprintf(out, "  %s\t%s\n", c.Name(), c.Short)
			}
		}
		fmt.Fprintln(out)
	}

	// Print the rest of the commands that are not included in the groups
	fmt.Fprintln(out, "Other Commands:")
	for _, c := range cmd.Commands() {
		if !c.Runnable() || isInGroup(c, cmdGrp) {
			continue
		}
		fmt.Fprintf(out, "  %s\t%s\n", c.Name(), c.Short)
	}
	fmt.Fprintln(out)

	fmt.Fprintf(out, "Use \"%s [command] --help\" for more information about a command.\n", cmd.CommandPath())
}

// Helper function to determine if the command is in any group
func isInGroup(c *cobra.Command, groups cmdGroup) bool {
	for _, group := range groups {
		for _, gc := range group.Commands {
			if gc == c {
				return true
			}
		}
	}
	return false
}
