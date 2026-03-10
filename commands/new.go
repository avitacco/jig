package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/avitacco/jig/internal/scaffold"
	"github.com/spf13/cobra"
)

func (a *App) newCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new",
		Short: "Create new things",
	}
	cmd.AddCommand(a.newModuleCmd())
	return cmd
}

func (a *App) newModuleCmd() *cobra.Command {
	newModuleCmd := &cobra.Command{
		Use:   "module <name>",
		Short: "Create a new Puppet module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			forgeUser, _ := cmd.Flags().GetString("forge-user")
			license, _ := cmd.Flags().GetString("license")
			summary, _ := cmd.Flags().GetString("summary")
			source, _ := cmd.Flags().GetString("source")
			author, _ := cmd.Flags().GetString("author")
			force, _ := cmd.Flags().GetBool("force")

			if forgeUser == "" {
				forgeUser = a.Config.ForgeUsername
			}

			if author == "" {
				author = a.Config.Author
			}

			if forgeUser == "" || author == "" {
				currentUser, err := user.Current()
				if err != nil {
					return err
				}
				if forgeUser == "" {
					forgeUser = currentUser.Username
				}
				if author == "" {
					author = currentUser.Name
				}
			}

			opts := scaffold.Options{
				ForgeUser: forgeUser,
				Name:      args[0],
				License:   license,
				Summary:   summary,
				Source:    source,
				Author:    author,
				Force:     force,
			}

			err := runModuleInterview(&opts)
			if err != nil {
				return err
			}

			return scaffold.NewModule(opts)
		},
	}

	newModuleCmd.Flags().StringP("forge-user", "u", "", "Forge username")
	newModuleCmd.Flags().StringP("author", "a", "", "Author name")
	newModuleCmd.Flags().StringP("license", "l", "Apache-2.0", "License type")
	newModuleCmd.Flags().StringP("summary", "s", "", "Summary of the module")
	newModuleCmd.Flags().StringP("source", "S", "", "Source URL for the module")
	newModuleCmd.Flags().BoolP("force", "f", false, "Force creation of the module even if it already exists. Note: a backup of the existing directory will be created.")

	return newModuleCmd
}

func runModuleInterview(opts *scaffold.Options) error {
	opts.ForgeUser, _ = prompt("Forge username", opts.ForgeUser)
	opts.Author, _ = prompt("Author name", opts.Author)
	opts.License, _ = prompt("License type", opts.License)
	opts.Summary, _ = prompt("Summary of the module", opts.Summary)
	opts.Source, _ = prompt("Source URL for the module", opts.Source)
	return nil
}

func prompt(question string, defaultVal string) (string, error) {
	fmt.Printf("%s [%s]: ", question, defaultVal)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		return defaultVal, nil
	}
	return input, nil
}
