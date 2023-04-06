package provider

type Totals struct {
	Services     int
	Resources    int
	DataSources  int
	SdkTrack1    int
	SdkPandora   int
	SdkBoth      int
	Typed        int
	CreateUpdate int
	BuiltInParse int
}

func (t Totals) Add(t2 Totals) Totals {
	t.Services += t2.Services
	t.Resources += t2.Resources
	t.DataSources += t2.DataSources
	t.SdkTrack1 += t2.SdkTrack1
	t.SdkPandora += t2.SdkPandora
	t.SdkBoth += t2.SdkBoth
	t.Typed += t2.Typed
	t.BuiltInParse += t2.BuiltInParse
	t.CreateUpdate += t2.CreateUpdate
	return t
}

func (rds ResourceOrData) GetTotal() Totals {
	t := Totals{}

	if rds.IsTyped {
		t.Typed++
	}

	if rds.SdkTrack1 {
		t.SdkTrack1++
	}

	if rds.SdkPandora {
		t.SdkPandora++
	}

	if rds.SdkPandora && rds.SdkTrack1 {
		t.SdkBoth++
	}

	if rds.UsesBuiltInParse {
		t.BuiltInParse++
	}

	return t
}
