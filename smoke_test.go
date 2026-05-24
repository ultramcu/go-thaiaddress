package thaiaddress

import "testing"

func TestSmoke(t *testing.T) {
	if got := len(Provinces()); got != 77 {
		t.Fatalf("provinces=%d want 77", got)
	}
	if got := len(Districts()); got != 930 {
		t.Fatalf("districts=%d want 930", got)
	}
	if got := len(Subdistricts()); got != 7452 {
		t.Fatalf("subdistricts=%d want 7452", got)
	}
	bkk, ok := ProvinceByCode(10)
	if !ok || bkk.NameEN != "Bangkok" || bkk.Region != RegionCentral {
		t.Fatalf("bkk=%+v ok=%v", bkk, ok)
	}
	d, ok := DistrictByCode(1001)
	if !ok || d.ProvinceCode != 10 {
		t.Fatalf("district 1001=%+v ok=%v", d, ok)
	}
	s, ok := SubdistrictByCode(100101)
	if !ok || s.DistrictCode != 1001 || s.Postcode != 10200 {
		t.Fatalf("sub 100101=%+v ok=%v", s, ok)
	}
	if p, ok := s.Province(); !ok || p.Code != 10 {
		t.Fatalf("sub->prov=%+v ok=%v", p, ok)
	}
	if subs := ByPostcode(10200); len(subs) == 0 {
		t.Fatal("ByPostcode(10200) empty")
	}
	// referential integrity: every district resolves to a province, every sub to a district
	for _, d := range Districts() {
		if _, ok := ProvinceByCode(d.ProvinceCode); !ok {
			t.Fatalf("district %d orphan province %d", d.Code, d.ProvinceCode)
		}
	}
	for _, s := range Subdistricts() {
		if _, ok := DistrictByCode(s.DistrictCode); !ok {
			t.Fatalf("sub %d orphan district %d", s.Code, s.DistrictCode)
		}
	}
}
