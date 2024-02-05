package provider

type Totals struct {
	Services     int
	Resources    int
	DataSources  int
	SdkTrack1    int
	SdkPandora   int
	SdkKermit    int
	SdkGiovanni  int
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
	t.SdkKermit += t2.SdkKermit
	t.SdkGiovanni += t2.SdkGiovanni
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

	if rds.SdkAzureSdkGo {
		t.SdkTrack1++
	}

	if rds.SdkPandora {
		t.SdkPandora++
	}

	if rds.SdkKermit {
		t.SdkKermit++
	}

	if rds.SdkGiovanni {
		t.SdkGiovanni++
	}

	if rds.SdkKermit && rds.SdkAzureSdkGo {
		t.SdkTrack1++
	}

	if rds.SdkPandora && (rds.SdkAzureSdkGo || rds.SdkKermit) {
		t.SdkBoth++
	}

	if rds.UsesBuiltInParse {
		t.BuiltInParse++
	}

	return t
}
