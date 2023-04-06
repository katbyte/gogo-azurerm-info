package cli

import (
	"fmt"
	"time"

	c "github.com/gookit/color" // nolint:misspell
	"github.com/katbyte/gogo-azurerm-info/lib/provider"
	"github.com/spf13/cobra"
)

func CmdList(_ *cobra.Command, args []string) error {
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

	switch args[1] {
	case "track1":
		ListTrack1(v)
	case "typed":
		ListTyped(v)
	case "create-update":
		ListSharedCreateUpdate(v)
	default:
		return fmt.Errorf("unknown list type '%s'", args[1])
	}

	return nil
}

func ListTrack1(v provider.Version) {
	total := 0
	toMigrate := 0
	for _, s := range v.Services {
		t := s.CalculateTotals()
		total += t.Resources
		total += t.DataSources

		if t.SdkTrack1 == 0 {
			continue
		}

		toMigrate += t.SdkTrack1

		c.Printf(" <cyan>%s</> (<lightMagenta>%d</>/<magenta>%d</> using track1)\n", s.Name, t.SdkTrack1, t.Resources+t.DataSources)
		rds := s.FilterResourcesDatas(func(rds provider.ResourceOrData) bool {
			return rds.SdkTrack1
		})

		for _, r := range rds {
			if r.SdkPandora {
				c.Printf("    <gray>%s/</>%s <yellow>(partial)</>\n", r.Service.Path, r.GoFileName)
			} else if r.SdkTrack1 {
				c.Printf("    <gray>%s/</>%s \n", r.Service.Path, r.GoFileName)
			}
		}

		fmt.Println()
	}

	fmt.Println()
	fmt.Println()

	c.Printf("<red>%d</>/<yellow>%d</> resources and data sources  still using track1\n", toMigrate, total)
}

func ListTyped(v provider.Version) {
	total := 0
	toMigrate := 0
	for _, s := range v.Services {
		t := s.CalculateTotals()
		total += t.Resources
		total += t.DataSources

		eTotal := s.CountResourcesDataSources()

		if t.Typed == eTotal {
			continue
		}

		toMigrate += eTotal - t.Typed

		c.Printf(" <cyan>%s</> (<lightMagenta>%d</> not typed)\n", s.Name, eTotal-t.Typed)

		rds := s.FilterResourcesDatas(func(rds provider.ResourceOrData) bool {
			return !rds.IsTyped
		})

		for _, r := range rds {
			c.Printf("    <gray>%s/</>%s \n", r.Service.Path, r.GoFileName)
		}

		fmt.Println()
	}

	fmt.Println()
	fmt.Println()

	c.Printf("<red>%d</>/<yellow>%d</> resources and data sources to migrated to being typted\n", toMigrate, total)
}

func ListSharedCreateUpdate(v provider.Version) {
	total := 0
	toMigrate := 0
	for _, s := range v.Services {
		t := s.CalculateTotals()
		total += t.Resources
		total += t.DataSources

		if t.CreateUpdate == 0 {
			continue
		}

		toMigrate += t.CreateUpdate

		c.Printf(" <cyan>%s</> (<lightMagenta>%d</> is sharing a create/update function)\n", s.Name, t.CreateUpdate)

		rds := s.FilterResourcesDatasInterfaced(func(rds interface{}) bool {
			if r, ok := rds.(provider.Resource); ok {
				return r.SharedCreateUpdate
			}
			return false
		})

		for _, r := range rds {
			c.Printf("    <gray>%s/</>%s \n", r.Service.Path, r.GoFileName)
		}

		fmt.Println()
	}

	fmt.Println()
	fmt.Println()

	c.Printf("<red>%d</>/<yellow>%d</> resources that need their shared create/update function split\n", toMigrate, total)
}
