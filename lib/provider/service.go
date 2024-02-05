package provider

import (
	"sort"
)

type Service struct {
	Name string
	Path string // should this just be a *Repo?

	Resources   []Resource
	DataSources []DataSource

	// Clients []Client
}

func (s *Service) CountResourcesDataSources() int {
	return len(s.Resources) + len(s.DataSources)
}

func (s *Service) CalculateTotals() Totals {
	t := s.CalculateDataSourceTotals().Add(s.CalculateResourceTotals())
	t.Services = 1
	return t
}

func (s *Service) CalculateResourceTotals() Totals {
	totals := Totals{Services: 1}
	for _, r := range s.Resources {
		totals = totals.Add(r.GetTotal())
	}
	return totals
}

func (s *Service) CalculateDataSourceTotals() Totals {
	totals := Totals{Services: 1}
	for _, ds := range s.DataSources {
		totals = totals.Add(ds.GetTotal())
	}
	return totals
}

func (s *Service) FilterResourcesDatasInterfaced(f func(rds interface{}) bool) []ResourceOrData {
	rds := []ResourceOrData{}

	for _, r := range s.Resources {
		if f(r.ResourceOrData) {
			rds = append(rds, r.ResourceOrData)
		}
	}

	for _, d := range s.DataSources {
		if f(d.ResourceOrData) {
			rds = append(rds, d.ResourceOrData)
		}
	}

	sort.Slice(rds, func(i, j int) bool {
		if rds[i].GoFileName < rds[j].GoFileName {
			return true
		} else {
			return false
		}
	})

	return rds
}

func (s *Service) FilterResourcesDatas(f func(rds ResourceOrData) bool) []ResourceOrData {
	return s.FilterResourcesDatasInterfaced(func(rds interface{}) bool {
		return f(rds.(ResourceOrData))
	})
}
