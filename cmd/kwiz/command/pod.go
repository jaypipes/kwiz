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

	kwcontext "github.com/jaypipes/kwiz/pkg/context"
	"github.com/jaypipes/kwiz/pkg/kube"
	kconnect "github.com/jaypipes/kwiz/pkg/kube/connect"
	kpod "github.com/jaypipes/kwiz/pkg/kube/pod"
)

// podCmd represents the node command
var podCmd = &cobra.Command{
	Use:   "pod",
	Short: "Show pod resource summary",
	RunE:  showPodResourceSummary,
}

func init() {
	rootCmd.AddCommand(podCmd)
}

func showPodResourceSummary(cmd *cobra.Command, args []string) error {
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

	pods, err := kpod.Get(ctx, conn)
	if err != nil {
		return err
	}
	colors := []tablewriter.Colors{
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
		tablewriter.Colors{},
	}

	switch outputFormat {
	case outputFormatHuman:
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoMergeCellsByColumnIndex([]int{0, 1})
		table.SetHeader([]string{"POD", "NAMESPACE", "RESOURCE", "Req", "Lim"})
		table.SetColumnAlignment([]int{
			tablewriter.ALIGN_LEFT,
			tablewriter.ALIGN_LEFT,
			tablewriter.ALIGN_RIGHT,
			tablewriter.ALIGN_RIGHT,
			tablewriter.ALIGN_RIGHT,
			tablewriter.ALIGN_RIGHT,
		})
		for _, pod := range pods {
			cpu := pod.ResourceRequests.CPU
			cpuFloor := fmt.Sprintf("%.0f", cpu.Floor)
			if cpu.Floor == float64(0) {
				cpuFloor = "-"
			}
			cpuCeiling := fmt.Sprintf("%.0f", cpu.Ceiling)
			if cpu.Ceiling == float64(-1) {
				cpuCeiling = "-"
			}
			data := []string{
				pod.Name,
				pod.Namespace,
				"CPU",
				cpuFloor,
				cpuCeiling,
			}
			table.Rich(data, colors)

			mem := pod.ResourceRequests.Memory
			memFloor := fmt.Sprintf("%.0f", mem.Floor)
			if mem.Floor == float64(0) {
				memFloor = "-"
			}
			memCeiling := fmt.Sprintf("%.0f", mem.Ceiling)
			if mem.Ceiling == float64(-1) {
				memCeiling = "-"
			}
			data = []string{
				pod.Name,
				pod.Namespace,
				"Memory",
				memFloor,
				memCeiling,
			}
			table.Rich(data, colors)
		}
		table.Render()
	}
	return nil
}
