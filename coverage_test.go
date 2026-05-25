package thaiaddress

// coverage_test.go fills the remaining coverage gaps identified by go tool cover:
//
//   data.go    SubdistrictsOf (0%), PostcodesOf (0%)
//   model.go   District.Province (0%), District.Subdistricts (0%), District.Postcodes (0%),
//              Subdistrict.District (0%), Subdistrict.Province (0%)
//   region.go  Region.Valid (0%), Region.NameTH (0%), Region.String (66% — unknown branch)
//   normalize  stripPrefix English-empty-rest branch (91.7%)
//   resolve.go Resolve (92.6% — postcode-only candidate path with district filter)
//
// All assertions are grounded in the real embedded dataset.

import (
	"errors"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// data.go: SubdistrictsOf and PostcodesOf
// ---------------------------------------------------------------------------

func TestSubdistrictsOf(t *testing.T) {
	// District 1001 = เขตพระนคร (Bangkok Phra Nakhon); it has 12 subdistricts
	// that all use postcode 10200.
	subs := SubdistrictsOf(1001)
	if len(subs) == 0 {
		t.Fatal("SubdistrictsOf(1001) returned nil, want non-empty")
	}
	// Verify first entry is 100101 (ordered by code).
	if subs[0].Code != 100101 {
		t.Errorf("SubdistrictsOf(1001)[0].Code = %d, want 100101", subs[0].Code)
	}
	// Every entry's DistrictCode must be 1001.
	for _, s := range subs {
		if s.DistrictCode != 1001 {
			t.Errorf("SubdistrictsOf(1001): sub %d has DistrictCode %d, want 1001", s.Code, s.DistrictCode)
		}
	}
	// Ordered by code.
	for i := 1; i < len(subs); i++ {
		if subs[i-1].Code > subs[i].Code {
			t.Errorf("SubdistrictsOf not ordered by code at index %d", i)
		}
	}

	// Unknown district code returns nil (not a panic or empty slice mistaken for nil).
	if got := SubdistrictsOf(99999); got != nil {
		t.Errorf("SubdistrictsOf(99999) = %v, want nil", got)
	}

	// District 5001 = Mueang Chiang Mai; its subdistricts span multiple postcodes.
	subs5001 := SubdistrictsOf(5001)
	if len(subs5001) == 0 {
		t.Fatal("SubdistrictsOf(5001) returned nil")
	}
	// First sub should be 500101 (Si Phum).
	if subs5001[0].Code != 500101 {
		t.Errorf("SubdistrictsOf(5001)[0].Code = %d, want 500101", subs5001[0].Code)
	}
}

func TestPostcodesOf(t *testing.T) {
	// District 5001 (Mueang Chiang Mai) spans four postcodes: 50000, 50100, 50200, 50300.
	zips := PostcodesOf(5001)
	want := []int{50000, 50100, 50200, 50300}
	if len(zips) != len(want) {
		t.Fatalf("PostcodesOf(5001) = %v, want %v", zips, want)
	}
	for i, z := range zips {
		if z != want[i] {
			t.Errorf("PostcodesOf(5001)[%d] = %d, want %d", i, z, want[i])
		}
	}
	// Must be sorted ascending.
	for i := 1; i < len(zips); i++ {
		if zips[i-1] > zips[i] {
			t.Errorf("PostcodesOf(5001) not sorted at index %d: %v", i, zips)
		}
	}

	// District 1303 (Pathum Thani province) spans two postcodes: 12110, 12130.
	zips1303 := PostcodesOf(1303)
	if len(zips1303) != 2 {
		t.Fatalf("PostcodesOf(1303) len = %d, want 2 (got %v)", len(zips1303), zips1303)
	}
	if zips1303[0] != 12110 || zips1303[1] != 12130 {
		t.Errorf("PostcodesOf(1303) = %v, want [12110 12130]", zips1303)
	}

	// District 1001 (Phra Nakhon Bangkok) has exactly one postcode (10200).
	zips1001 := PostcodesOf(1001)
	if len(zips1001) != 1 || zips1001[0] != 10200 {
		t.Errorf("PostcodesOf(1001) = %v, want [10200]", zips1001)
	}

	// Unknown district code returns nil.
	if got := PostcodesOf(99999); got != nil {
		t.Errorf("PostcodesOf(99999) = %v, want nil", got)
	}
}

// ---------------------------------------------------------------------------
// model.go: method receivers on District and Subdistrict
// ---------------------------------------------------------------------------

func TestDistrictProvince(t *testing.T) {
	// District 1001 belongs to province 10 (Bangkok).
	d, ok := DistrictByCode(1001)
	if !ok {
		t.Fatal("DistrictByCode(1001) not found")
	}
	p, ok := d.Province()
	if !ok {
		t.Fatal("District.Province() returned ok=false for district 1001")
	}
	if p.Code != 10 {
		t.Errorf("District(1001).Province().Code = %d, want 10", p.Code)
	}
	if p.NameEN != "Bangkok" {
		t.Errorf("District(1001).Province().NameEN = %q, want Bangkok", p.NameEN)
	}

	// District 5001 belongs to province 50 (Chiang Mai).
	d2, ok := DistrictByCode(5001)
	if !ok {
		t.Fatal("DistrictByCode(5001) not found")
	}
	p2, ok := d2.Province()
	if !ok {
		t.Fatal("District.Province() returned ok=false for district 5001")
	}
	if p2.Code != 50 {
		t.Errorf("District(5001).Province().Code = %d, want 50", p2.Code)
	}
}

func TestDistrictSubdistricts(t *testing.T) {
	// District.Subdistricts() is a thin wrapper over SubdistrictsOf.
	d, _ := DistrictByCode(1001)
	subs := d.Subdistricts()
	if len(subs) == 0 {
		t.Fatal("District(1001).Subdistricts() returned empty")
	}
	// Cross-check: same result as package-level SubdistrictsOf.
	direct := SubdistrictsOf(1001)
	if len(subs) != len(direct) {
		t.Errorf("District.Subdistricts() len=%d vs SubdistrictsOf len=%d", len(subs), len(direct))
	}
}

func TestDistrictPostcodes(t *testing.T) {
	// District.Postcodes() is a thin wrapper over PostcodesOf.
	d, _ := DistrictByCode(5001)
	zips := d.Postcodes()
	if len(zips) == 0 {
		t.Fatal("District(5001).Postcodes() returned empty")
	}
	direct := PostcodesOf(5001)
	if len(zips) != len(direct) {
		t.Errorf("District.Postcodes() len=%d vs PostcodesOf len=%d", len(zips), len(direct))
	}
	for i := range zips {
		if zips[i] != direct[i] {
			t.Errorf("District.Postcodes()[%d] = %d, want %d", i, zips[i], direct[i])
		}
	}
}

func TestSubdistrictDistrict(t *testing.T) {
	// Subdistrict 100101 (Phra Borom Maha Ratchawang) belongs to district 1001.
	s, ok := SubdistrictByCode(100101)
	if !ok {
		t.Fatal("SubdistrictByCode(100101) not found")
	}
	d, ok := s.District()
	if !ok {
		t.Fatal("Subdistrict.District() returned ok=false")
	}
	if d.Code != 1001 {
		t.Errorf("Subdistrict(100101).District().Code = %d, want 1001", d.Code)
	}

	// Walk deeper: Subdistrict 500101 (Si Phum) → district 5001 → province 50.
	s2, ok := SubdistrictByCode(500101)
	if !ok {
		t.Fatal("SubdistrictByCode(500101) not found")
	}
	d2, ok := s2.District()
	if !ok {
		t.Fatal("Subdistrict(500101).District() returned ok=false")
	}
	if d2.Code != 5001 {
		t.Errorf("Subdistrict(500101).District().Code = %d, want 5001", d2.Code)
	}
}

func TestSubdistrictProvince(t *testing.T) {
	// Subdistrict.Province() uses DistrictCode/100 for the province code.
	// 100101 → DistrictCode=1001 → 1001/100 = 10 (Bangkok).
	s, _ := SubdistrictByCode(100101)
	p, ok := s.Province()
	if !ok {
		t.Fatal("Subdistrict.Province() returned ok=false for 100101")
	}
	if p.Code != 10 {
		t.Errorf("Subdistrict(100101).Province().Code = %d, want 10", p.Code)
	}

	// 500101 → DistrictCode=5001 → 5001/100 = 50 (Chiang Mai).
	s2, _ := SubdistrictByCode(500101)
	p2, ok := s2.Province()
	if !ok {
		t.Fatal("Subdistrict(500101).Province() returned ok=false")
	}
	if p2.Code != 50 {
		t.Errorf("Subdistrict(500101).Province().Code = %d, want 50", p2.Code)
	}
}

// ---------------------------------------------------------------------------
// region.go: Valid, NameTH, and String unknown-region branch
// ---------------------------------------------------------------------------

func TestRegionValid(t *testing.T) {
	valid := []Region{RegionNorth, RegionCentral, RegionNortheast, RegionWest, RegionEast, RegionSouth}
	for _, r := range valid {
		if !r.Valid() {
			t.Errorf("Region(%d).Valid() = false, want true", r)
		}
	}
	// Unknown region codes must be invalid.
	for _, r := range []Region{0, 7, 99, -1} {
		if Region(r).Valid() {
			t.Errorf("Region(%d).Valid() = true, want false", r)
		}
	}
}

func TestRegionNameTH(t *testing.T) {
	cases := []struct {
		r    Region
		want string
	}{
		{RegionNorth, "ภาคเหนือ"},
		{RegionCentral, "ภาคกลาง"},
		{RegionNortheast, "ภาคตะวันออกเฉียงเหนือ"},
		{RegionWest, "ภาคตะวันตก"},
		{RegionEast, "ภาคตะวันออก"},
		{RegionSouth, "ภาคใต้"},
	}
	for _, tc := range cases {
		got := tc.r.NameTH()
		if got != tc.want {
			t.Errorf("Region(%d).NameTH() = %q, want %q", tc.r, got, tc.want)
		}
	}
	// Unknown region returns "".
	if got := Region(99).NameTH(); got != "" {
		t.Errorf("Region(99).NameTH() = %q, want empty", got)
	}
}

func TestRegionStringUnknown(t *testing.T) {
	// The "Region(unknown)" fallback in String() is not exercised by any existing test.
	got := Region(99).String()
	if got != "Region(unknown)" {
		t.Errorf("Region(99).String() = %q, want %q", got, "Region(unknown)")
	}
	// Known regions return English name.
	if got := RegionNorth.String(); got != "North" {
		t.Errorf("RegionNorth.String() = %q, want North", got)
	}
}

// ---------------------------------------------------------------------------
// normalize.go: stripPrefix — English prefix that matches but leaves empty rest
// ---------------------------------------------------------------------------

func TestStripPrefixEnglishEmptyRest(t *testing.T) {
	// "tambon " is an English prefix (includes trailing space), so only words
	// of the form "tambon word" can match. The branch that returns s unchanged
	// is reached when the english prefix check happens but the text after
	// TrimSpace is "". This is structurally impossible for the current english
	// prefix list (each includes a trailing space so the remaining text is never
	// empty after the prefix length), but the `rest != ""` guard in the Thai
	// branch CAN fire. Let's confirm the Thai-prefix guard works: a Thai prefix
	// that exactly equals the whole string must NOT be stripped.
	cases := []struct {
		in   string
		want string
	}{
		// Thai prefix whose remainder, after TrimSpace, would be empty:
		// strip should be skipped; the string is returned as-is.
		{"เขต", "เขต"},
		{"อำเภอ", "อำเภอ"},
		{"ตำบล", "ตำบล"},
		{"แขวง", "แขวง"},
		{"กิ่งอำเภอ", "กิ่งอำเภอ"},
		{"จังหวัด", "จังหวัด"},
		// English prefix with no trailing content → lowercased, not stripped.
		{"tambon", "tambon"},
		{"amphoe", "amphoe"},
		{"khet", "khet"},
		{"changwat", "changwat"},
		{"khwaeng", "khwaeng"},
	}
	for _, tc := range cases {
		// stripPrefix operates on an already-lowercased, NFC string.
		got := stripPrefix(strings.ToLower(tc.in))
		if got != tc.want {
			t.Errorf("stripPrefix(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// resolve.go: Resolve — postcode-only candidate path filtered by district name
// ---------------------------------------------------------------------------

func TestResolvePostcodeWithDistrictFilter(t *testing.T) {
	// Postcode 50200 belongs to subdistricts in district 1001 (Phra Nakhon, Bangkok)
	// AND Si Phum/Phra Sing (Mueang Chiang Mai district 5001).
	// By specifying district "เขตพระนคร" we select only the Bangkok subdistricts.
	got, err := Resolve(Query{
		Postcode: 10200,
		District: "เขตพระนคร",
	})
	if err != nil {
		t.Fatalf("Resolve(postcode+district) error: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("Resolve(10200, เขตพระนคร) returned no matches")
	}
	for _, m := range got {
		if m.District.Code != 1001 {
			t.Errorf("Resolve filter by district: unexpected district %d", m.District.Code)
		}
		if m.Subdistrict.Postcode != 10200 {
			t.Errorf("Resolve filter by district: unexpected postcode %d", m.Subdistrict.Postcode)
		}
	}
}

func TestResolvePostcodeWithProvinceFilter(t *testing.T) {
	// Postcode 10200 is exclusive to Bangkok subdistricts; filtering by province
	// "กรุงเทพมหานคร" should return all of them (same set as unfiltered).
	all, _ := Resolve(Query{Postcode: 10200})
	filtered, err := Resolve(Query{Postcode: 10200, Province: "กรุงเทพมหานคร"})
	if err != nil {
		t.Fatalf("Resolve(postcode+province) error: %v", err)
	}
	if len(filtered) != len(all) {
		t.Errorf("province filter with matching province changed count %d→%d", len(all), len(filtered))
	}
}

func TestResolvePostcodeWithNonMatchingProvince(t *testing.T) {
	// Postcode 10200 is Bangkok only; specifying province เชียงใหม่ must yield ErrNotFound.
	_, err := Resolve(Query{Postcode: 10200, Province: "เชียงใหม่"})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Resolve(10200, เชียงใหม่) err = %v, want ErrNotFound", err)
	}
}

func TestResolveDistrictEnglishFilter(t *testing.T) {
	// Using an English district name via the postcode-only candidate path.
	got, err := Resolve(Query{
		Postcode: 10200,
		District: "Khet Phra Nakhon",
	})
	if err != nil {
		t.Fatalf("Resolve(postcode, EN district) error: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("Resolve returned no matches for English district name")
	}
	for _, m := range got {
		if m.District.Code != 1001 {
			t.Errorf("got unexpected district %d", m.District.Code)
		}
	}
}

// ---------------------------------------------------------------------------
// data.go: clone behaviour — mutating the returned slice must not affect the store
// ---------------------------------------------------------------------------

func TestCloneIsolation(t *testing.T) {
	// Mutating the slice returned by Provinces must not affect subsequent calls.
	ps := Provinces()
	if len(ps) == 0 {
		t.Fatal("Provinces() returned empty")
	}
	original := ps[0].NameEN
	ps[0].NameEN = "MUTATED"
	ps2 := Provinces()
	if ps2[0].NameEN != original {
		t.Errorf("mutating returned slice affected the store: got %q, want %q", ps2[0].NameEN, original)
	}
}

func TestSubdistrictsCloneIsolation(t *testing.T) {
	subs := SubdistrictsOf(1001)
	if len(subs) == 0 {
		t.Fatal("SubdistrictsOf returned empty")
	}
	subs[0].NameEN = "MUTATED"
	subs2 := SubdistrictsOf(1001)
	if subs2[0].NameEN == "MUTATED" {
		t.Error("mutating SubdistrictsOf result affected the store")
	}
}

// ---------------------------------------------------------------------------
// model.go: Province.Districts() — already covered by smoke/example but verify
// here for completeness and to confirm hierarchy ordering
// ---------------------------------------------------------------------------

func TestProvinceDistrictsOrdering(t *testing.T) {
	p, ok := ProvinceByCode(10)
	if !ok {
		t.Fatal("ProvinceByCode(10) not found")
	}
	ds := p.Districts()
	if len(ds) == 0 {
		t.Fatal("Bangkok has no districts")
	}
	for i := 1; i < len(ds); i++ {
		if ds[i-1].Code > ds[i].Code {
			t.Errorf("Province.Districts() not ordered at index %d", i)
		}
	}
	// Every district's ProvinceCode must be 10.
	for _, d := range ds {
		if d.ProvinceCode != 10 {
			t.Errorf("Province(10).Districts() contains district %d with ProvinceCode %d", d.Code, d.ProvinceCode)
		}
	}
}
