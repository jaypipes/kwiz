//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package command

import (
	"context"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	kwcontext "github.com/jaypipes/kwiz/pkg/context"
	"github.com/jaypipes/kwiz/pkg/kube"
	kconnect "github.com/jaypipes/kwiz/pkg/kube/connect"
	knode "github.com/jaypipes/kwiz/pkg/kube/node"
	"github.com/jaypipes/kwiz/pkg/unit"
)

var nodeGetOpts = knode.NodeGetOptions{}

// nodeCmd represents the node command
var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Show node resource summary",
	RunE:  showNodeResourceSummary,
}

func init() {
	cmdutil.AddLabelSelectorFlagVar(nodeCmd, &nodeGetOpts.LabelSelector)
	rootCmd.AddCommand(nodeCmd)
}

func showNodeResourceSummary(cmd *cobra.Command, args []string) error {
	var cancel context.CancelFunc

	ctx := kwcontext.New()
	cfg, err := kube.Config(ctx)
	if err != nil {
		return err
	}
	conn, err := kconnect.Connect(cfg)
	if err != nil {
		return err
	}
	kwcontext.RegisterConnection(ctx, conn)

	ctx, cancel = context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	nodes, err := knode.Get(ctx, conn, &nodeGetOpts)
	if err != nil {
		return err
	}

	switch outputFormat {
	case outputFormatHuman:
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoMergeCells(true)
		table.SetHeader([]string{"NODE", "RESOURCE", "CAPACITY", "RESERVED", "REQUESTED", "USED"})
		table.SetColumnAlignment([]int{
			tablewriter.ALIGN_LEFT,
			tablewriter.ALIGN_LEFT,
			tablewriter.ALIGN_RIGHT,
			tablewriter.ALIGN_RIGHT,
			tablewriter.ALIGN_RIGHT,
			tablewriter.ALIGN_RIGHT,
		})
		for _, node := range nodes {
			cpu := node.Resources.CPU
			cpuUsedPct := (cpu.Used / cpu.Allocatable) * 100
			cpuUsed := fmt.Sprintf("%.0f (%.2f%%)", cpu.Used, cpuUsedPct)

			data := []string{
				node.Name,
				"CPU",
				fmt.Sprintf("%.2f", cpu.Allocatable),
				fmt.Sprintf("%.2f", cpu.Reserved),
				fmt.Sprintf("%.2f", cpu.RequestedFloor),
				cpuUsed,
			}
			fieldColors := fieldColorsByPct(cpuUsedPct)
			table.Rich(data, fieldColors)

			mem := node.Resources.Memory
			memUsedPct := (mem.Used / mem.Allocatable) * 100
			memUsed := fmt.Sprintf(
				"%s (%.2f%%)",
				unit.BytesToSizeString(mem.Used),
				memUsedPct,
			)

			data = []string{
				node.Name,
				"Memory",
				unit.BytesToSizeString(mem.Allocatable),
				unit.BytesToSizeString(mem.Reserved),
				unit.BytesToSizeString(mem.RequestedFloor),
				memUsed,
			}
			fieldColors = fieldColorsByPct(cpuUsedPct)
			table.Rich(data, fieldColors)

			pod := node.Resources.Pods
			podUsedPct := (pod.Used / pod.Allocatable) * 100
			podUsed := fmt.Sprintf("%.0f (%.2f%%)", pod.Used, podUsedPct)

			data = []string{
				node.Name,
				"Pods",
				fmt.Sprintf("%.0f", pod.Allocatable),
				fmt.Sprintf("%.0f", pod.Reserved),
				fmt.Sprintf("%.0f", pod.RequestedFloor),
				podUsed,
			}
			fieldColors = fieldColorsByPct(podUsedPct)
			table.Rich(data, fieldColors)
		}
		table.Render()
	}
	return nil
}

func fieldColorsByPct(pct float64) []tablewriter.Colors {
	if pct > float64(85) {
		return []tablewriter.Colors{
			tablewriter.Colors{},
			twColorRedNormal,
			twColorRedNormal,
			twColorRedNormal,
			twColorRedNormal,
		}
	} else if pct > float64(75) {
		return []tablewriter.Colors{
			tablewriter.Colors{},
			twColorYellowNormal,
			twColorYellowNormal,
			twColorYellowNormal,
			twColorYellowNormal,
		}
	} else {
		return []tablewriter.Colors{
			tablewriter.Colors{},
			twColorGreenNormal,
			twColorGreenNormal,
			twColorGreenNormal,
			twColorGreenNormal,
		}
	}
}
