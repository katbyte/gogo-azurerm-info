package provider

import (
	"fmt"
	"os"
	"regexp"
)

type DataSource struct {
	ResourceOrData
}

func (ds DataSource) GetTotal() Totals {
	t := ds.ResourceOrData.GetTotal()
	t.DataSources++
	return t
}

func (s *Service) ScanDataSources() error {
	files, err := os.ReadDir(s.Path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", s.Path, err)
	}

	resourceFileRegex := regexp.MustCompile("[a-z_]+_data_source.go$")

	for _, f := range files {
		name := f.Name()

		if !resourceFileRegex.MatchString(name) {
			continue
		}

		bytes, err := os.ReadFile(s.Path + "/" + f.Name())
		if err != nil {
			return fmt.Errorf("reading %s: %w", f.Name(), err)
		}
		content := string(bytes)

		r := DataSource{
			ResourceOrData: s.GetResourceOrDataFor(f, content),
		}

		s.DataSources = append(s.DataSources, r)
	}

	return nil
}
