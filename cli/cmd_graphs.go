package cli

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	c "github.com/gookit/color" // nolint:misspell
	"github.com/hashicorp/go-version"
	"github.com/katbyte/gogo-azurerm-info/lib/provider"
	"github.com/spf13/cobra"
)

func CmdGraphs(_ *cobra.Command, args []string) error {
	repoPath := args[0]

	tillTag := "v2.10.0" // first version using service packages
	if len(args) > 1 {
		tillTag = args[1]
	}

	// todo make configurable
	outPath := "graphs"
	err := os.MkdirAll(outPath, 0755)
	if err != nil {
		return fmt.Errorf("making path %s: %w", outPath, err)
	}

	c.Printf("Scanning <cyan>%s</>... ", repoPath)
	r, err := provider.NewRepo(args[0])
	if err != nil {
		return fmt.Errorf("opening repo: %w", err)
	}

	versions, err := r.GetVersions()
	if err != nil {
		return fmt.Errorf("getting versions for %s: %w", repoPath, err)
	}
	c.Printf("found <green>%d</> versions\n", len(*versions))

	versionsToGraph := []provider.Version{}
	for _, v := range *versions {
		// skip x.x.1 versions
		if !strings.HasSuffix(v.Name, ".0") {
			c.Printf("  checking out <green>%s</>... <red>HOTFIX</> skipped\n", v.Name)
			continue
		}

		// version 3.1.0 is broken so skip it
		//		if v.Name == "v3.1.0" || v.Name == "v3.0.0" {
		//			c.Printf("  checking out <green>%s</>... <red>BROKEN/FAILS</> skipped\n", v.Name)
		//			continue
		//		}

		c.Printf("  checking out <green>%s</>...", v.Name)
		err := r.CheckoutTag(v.Name)
		if err != nil {
			return fmt.Errorf("checking out: %w", err)
		}

		err = v.ScanServices()
		if err != nil {
			return fmt.Errorf("scanning services: %w", err)
		}

		t := v.CalculateTotals()

		// versions = append(versions, version)
		c.Printf(" <magenta>%d</> services, <cyan>%d</> resources and <lightBlue>%d</> data sources\n", len(v.Services), t.Resources, t.DataSources)

		versionsToGraph = append(versionsToGraph, v)
		if v.Name == tillTag {
			break
		}
	}

	sort.Slice(versionsToGraph, func(i, j int) bool {
		vi, _ := version.NewVersion(versionsToGraph[i].Name)
		vj, _ := version.NewVersion(versionsToGraph[j].Name)
		return vj.GreaterThan(vi)
	})

	// genreate graphs
	if err = GraphsResourcesDataSourcesOverTime(&versionsToGraph, outPath); err != nil {
		return fmt.Errorf("charting resources and data sources: %w", err)
	}
	if err = GraphsPandoraSDKMigration(&versionsToGraph, outPath); err != nil {
		return fmt.Errorf("charting pandora migration: %w", err)
	}

	return nil
}

func GraphsResourcesDataSourcesOverTime(versions *[]provider.Version, outPath string) error {
	var xAxis []string
	var resources, dataSources []opts.LineData

	var data [][]string
	data = append(data, []string{"version", "services", "resources", "resources", "data-sources"})
	for _, v := range *versions {
		t := v.CalculateTotals()

		// todo add a 2nd axis for services

		xAxis = append(xAxis, v.Name)
		resources = append(resources, opts.LineData{Value: t.Resources})
		dataSources = append(dataSources, opts.LineData{Value: t.DataSources})

		data = append(data,
			[]string{v.Name,
				strconv.Itoa(t.Services),
				strconv.Itoa(t.Resources),
				strconv.Itoa(t.DataSources),
			})
	}

	// write raw data
	file, err := os.Create(outPath + "/resources-data-sources.csv")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	csv := csv.NewWriter(file)
	defer csv.Flush()

	for _, r := range data {
		err := csv.Write(r)
		if err != nil {
			panic(err)
		}
	}

	// render graph
	graph := charts.NewLine()
	graph.SetGlobalOptions(
		// charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWesteros}),
		charts.WithTitleOpts(opts.Title{
			Title: "Resources and Data Sources ",
			Left:  "center"}), // nolint:misspell

		charts.WithXAxisOpts(opts.XAxis{
			Name: "Version",
			// AxisLabel: &opts.AxisLabel{Show: true, Formatter: "{value} x-unit"},
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "Total",
			// AxisLabel: &opts.AxisLabel{Show: true, Formatter: "{value} x-unit"},
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "1500px",
			Height: "750px",
		}),
		charts.WithColorsOpts(opts.Colors{"#2E4555", "#62A0A8", "#C13530"}),
		charts.WithToolboxOpts(opts.Toolbox{Show: true}),
		charts.WithLegendOpts(opts.Legend{
			Show: true,
			Top:  "bottom",
			Left: "center", // nolint:misspell
		}),
	)

	// Put data into instance
	graph.SetXAxis(xAxis).
		AddSeries("Resources", resources).
		AddSeries("Data Sources", dataSources).
		SetSeriesOptions(charts.WithAreaStyleOpts(opts.AreaStyle{
			Opacity: 0.7,
		}),
			charts.WithLineChartOpts(opts.LineChart{
				Stack: "elements",
			}))

	// Where the magic happens
	file, err = os.Create(outPath + "/resources-data-sources.html")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	err = graph.Render(file)
	if err != nil {
		return fmt.Errorf("failed to render graph graph: %w", err)
	}

	return nil
}

