package platform

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/superfly/flyctl/api"
	"github.com/superfly/flyctl/client"
	"github.com/superfly/flyctl/internal/command"
	"github.com/superfly/flyctl/internal/config"
	"github.com/superfly/flyctl/internal/render"
	"github.com/superfly/flyctl/iostreams"
)

func newVMSizes() (cmd *cobra.Command) {
	const (
		long = `View a list of VM sizes which can be used with the FLYCTL SCALE VM command
`
		short = "List VM Sizes"
	)

	cmd = command.New("vm-sizes", short, long, runVMSizes,
		command.RequireSession,
	)

	cmd.Args = cobra.NoArgs

	return
}

func runVMSizes(ctx context.Context) error {
	client := client.FromContext(ctx).API()

	sizes, err := client.PlatformVMSizes(ctx)
	if err != nil {
		return fmt.Errorf("failed retrieving sizes: %w", err)
	}

	out := iostreams.FromContext(ctx).Out
	if config.FromContext(ctx).JSONOutput {
		return render.JSON(out, sizes)
	}

	var rows [][]string
	for _, size := range sizes {
		rows = append(rows, []string{
			size.Name,
			cores(size.CPUCores),
			memory(size.MemoryMB),
		})
	}

	render.Table(out, "Nomad platform", rows, "Name", "CPU Cores", "Memory")

	return runMachineVMSizes(ctx)
}

func runMachineVMSizes(ctx context.Context) error {
	out := iostreams.FromContext(ctx).Out
	presets := api.MachinePresets

	var shared [][]string
	for key, guest := range presets {
		if guest.CPUKind != "shared" {
			continue
		}
		shared = append(shared, []string{
			key,
			cores(float32(guest.CPUs)),
			memory(guest.MemoryMB),
		})
	}

	sort.Slice(shared, func(i, j int) bool {
		return shared[j][1] > shared[i][1]
	})

	render.Table(out, "Machines platform", shared, "Name", "CPU Cores", "Memory")
	var performance [][]string
	for key, guest := range presets {
		if guest.CPUKind != "performance" {
			continue
		}
		performance = append(performance, []string{
			key,
			cores(float32(guest.CPUs)),
			memory(guest.MemoryMB),
		})
	}

	sort.Slice(performance, func(i, j int) bool {
		return performance[j][0] > performance[i][0]
	})

	return render.Table(out, "", performance, "Name", "CPU Cores", "Memory")
}

func cores(cores float32) string {
	if cores < 1.0 {
		return fmt.Sprintf("%.2f", cores)
	}
	return fmt.Sprintf("%d", int(cores))
}

func memory(size int) string {
	if size < 1024 {
		return fmt.Sprintf("%d MB", size)
	}
	return fmt.Sprintf("%d GB", int(size))
}
