package provider

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Resource struct {
	ResourceOrData

	// resource specific
	SharedCreateUpdate bool
}

func (r Resource) GetTotal() Totals {
	t := r.ResourceOrData.GetTotal()
	t.Resources++

	if r.SharedCreateUpdate {
		t.CreateUpdate++
	}

	return t
}

func (s *Service) ScanResources() error {
	// find all services
	files, err := os.ReadDir(s.Path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", s.Path, err)
	}

	resourceFileRegex := regexp.MustCompile("[a-z_]+_resource.go$")
	// dataSourceFileRegex := regexp.MustCompile("[a-z_]+_data_source.go$")

	for _, f := range files {
		name := f.Name()

		if !resourceFileRegex.MatchString(name) {
			continue
		}

		// these are not resource files, skip
		skip := map[string]bool{
			"bot_service_base_resource.go":           true,
			"export_base_resource.go":                true,
			"assignment_base_resource.go":            true,
			"container_registry_migrate_resource.go": true,
			"resource_group_data_source_resource.go": true,
		}
		if _, ok := skip[name]; ok {
			continue
		}

		// skip older migration rsource files
		if strings.Contains(name, "migration_resource.go") ||
			strings.Contains(name, "migration_resource_test.go") ||
			strings.Contains(name, "migration_test_resource.go") {
			continue
		}

		bytes, err := os.ReadFile(s.Path + "/" + f.Name())
		if err != nil {
			return fmt.Errorf("reading %s: %w", f.Name(), err)
		}
		content := string(bytes)

		r := Resource{
			ResourceOrData: s.GetResourceOrDataFor(f, content),
		}

		// Shared Created/Update (only for plugin-sdk??)
		if !r.IsTyped {
			createFunctionRegex := regexp.MustCompile("Create: *[a-zA-Z0-9]+,")
			updateFunctionRegex := regexp.MustCompile("Update: *[a-zA-Z0-9]+,")

			creates := createFunctionRegex.FindAllString(content, -1)
			updates := updateFunctionRegex.FindAllString(content, -1)

			// sanity checks
			if len(creates) == 0 {
				return fmt.Errorf("matching 'Create:' for %s", r.GoFileName)
			}
			if len(creates) > 1 {
				return fmt.Errorf("found multiple 'Create:'s for %s: %s", r.GoFileName, strings.Join(creates, ", "))
			}
			if len(updates) > 1 {
				return fmt.Errorf("found multiple 'Update:'s for %s: %s", r.GoFileName, strings.Join(updates, ", "))
			}
			if len(updates) == 1 {
				createFunction := strings.Trim(strings.Split(creates[0], " ")[1], ",")
				updateFunction := strings.Trim(strings.Split(updates[0], " ")[1], ",")

				if createFunction == updateFunction {
					r.SharedCreateUpdate = true
				}
			}
		}
		s.Resources = append(s.Resources, r)
	}

	return nil
}