func GraphsPandoraSDKMigration(versions *[]provider.Version, outPath string) error {
	var xAxis []string
	var total, resourcesPandora, dataSourcesPandora []opts.LineData

	var curTotal, curDone int
	var ver string

	var data [][]string
	data = append(data, []string{"version", "services", "resources", "resources-pandora", "data-sources", "data-sources-pandora"})
	for _, v := range *versions {
		t := v.CalculateTotals()
		tr := v.CalculateResourceTotals()
		td := v.CalculateDataSourceTotals()

		// todo add a 2nd axis for services ??

		xAxis = append(xAxis, v.Name)
		total = append(total, opts.LineData{Value: t.Resources + t.DataSources})
		resourcesPandora = append(resourcesPandora, opts.LineData{Value: tr.SdkPandora})
		dataSourcesPandora = append(dataSourcesPandora, opts.LineData{Value: td.SdkPandora})

		// just keep doing this and the last one will be the most recent
		curTotal = t.Resources + t.DataSources
		curDone = t.SdkPandora
		ver = v.Name

		data = append(data,
			[]string{v.Name,
				strconv.Itoa(t.Services),
				strconv.Itoa(t.Resources),
				strconv.Itoa(tr.SdkPandora),
				strconv.Itoa(t.DataSources),
				strconv.Itoa(td.SdkPandora),
			})
	}

	// write raw data
	file, err := os.Create(outPath + "/pandora-sdk-migration.csv")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	csv := csv.NewWriter(file)
	defer csv.Flush()

	for _, r := range data {
		err := csv.Write(r)
		if err != nil {
			panic(err)
		}
	}

	// render graph
	graph := charts.NewLine()
	graph.SetGlobalOptions(
		// charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWesteros}),
		charts.WithTitleOpts(opts.Title{
			Title:    "Pandora SDK Migration",
			Subtitle: fmt.Sprintf("%d/%d (%00.00f%%) done as of %s", curDone, curTotal, float32(curDone)/float32(curTotal)*100, ver),
			Left:     "center"}), // nolint:misspell

		charts.WithXAxisOpts(opts.XAxis{
			Name: "Version",
			// AxisLabel: &opts.AxisLabel{Show: true, Formatter: "{value} x-unit"},
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "Total",
			// AxisLabel: &opts.AxisLabel{Show: true, Formatter: "{value} x-unit"},
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "1500px",
			Height: "750px",
		}),
		charts.WithColorsOpts(opts.Colors{"#000000", "#2E4555", "#62A0A8"}),
		charts.WithToolboxOpts(opts.Toolbox{Show: true}),
		charts.WithLegendOpts(opts.Legend{
			Show: true,
			Top:  "bottom",
			Left: "center", // nolint:misspell
		}),
	)

	// TODO put back in total resources line

	// Put data into instance
	graph.SetXAxis(xAxis).
		AddSeries("Total Resources/DataSources", total,
			charts.WithAreaStyleOpts(opts.AreaStyle{Opacity: 0.001})).
		AddSeries("Resources Migrated", resourcesPandora,
			charts.WithAreaStyleOpts(opts.AreaStyle{Opacity: 1.0}),
			charts.WithLineChartOpts(opts.LineChart{Stack: "elements"})).
		AddSeries("Data Sources Migrated", dataSourcesPandora,
			charts.WithAreaStyleOpts(opts.AreaStyle{Opacity: 1.0}),
			charts.WithLineChartOpts(opts.LineChart{Stack: "elements"}))

	// Where the magic happens
	file, err = os.Create(outPath + "/pandora-sdk-migration.html")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	err = graph.Render(file)
	if err != nil {
		return fmt.Errorf("failed to render graph graph: %w", err)
	}

	return nil
}
