package thaiaddress

import (
	"sort"
	"testing"
)

func provinceCodes(ps []Province) []int {
	out := make([]int, len(ps))
	for i, p := range ps {
		out[i] = p.Code
	}
	return out
}

func districtCodes(ds []District) []int {
	out := make([]int, len(ds))
	for i, d := range ds {
		out[i] = d.Code
	}
	return out
}

func subdistrictCodes(ss []Subdistrict) []int {
	out := make([]int, len(ss))
	for i, s := range ss {
		out[i] = s.Code
	}
	return out
}

func isSortedInts(xs []int) bool {
	return sort.IntsAreSorted(xs)
}

func TestFindProvinces(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantCodes []int
	}{
		{"TH exact", "เชียงใหม่", []int{50}},
		{"EN exact", "Chiang Mai", []int{50}},
		{"EN lowercase", "chiang mai", []int{50}},
		{"EN with changwat prefix", "Changwat Chiang Mai", []int{50}},
		{"TH with จังหวัด prefix", "จังหวัดเชียงใหม่", []int{50}},
		{"Bangkok EN", "Bangkok", []int{10}},
		{"Bangkok TH", "กรุงเทพมหานคร", []int{10}},
		{"empty", "", nil},
		{"whitespace", "   ", nil},
		{"no match", "Atlantis", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := provinceCodes(FindProvinces(tt.query))
			if !equalInts(got, tt.wantCodes) {
				t.Errorf("FindProvinces(%q) codes = %v, want %v", tt.query, got, tt.wantCodes)
			}
		})
	}
}

func TestFindDistricts(t *testing.T) {
	// "เขตพระนคร" (Bangkok) must be findable both with and without the เขต prefix,
	// and via its English name.
	for _, q := range []string{"เขตพระนคร", "พระนคร", "Khet Phra Nakhon", "Phra Nakhon", "phra nakhon"} {
		got := FindDistricts(q)
		found := false
		for _, d := range got {
			if d.Code == 1001 {
				found = true
			}
		}
		if !found {
			t.Errorf("FindDistricts(%q) did not include district 1001 (got %v)", q, districtCodes(got))
		}
	}

	// Upcountry "เมืองเชียงใหม่" has no prefix; the อำเภอ-prefixed form must match it.
	for _, q := range []string{"เมืองเชียงใหม่", "อำเภอเมืองเชียงใหม่", "Mueang Chiang Mai"} {
		got := FindDistricts(q)
		if len(got) != 1 || got[0].Code != 5001 {
			t.Errorf("FindDistricts(%q) = %v, want exactly [5001]", q, districtCodes(got))
		}
	}

	// Duplicate district name across provinces: "เฉลิมพระเกียรติ" exists in several.
	dup := FindDistricts("เฉลิมพระเกียรติ")
	if len(dup) < 2 {
		t.Errorf("FindDistricts(เฉลิมพระเกียรติ) = %v, want multiple", districtCodes(dup))
	}
	if !isSortedInts(districtCodes(dup)) {
		t.Errorf("FindDistricts results not ordered by code: %v", districtCodes(dup))
	}

	if got := FindDistricts(""); got != nil {
		t.Errorf("FindDistricts(empty) = %v, want nil", got)
	}
	if got := FindDistricts("   "); got != nil {
		t.Errorf("FindDistricts(whitespace) = %v, want nil", got)
	}
}

func TestFindSubdistricts(t *testing.T) {
	// "ในเมือง" is the canonical duplicate: it occurs in 22 provinces.
	dup := FindSubdistricts("ในเมือง")
	if len(dup) != 22 {
		t.Errorf("FindSubdistricts(ในเมือง) returned %d, want 22", len(dup))
	}
	if !isSortedInts(subdistrictCodes(dup)) {
		t.Errorf("FindSubdistricts results not ordered by code: %v", subdistrictCodes(dup))
	}
	// The ตำบล prefix must not change the result.
	if pref := FindSubdistricts("ตำบลในเมือง"); len(pref) != len(dup) {
		t.Errorf("FindSubdistricts(ตำบลในเมือง) = %d, want %d", len(pref), len(dup))
	}

	// A specific Bangkok แขวง by EN name (no prefix stored).
	got := FindSubdistricts("Phra Borom Maha Ratchawang")
	found := false
	for _, s := range got {
		if s.Code == 100101 {
			found = true
		}
	}
	if !found {
		t.Errorf("FindSubdistricts(Phra Borom Maha Ratchawang) missing 100101, got %v", subdistrictCodes(got))
	}

	if got := FindSubdistricts(""); got != nil {
		t.Errorf("FindSubdistricts(empty) = %v, want nil", got)
	}
}

func TestSearchProvinces(t *testing.T) {
	// EN prefix.
	got := SearchProvinces("Chiang")
	codes := provinceCodes(got)
	if !containsInt(codes, 50) || !containsInt(codes, 57) { // Chiang Mai + Chiang Rai
		t.Errorf("SearchProvinces(Chiang) = %v, want to include 50 and 57", codes)
	}
	if !isSortedInts(codes) {
		t.Errorf("SearchProvinces not ordered: %v", codes)
	}

	// TH prefix.
	th := SearchProvinces("เชียง")
	if !containsInt(provinceCodes(th), 50) {
		t.Errorf("SearchProvinces(เชียง) = %v, want to include 50", provinceCodes(th))
	}

	// Empty/whitespace must NOT return everything.
	if got := SearchProvinces(""); got != nil {
		t.Errorf("SearchProvinces(empty) = %v, want nil", got)
	}
	if got := SearchProvinces("   "); got != nil {
		t.Errorf("SearchProvinces(whitespace) = %v, want nil", got)
	}
}

func TestSearchDistricts(t *testing.T) {
	// "Mueang " is a very common district prefix word; searching it should yield
	// many districts, all ordered by code.
	got := SearchDistricts("Mueang Chiang")
	codes := districtCodes(got)
	if !containsInt(codes, 5001) { // Mueang Chiang Mai
		t.Errorf("SearchDistricts(Mueang Chiang) = %v, want to include 5001", codes)
	}
	if !isSortedInts(codes) {
		t.Errorf("SearchDistricts not ordered: %v", codes)
	}

	if got := SearchDistricts(""); got != nil {
		t.Errorf("SearchDistricts(empty) = %v, want nil", got)
	}
}

func TestSearchSubdistricts(t *testing.T) {
	got := SearchSubdistricts("ในเมือง")
	if len(got) < 22 { // at least all the exact "ในเมือง" plus any longer names
		t.Errorf("SearchSubdistricts(ในเมือง) returned %d, want >= 22", len(got))
	}
	if !isSortedInts(subdistrictCodes(got)) {
		t.Errorf("SearchSubdistricts not ordered: %v", subdistrictCodes(got))
	}

	if got := SearchSubdistricts(""); got != nil {
		t.Errorf("SearchSubdistricts(empty) = %v, want nil", got)
	}
}

// --- small slice helpers ---

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func containsInt(xs []int, want int) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}
