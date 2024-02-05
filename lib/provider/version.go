package provider

import (
	"fmt"
	"os"
	"regexp"
	"time"
)

type Version struct {
	Name string
	Date time.Time
	Path string

	Services []Service
}

func (v *Version) ScanServices() error {
	path := v.Path + "/internal/services"

	// service folder location changed in v3.1.0
	oldServicesPathRegex := regexp.MustCompile("v2.[123456]")
	oldServicesPathMap := map[string]bool{"v2.71.0": true, "v2.70.0": true} // todo get this into the regex pattern
	if _, ok := oldServicesPathMap[v.Name]; ok || oldServicesPathRegex.MatchString(v.Name) {
		path = v.Path + "/azurerm/internal/services"
	}

	// find all services
	folders, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	for _, f := range folders {
		s := Service{
			Name: f.Name(),
			Path: path + "/" + f.Name(),
		}

		// scan
		err = s.ScanResources()
		if err != nil {
			return fmt.Errorf("scanning resources for %s: %w", s.Name, err)
		}
		err = s.ScanDataSources()
		if err != nil {
			return fmt.Errorf("scanning data sources for %s: %w", s.Name, err)
		}

		/*s.ScanClients()
		if err != nil {
			return fmt.Errorf("scanning clients for %s: %w", s.Name, err)
		}
		*/
		v.Services = append(v.Services, s)
	}

	return nil
}

func (v *Version) CalculateTotals() Totals {
	totals := Totals{}
	for _, s := range v.Services {
		totals = totals.Add(s.CalculateTotals())
	}
	return totals
}

func (v *Version) CalculateResourceTotals() Totals {
	totals := Totals{}
	for _, s := range v.Services {
		totals = totals.Add(s.CalculateResourceTotals())
	}
	return totals
}

func (v *Version) CalculateDataSourceTotals() Totals {
	totals := Totals{}
	for _, s := range v.Services {
		totals = totals.Add(s.CalculateDataSourceTotals())
	}
	return totals
}
