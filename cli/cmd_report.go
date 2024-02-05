package cli

import (
	"fmt"
	"time"

	c "github.com/gookit/color" // nolint:misspell
	"github.com/katbyte/gogo-azurerm-info/lib/provider"
	"github.com/spf13/cobra"
)

func CmdReport(_ *cobra.Command, args []string) error {
	repoPath := args[0]

	c.Printf("Scanning <cyan>%s</>... ", repoPath)

	v := provider.Version{
		Name: "main",
		Path: repoPath,
		Date: time.Time{},
	}

	err := v.ScanServices()
	if err != nil {
		return fmt.Errorf("scanning services: %w", err)
	}

	t := v.CalculateTotals()
	c.Printf(" <magenta>%d</> services with %d resources and %d data sources\n", len(v.Services), t.Resources, t.DataSources)

	if len(args) == 1 {
		ReportDefault(v)
		return nil
	}

	switch args[1] {
	case "pandora-sdk-issue":
		ReportPandoraSdkIssue(v)
	default:
		return fmt.Errorf("unknown report type '%s': %w", args[1], err)
	}

	//
	//  servicename: # resources #  data sources
	//     # track 1, # pandora, # mixed (% migrated)
	//     # typed, # migrated

	return nil
}

func ReportDefault(v provider.Version) {
	for _, s := range v.Services {
		t := s.CalculateTotals()

		eCount := len(s.Resources) + len(s.DataSources)

		// pandoraDone := (t.SdkPandora - t.SdkBoth) / eCount * 100

		// light green 100% migrated
		// light yellow partial
		// light red 0

		c.Printf(" <lightCyan>%s</> (<magenta>%d</> resources, <magenta>%d</> data sources)\n", s.Name, len(s.Resources), len(s.DataSources))

		if t.SdkBoth != 0 {
			c.Printf("    Pandora: %d / %d (%d partial)\n", t.SdkPandora-t.SdkBoth, eCount, t.SdkBoth)
		} else {
			c.Printf("    Pandora: %d / %d\n", t.SdkPandora-t.SdkBoth, eCount)
		}

		c.Printf("    Typed:   %d / %d\n", t.Typed, eCount)
		c.Printf("\n")
	}
}

func ReportPandoraSdkIssue(v provider.Version) {
	fmt.Println()
	fmt.Println("## Service Packages")
	fmt.Println()

	var servicesDone, servicesPartial, elementsTotal, elementsDone, elementsPartial int
	for _, s := range v.Services {
		t := s.CalculateTotals()

		eCount := len(s.Resources) + len(s.DataSources)

		done := false
		if t.SdkTrack1 == 0 && t.SdkBoth == 0 {
			done = true
			servicesDone++
		}

		if t.SdkBoth != 0 {
			servicesPartial++
		}

		elementsTotal += eCount
		elementsDone += eCount - t.SdkTrack1
		elementsPartial += t.SdkBoth

		if done {
			fmt.Printf("- [X] `%s` _(%d)_\n", s.Name, eCount)
		} else {
			fmt.Printf("- [ ] `%s` _(%d/%d)_\n", s.Name, eCount-t.SdkTrack1, eCount)
		}
	}

	fmt.Printf("services: %d of %d (+%d partial)\n", servicesDone, len(v.Services), servicesPartial)
	fmt.Printf("resources/datasources: %d of %d (-%d partial)\n", elementsDone, elementsTotal, elementsPartial)
}
