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
	"github.com/jaypipes/kwiz/pkg/types"
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

	resourceTotals := types.Resources{}
	for _, node := range nodes {
		cpu := node.Resources.CPU
		resourceTotals.CPU.Capacity += cpu.Capacity
		resourceTotals.CPU.Allocatable += cpu.Allocatable
		resourceTotals.CPU.Reserved += cpu.Reserved
		resourceTotals.CPU.RequestedFloor += cpu.RequestedFloor
		// If there is any Pod on the Node that has no limits set for this
		// resource, it can potentially consume all of the resource on the
		// Node. So, we treat ceiling == -1 specially.
		if cpu.RequestedCeiling == -1 || resourceTotals.CPU.RequestedCeiling == -1 {
			resourceTotals.CPU.RequestedCeiling = -1
		} else {
			resourceTotals.CPU.RequestedCeiling += cpu.RequestedCeiling
		}
		resourceTotals.CPU.Used += cpu.Used
		mem := node.Resources.Memory
		resourceTotals.Memory.Capacity += mem.Capacity
		resourceTotals.Memory.Allocatable += mem.Allocatable
		resourceTotals.Memory.Reserved += mem.Reserved
		resourceTotals.Memory.RequestedFloor += mem.RequestedFloor
		if mem.RequestedCeiling == -1 || resourceTotals.Memory.RequestedCeiling == -1 {
			resourceTotals.Memory.RequestedCeiling = -1
		} else {
			resourceTotals.Memory.RequestedCeiling += mem.RequestedCeiling
		}
		resourceTotals.Memory.Used += mem.Used
		pods := node.Resources.Pods
		resourceTotals.Pods.Capacity += pods.Capacity
		resourceTotals.Pods.Allocatable += pods.Allocatable
		resourceTotals.Pods.Reserved += pods.Reserved
		resourceTotals.Pods.RequestedFloor += pods.RequestedFloor
		resourceTotals.Pods.RequestedCeiling += pods.RequestedCeiling
		resourceTotals.Pods.Used += pods.Used
	}

	maxNodeNameLen := 0

	switch outputFormat {
	case outputFormatHuman:
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoMergeCells(true)
		table.SetBorders(tablewriter.Border{Left: false, Right: false, Bottom: true, Top: true})
		table.SetHeader([]string{"NODE", "RESOURCE", "CAPACITY", "RESERVED", "REQ FLOOR", "REQ CEIL", "ACTUAL"})
		table.SetColumnAlignment([]int{
			tablewriter.ALIGN_LEFT,
			tablewriter.ALIGN_LEFT,
			tablewriter.ALIGN_RIGHT,
			tablewriter.ALIGN_RIGHT,
			tablewriter.ALIGN_RIGHT,
			tablewriter.ALIGN_RIGHT,
			tablewriter.ALIGN_RIGHT,
		})
		for _, node := range nodes {
			cpu := node.Resources.CPU
			cpuFloorPct := (cpu.RequestedFloor / cpu.Allocatable) * 100
			cpuFloorStr := fmt.Sprintf("%.0f (%.2f%%)", cpu.RequestedFloor, cpuFloorPct)
			cpuCeil := cpu.RequestedCeiling
			if cpuCeil == -1 {
				// If any Pod has no limits, that means it can consume all of
				// the node's resources...
				cpuCeil = cpu.Allocatable
			}
			cpuCeilPct := (cpuCeil / cpu.Allocatable) * 100
			cpuCeilStr := fmt.Sprintf("%.0f (%.2f%%)", cpuCeil, cpuCeilPct)
			cpuUsedPct := (cpu.Used / cpu.Allocatable) * 100
			cpuUsedStr := fmt.Sprintf("%.0f (%.2f%%)", cpu.Used, cpuUsedPct)

			data := []string{
				node.Name,
				"CPU",
				fmt.Sprintf("%.2f", cpu.Allocatable),
				fmt.Sprintf("%.2f", cpu.Reserved),
				cpuFloorStr,
				cpuCeilStr,
				cpuUsedStr,
			}
			fieldColors := fieldColorsByPct(cpuFloorPct, cpuCeilPct, cpuUsedPct)
			table.Rich(data, fieldColors)

			mem := node.Resources.Memory
			memFloorPct := (mem.RequestedFloor / mem.Allocatable) * 100
			memFloorStr := fmt.Sprintf("%s (%.2f%%)", unit.BytesToSizeString(mem.RequestedFloor), memFloorPct)
			memCeil := mem.RequestedCeiling
			if memCeil == -1 {
				// If any Pod has no limits, that means it can consume all of
				// the node's resources...
				memCeil = mem.Allocatable
			}
			memCeilPct := (memCeil / mem.Allocatable) * 100
			memCeilStr := fmt.Sprintf("%s (%.2f%%)", unit.BytesToSizeString(memCeil), memCeilPct)
			memUsedPct := (mem.Used / mem.Allocatable) * 100
			memUsedStr := fmt.Sprintf(
				"%s (%.2f%%)",
				unit.BytesToSizeString(mem.Used),
				memUsedPct,
			)

			data = []string{
				node.Name,
				"Memory",
				unit.BytesToSizeString(mem.Allocatable),
				unit.BytesToSizeString(mem.Reserved),
				memFloorStr,
				memCeilStr,
				memUsedStr,
			}
			fieldColors = fieldColorsByPct(memFloorPct, memCeilPct, memUsedPct)
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
				fmt.Sprintf("%.0f", pod.RequestedCeiling),
				podUsed,
			}
			maxNodeNameLen = max(maxNodeNameLen, len(node.Name))
			fieldColors = fieldColorsByPct(podUsedPct, podUsedPct, podUsedPct)
			table.Rich(data, fieldColors)
		}
		table.Render()

		// Print out the totals table as a separate entity
		totalsFormatStr := fmt.Sprintf("%%%ds", maxNodeNameLen)
		totTable := tablewriter.NewWriter(os.Stdout)
		totTable.SetHeader([]string{"", "RESOURCE", "CAPACITY", "RESERVED", "REQ FLOOR", "REQ CEIL", "ACTUAL"})
		totTable.SetAutoMergeCells(true)
		totTable.SetBorders(tablewriter.Border{Left: false, Right: false, Bottom: true, Top: false})
		totTable.SetColumnAlignment([]int{
			tablewriter.ALIGN_LEFT,
			tablewriter.ALIGN_RIGHT,
			tablewriter.ALIGN_RIGHT,
			tablewriter.ALIGN_RIGHT,
			tablewriter.ALIGN_RIGHT,
			tablewriter.ALIGN_RIGHT,
			tablewriter.ALIGN_RIGHT,
		})

		cpu := resourceTotals.CPU
		cpuFloorPct := (cpu.RequestedFloor / cpu.Allocatable) * 100
		cpuFloorStr := fmt.Sprintf("%.0f (%.2f%%)", cpu.RequestedFloor, cpuFloorPct)
		cpuCeil := cpu.RequestedCeiling
		if cpuCeil == -1 {
			// If any Pod has no limits, that means it can consume all of
			// the node's resources...
			cpuCeil = cpu.Allocatable
		}
		cpuCeilPct := (cpuCeil / cpu.Allocatable) * 100
		cpuCeilStr := fmt.Sprintf("%.0f (%.2f%%)", cpuCeil, cpuCeilPct)
		cpuUsedPct := (cpu.Used / cpu.Allocatable) * 100
		cpuUsedStr := fmt.Sprintf("%.0f (%.2f%%)", cpu.Used, cpuUsedPct)

		data := []string{
			fmt.Sprintf(totalsFormatStr, "Totals"),
			"CPU",
			fmt.Sprintf("%.2f", cpu.Allocatable),
			fmt.Sprintf("%.2f", cpu.Reserved),
			cpuFloorStr,
			cpuCeilStr,
			cpuUsedStr,
		}
		fieldColors := fieldColorsByPct(cpuFloorPct, cpuCeilPct, cpuUsedPct)
		totTable.Rich(data, fieldColors)

		mem := resourceTotals.Memory
		memFloorPct := (mem.RequestedFloor / mem.Allocatable) * 100
		memFloorStr := fmt.Sprintf("%s (%.2f%%)", unit.BytesToSizeString(mem.RequestedFloor), memFloorPct)
		memCeil := mem.RequestedCeiling
		if memCeil == -1 {
			// If any Pod has no limits, that means it can consume all of
			// the node's resources...
			memCeil = mem.Allocatable
		}
		memCeilPct := (memCeil / mem.Allocatable) * 100
		memCeilStr := fmt.Sprintf("%s (%.2f%%)", unit.BytesToSizeString(memCeil), memCeilPct)
		memUsedPct := (mem.Used / mem.Allocatable) * 100
		memUsedStr := fmt.Sprintf(
			"%s (%.2f%%)",
			unit.BytesToSizeString(mem.Used),
			memUsedPct,
		)

		data = []string{
			fmt.Sprintf(totalsFormatStr, "Totals"),
			"Memory",
			unit.BytesToSizeString(mem.Allocatable),
			unit.BytesToSizeString(mem.Reserved),
			memFloorStr,
			memCeilStr,
			memUsedStr,
		}
		fieldColors = fieldColorsByPct(memFloorPct, memCeilPct, memUsedPct)
		totTable.Rich(data, fieldColors)

		pod := resourceTotals.Pods
		podUsedPct := (pod.Used / pod.Allocatable) * 100
		podUsed := fmt.Sprintf("%.0f (%.2f%%)", pod.Used, podUsedPct)

		data = []string{
			fmt.Sprintf(totalsFormatStr, "Totals"),
			"Pods",
			fmt.Sprintf("%.0f", pod.Allocatable),
			fmt.Sprintf("%.0f", pod.Reserved),
			fmt.Sprintf("%.0f", pod.RequestedFloor),
			fmt.Sprintf("%.0f", pod.RequestedCeiling),
			podUsed,
		}
		fieldColors = fieldColorsByPct(podUsedPct, podUsedPct, podUsedPct)
		totTable.Rich(data, fieldColors)

		totTable.Render()
	}
	return nil
}

func fieldColorsByPct(floorPct, ceilPct, usedPct float64) []tablewriter.Colors {
	colors := []tablewriter.Colors{
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
		twColorGreenNormal,
		twColorGreenNormal,
		twColorGreenNormal,
	}
	if floorPct > float64(85) {
		colors[4] = twColorRedNormal
	} else if floorPct > float64(75) {
		colors[4] = twColorYellowNormal
	}
	if ceilPct > float64(85) {
		colors[5] = twColorRedNormal
	} else if ceilPct > float64(75) {
		colors[5] = twColorYellowNormal
	}
	if usedPct > float64(85) {
		colors[6] = twColorRedNormal
	} else if usedPct > float64(75) {
		colors[6] = twColorYellowNormal
	}
	return colors
}
