//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package command

import (
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

const (
	requestTimeout    = time.Second * 10
	outputFormatHuman = "human"
	outputFormatJSON  = "json"
	outputFormatYAML  = "yaml"
	usageOutputFormat = `Output format.
Choices are 'json','yaml', and 'human'.`
)

var (
	version       string
	buildHash     string
	buildDate     string
	debug         bool
	outputFormat  string
	outputFormats = []string{
		outputFormatHuman,
		outputFormatJSON,
		outputFormatYAML,
	}
)

var (
	twColorRedNormal    = tablewriter.Colors{tablewriter.Normal, tablewriter.FgHiRedColor}
	twColorYellowNormal = tablewriter.Colors{tablewriter.Normal, tablewriter.FgHiYellowColor}
	twColorGreenNormal  = tablewriter.Colors{tablewriter.Normal, tablewriter.FgHiGreenColor}
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kwiz",
	Short: "kwiz - NUMA-aware Kubernetes capacity magician.",
	Args:  validateRootCommand,
	Long: `
 _              _     
| |            (_)    
| | ____      ___ ____
| |/ /\ \ /\ / / |_  /
|   <  \ V  V /| |/ / 
|_|\_\  \_/\_/ |_/___|

The NUMA-aware Kubernetes capacity magician.

https://github.com/jaypipes/kwiz
`,
	RunE: showNodeResourceSummary,
}

// Execute adds all child commands to the root command and sets flags
// appropriately. This is called by main.main(). It only needs to happen once
// to the rootCmd.
func Execute(v string, bh string, bd string) {
	version = v
	buildHash = bh
	buildDate = bd

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func haveValidOutputFormat() bool {
	for _, choice := range outputFormats {
		if choice == outputFormat {
			return true
		}
	}
	return false
}

// validateRootCommand ensures any CLI options or arguments are valid,
// returning an error if not
func validateRootCommand(rootCmd *cobra.Command, args []string) error {
	if !haveValidOutputFormat() {
		return fmt.Errorf("invalid output format %q", outputFormat)
	}
	return nil
}

func init() {
	rootCmd.PersistentFlags().BoolVar(
		&debug, "debug", false, "Enable or disable debug mode",
	)
	rootCmd.PersistentFlags().StringVarP(
		&outputFormat,
		"format", "f",
		outputFormatHuman,
		usageOutputFormat,
	)
}
