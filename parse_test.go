package thaiaddress

import (
	"encoding/json"
	"strings"
	"testing"
)

// parse_test.go
//
// BLIND test suite for the public ParseThaiAddress API contract + the new
// lowerCamel JSON tags on Province/District/Subdistrict.
//
// Authored against the PUBLIC CONTRACT + dataset GROUND TRUTH only; parse.go
// (the implementation, written concurrently) was NOT read. Ground-truth codes
// were confirmed against the Go data (data.go / data/areas.json via the public
// ByPostcode / *ByCode lookups) and the Dart oracle test/parse.dart +
// test/parse_test.dart.
//
// Ground-truth dataset facts used:
//
//   Bangkok chain:
//     province 10   = กรุงเทพมหานคร / Bangkok
//     district 1001 = เขตพระนคร     (province 10)   [1001/100 == 10]
//     subdistrict 100101 = พระบรมมหาราชวัง (district 1001, postcode 10200)
//                   [100101/100 == 1001]
//
//   Chiang Mai chain:
//     province 50   = เชียงใหม่ / Chiang Mai
//     district 5001 = เมืองเชียงใหม่ (province 50)
//     subdistrict 500108 = สุเทพ / Suthep (district 5001, postcode 50200)
//
//   Postcode 50200 -> multiple subdistricts, ALL in district 5001 / province 50
//     => district 5001 pinned, subdistrict nil (ambiguous, never guessed).
//   Postcode 10110 -> spans 2 districts, both province 10
//     => province 10 pinned, district nil.
//   Postcode 13240 -> multi-province => postcode set, province nil, !isEmpty.

// ---------------------------------------------------------------------------
// Helpers for nil-safe code assertions.
// ---------------------------------------------------------------------------

func wantProvince(t *testing.T, p *Province, code int) {
	t.Helper()
	if p == nil {
		t.Errorf("Province = nil; want code %d", code)
		return
	}
	if p.Code != code {
		t.Errorf("Province.Code = %d; want %d", p.Code, code)
	}
}

func wantDistrict(t *testing.T, d *District, code int) {
	t.Helper()
	if d == nil {
		t.Errorf("District = nil; want code %d", code)
		return
	}
	if d.Code != code {
		t.Errorf("District.Code = %d; want %d", d.Code, code)
	}
}

func wantSubdistrict(t *testing.T, s *Subdistrict, code int) {
	t.Helper()
	if s == nil {
		t.Errorf("Subdistrict = nil; want code %d", code)
		return
	}
	if s.Code != code {
		t.Errorf("Subdistrict.Code = %d; want %d", s.Code, code)
	}
}

// ---------------------------------------------------------------------------
// 1. Full Bangkok address with markers.
// ---------------------------------------------------------------------------

func TestParseThaiAddress_FullBangkok(t *testing.T) {
	// dataset: 100101 พระบรมมหาราชวัง / 1001 เขตพระนคร / 10 กทม / zip 10200
	const input = "123/45 ถนนหน้าพระลาน แขวงพระบรมมหาราชวัง เขตพระนคร กรุงเทพมหานคร 10200"

	t.Run("complete chain to exact codes", func(t *testing.T) {
		// FAIL-BEFORE: if แขวง/เขต markers are ignored, the province is mis-pinned,
		// or name+postcode are not combined, a code is wrong/nil and this fails.
		r := ParseThaiAddress(input)
		if !r.IsComplete() {
			t.Errorf("IsComplete() = false; want true (got %+v)", r)
		}
		if r.IsEmpty() {
			t.Errorf("IsEmpty() = true; want false")
		}
		wantProvince(t, r.Province, 10)
		wantDistrict(t, r.District, 1001)
		wantSubdistrict(t, r.Subdistrict, 100101)
		if r.Postcode != 10200 {
			t.Errorf("Postcode = %d; want 10200", r.Postcode)
		}
	})

	t.Run("remainder keeps house/road, drops area words + postcode", func(t *testing.T) {
		// FAIL-BEFORE: if area marker words/names or the postcode leak into the
		// remainder, or the house/road text is discarded, this fails.
		r := ParseThaiAddress(input)
		rem := r.Remainder
		if !strings.Contains(rem, "123/45") {
			t.Errorf("Remainder %q missing house number 123/45", rem)
		}
		if !strings.Contains(rem, "หน้าพระลาน") {
			t.Errorf("Remainder %q missing road name หน้าพระลาน", rem)
		}
		for _, drop := range []string{"แขวง", "เขต", "พระบรมมหาราชวัง", "พระนคร", "กรุงเทพมหานคร", "10200"} {
			if strings.Contains(rem, drop) {
				t.Errorf("Remainder %q should not contain %q", rem, drop)
			}
		}
		if rem != strings.TrimSpace(rem) {
			t.Errorf("Remainder %q is not trimmed", rem)
		}
	})
}

// ---------------------------------------------------------------------------
// 2. Upcountry full markers ตำบล/อำเภอ/จังหวัด (Chiang Mai).
// ---------------------------------------------------------------------------

func TestParseThaiAddress_UpcountryFullMarkers(t *testing.T) {
	// dataset: 500108 สุเทพ / 5001 เมืองเชียงใหม่ / 50 เชียงใหม่ / zip 50200
	const input = "99 หมู่ 2 ตำบลสุเทพ อำเภอเมืองเชียงใหม่ จังหวัดเชียงใหม่ 50200"

	t.Run("complete Chiang Mai chain", func(t *testing.T) {
		// FAIL-BEFORE: if ตำบล/อำเภอ/จังหวัด markers are not recognised, or the
		// levels are mis-resolved, this fails.
		r := ParseThaiAddress(input)
		if !r.IsComplete() {
			t.Errorf("IsComplete() = false; want true (got %+v)", r)
		}
		wantProvince(t, r.Province, 50)
		wantDistrict(t, r.District, 5001)
		wantSubdistrict(t, r.Subdistrict, 500108)
		if r.Postcode != 50200 {
			t.Errorf("Postcode = %d; want 50200", r.Postcode)
		}
	})

	t.Run("remainder preserves house/หมู่, drops area words", func(t *testing.T) {
		// FAIL-BEFORE: if หมู่/house number are dropped, or area names/markers/
		// postcode survive, this fails.
		r := ParseThaiAddress(input)
		rem := r.Remainder
		if !strings.Contains(rem, "99") {
			t.Errorf("Remainder %q missing house number 99", rem)
		}
		if !strings.Contains(rem, "หมู่") {
			t.Errorf("Remainder %q missing หมู่", rem)
		}
		for _, drop := range []string{"ตำบล", "อำเภอ", "จังหวัด", "สุเทพ", "เมืองเชียงใหม่", "50200"} {
			if strings.Contains(rem, drop) {
				t.Errorf("Remainder %q should not contain %q", rem, drop)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 3. Abbreviated markers ต./อ./จ. and the "space after dotted marker" case.
// ---------------------------------------------------------------------------

func TestParseThaiAddress_AbbreviatedMarkers(t *testing.T) {
	// Same Chiang Mai chain; abbreviated form must resolve identically.
	t.Run("ต./อ./จ. no space resolves same chain", func(t *testing.T) {
		// FAIL-BEFORE: if only the long markers (ตำบล/อำเภอ/จังหวัด) are matched
		// and not the dotted abbreviations, levels go nil and this fails.
		r := ParseThaiAddress("ต.สุเทพ อ.เมืองเชียงใหม่ จ.เชียงใหม่ 50200")
		if !r.IsComplete() {
			t.Errorf("IsComplete() = false; want true (got %+v)", r)
		}
		wantProvince(t, r.Province, 50)
		wantDistrict(t, r.District, 5001)
		wantSubdistrict(t, r.Subdistrict, 500108)
		if r.Postcode != 50200 {
			t.Errorf("Postcode = %d; want 50200", r.Postcode)
		}
	})

	t.Run("space after dotted marker (ต. สุเทพ ...) resolves same chain", func(t *testing.T) {
		// KEY REGRESSION (Dart fixed it): a space after the dotted abbreviation is
		// common in Thai ("ต. สุเทพ"). FAIL-BEFORE: if the name reader stops at the
		// space right after `ต.`, the name is empty and every level goes nil.
		r := ParseThaiAddress("ต. สุเทพ อ. เมืองเชียงใหม่ จ. เชียงใหม่ 50200")
		if !r.IsComplete() {
			t.Errorf("IsComplete() = false; want true (got %+v)", r)
		}
		wantProvince(t, r.Province, 50)
		wantDistrict(t, r.District, 5001)
		wantSubdistrict(t, r.Subdistrict, 500108)
		if r.Postcode != 50200 {
			t.Errorf("Postcode = %d; want 50200", r.Postcode)
		}
	})

	t.Run("Bangkok aliases map to province 10", func(t *testing.T) {
		// FAIL-BEFORE: if กทม./กรุงเทพฯ/กรุงเทพมหานคร aliases are not mapped to
		// province 10, the province is nil and this fails.
		for _, alias := range []string{"กทม.", "กรุงเทพฯ", "กรุงเทพมหานคร"} {
			r := ParseThaiAddress("แขวงพระบรมมหาราชวัง เขตพระนคร " + alias)
			wantProvince(t, r.Province, 10)
			wantDistrict(t, r.District, 1001)
		}
	})

	t.Run("dangling dotted marker never panics", func(t *testing.T) {
		// FAIL-BEFORE: skipping spaces after a dotted marker must not run past
		// end-of-string when nothing follows; otherwise this panics.
		for _, s := range []string{"ต.สุเทพ", "อ. ", "จ.", "ต. "} {
			func() {
				defer func() {
					if p := recover(); p != nil {
						t.Errorf("ParseThaiAddress(%q) panicked: %v", s, p)
					}
				}()
				_ = ParseThaiAddress(s)
			}()
		}
	})
}

// ---------------------------------------------------------------------------
// Names-only (no postcode) — pinning must not depend on a postcode token.
// ---------------------------------------------------------------------------

func TestParseThaiAddress_NamesOnlyNoPostcode(t *testing.T) {
	// FAIL-BEFORE: if the parser relies on a postcode to pin levels, dropping the
	// postcode would leave province/district/subdistrict nil.
	r := ParseThaiAddress("แขวงพระบรมมหาราชวัง เขตพระนคร กรุงเทพมหานคร")
	if !r.IsComplete() {
		t.Errorf("IsComplete() = false; want true (got %+v)", r)
	}
	wantProvince(t, r.Province, 10)
	wantDistrict(t, r.District, 1001)
	wantSubdistrict(t, r.Subdistrict, 100101)
	if r.Postcode != 0 {
		t.Errorf("Postcode = %d; want 0 (none)", r.Postcode)
	}
}

// ---------------------------------------------------------------------------
// 4. Postcode-only 50200 -> district pinned, subdistrict nil.
// ---------------------------------------------------------------------------

func TestParseThaiAddress_PostcodeOnly_50200(t *testing.T) {
	// dataset: zip 50200 -> several subdistricts, ALL in district 5001 / prov 50.
	// FAIL-BEFORE: if the parser guesses one of the subdistricts, or fails to pin
	// the shared district/province, this fails.
	r := ParseThaiAddress("50200")
	if r.Postcode != 50200 {
		t.Errorf("Postcode = %d; want 50200", r.Postcode)
	}
	wantProvince(t, r.Province, 50)
	wantDistrict(t, r.District, 5001)
	if r.Subdistrict != nil {
		t.Errorf("Subdistrict = %+v; want nil (ambiguous, must not guess)", r.Subdistrict)
	}
	if r.IsComplete() {
		t.Errorf("IsComplete() = true; want false")
	}
	if r.IsEmpty() {
		t.Errorf("IsEmpty() = true; want false")
	}
}

// ---------------------------------------------------------------------------
// 5. Postcode 10110 -> province pinned, district nil.
// ---------------------------------------------------------------------------

func TestParseThaiAddress_PostcodeOnly_10110(t *testing.T) {
	// dataset: zip 10110 -> spans 2 districts, both province 10. Province is the
	// only shared level.
	// FAIL-BEFORE: if the parser picks one of the districts (guessing) or fails
	// to pin the shared province, this fails.
	r := ParseThaiAddress("10110")
	if r.Postcode != 10110 {
		t.Errorf("Postcode = %d; want 10110", r.Postcode)
	}
	wantProvince(t, r.Province, 10)
	if r.District != nil {
		t.Errorf("District = %+v; want nil (spans 2 districts)", r.District)
	}
	if r.Subdistrict != nil {
		t.Errorf("Subdistrict = %+v; want nil", r.Subdistrict)
	}
	if r.IsEmpty() {
		t.Errorf("IsEmpty() = true; want false")
	}
}

// ---------------------------------------------------------------------------
// 6. Multi-province postcode 13240 -> postcode set, province nil, !empty.
// ---------------------------------------------------------------------------

func TestParseThaiAddress_PostcodeOnly_MultiProvince_13240(t *testing.T) {
	// dataset: zip 13240 -> subdistricts in more than one province => no shared
	// province. FAIL-BEFORE: if the parser pins a province anyway (guessing
	// across provinces), or reports IsEmpty for a recognised postcode, this fails.
	r := ParseThaiAddress("13240")
	if r.Postcode != 13240 {
		t.Errorf("Postcode = %d; want 13240", r.Postcode)
	}
	if r.Province != nil {
		t.Errorf("Province = %+v; want nil (multi-province postcode)", r.Province)
	}
	if r.District != nil {
		t.Errorf("District = %+v; want nil", r.District)
	}
	if r.Subdistrict != nil {
		t.Errorf("Subdistrict = %+v; want nil", r.Subdistrict)
	}
	if r.IsEmpty() {
		t.Errorf("IsEmpty() = true; want false (postcode is set)")
	}
}

// ---------------------------------------------------------------------------
// 7. Thai-digit postcodes ๐–๙.
// ---------------------------------------------------------------------------

func TestParseThaiAddress_ThaiDigitPostcode(t *testing.T) {
	t.Run("๑๐๒๐๐ parses identically to 10200", func(t *testing.T) {
		// FAIL-BEFORE: if Thai numerals are not normalised to Arabic digits, the
		// postcode is unrecognised and the whole parse degrades.
		arabic := ParseThaiAddress("แขวงพระบรมมหาราชวัง เขตพระนคร กรุงเทพมหานคร 10200")
		thai := ParseThaiAddress("แขวงพระบรมมหาราชวัง เขตพระนคร กรุงเทพมหานคร ๑๐๒๐๐")
		if thai.Postcode != 10200 {
			t.Errorf("thai Postcode = %d; want 10200", thai.Postcode)
		}
		if thai.Postcode != arabic.Postcode {
			t.Errorf("thai Postcode %d != arabic Postcode %d", thai.Postcode, arabic.Postcode)
		}
		if thai.IsComplete() != arabic.IsComplete() {
			t.Errorf("IsComplete mismatch: thai=%v arabic=%v", thai.IsComplete(), arabic.IsComplete())
		}
		wantProvince(t, thai.Province, 10)
		wantDistrict(t, thai.District, 1001)
		wantSubdistrict(t, thai.Subdistrict, 100101)
	})

	t.Run("bare ๕๐๒๐๐ resolves like 50200", func(t *testing.T) {
		// FAIL-BEFORE: bare Thai-digit postcode must normalise and pin province 50.
		r := ParseThaiAddress("๕๐๒๐๐")
		if r.Postcode != 50200 {
			t.Errorf("Postcode = %d; want 50200", r.Postcode)
		}
		wantProvince(t, r.Province, 50)
		wantDistrict(t, r.District, 5001)
	})
}

// ---------------------------------------------------------------------------
// 8. Garbage / empty -> IsEmpty, never panics.
// ---------------------------------------------------------------------------

func TestParseThaiAddress_GarbageAndEmpty(t *testing.T) {
	t.Run("never panics on assorted junk", func(t *testing.T) {
		// FAIL-BEFORE: if the parser throws on garbage/empty/digit-only/marker-only
		// input, the recover fires and this fails.
		inputs := []string{
			"", "   ", "hello world", "this is not a thai address",
			"1234567890", "อำเภอ ตำบล จังหวัด", "99999", "ก", "\n\t  \n",
			// Invalid UTF-8, incl. a bad byte right after a dotted marker, must
			// not over-advance the rune scanner past the string end.
			"ต.\xff", "ต.ก\xff", "อ. \xfe\xff", "\xff\xfe\xfd",
		}
		for _, s := range inputs {
			func() {
				defer func() {
					if p := recover(); p != nil {
						t.Errorf("ParseThaiAddress(%q) panicked: %v", s, p)
					}
				}()
				_ = ParseThaiAddress(s)
			}()
		}
	})

	t.Run("empty string => IsEmpty, empty remainder, all nil/zero", func(t *testing.T) {
		// FAIL-BEFORE: if IsEmpty wrongly treats "" as non-empty or returns a
		// non-empty remainder for "", this fails.
		r := ParseThaiAddress("")
		if !r.IsEmpty() {
			t.Errorf("IsEmpty() = false; want true")
		}
		if r.IsComplete() {
			t.Errorf("IsComplete() = true; want false")
		}
		if r.Province != nil || r.District != nil || r.Subdistrict != nil {
			t.Errorf("levels should all be nil; got %+v", r)
		}
		if r.Postcode != 0 {
			t.Errorf("Postcode = %d; want 0", r.Postcode)
		}
		if r.Remainder != "" {
			t.Errorf("Remainder = %q; want empty", r.Remainder)
		}
	})

	t.Run("pure garbage => IsEmpty, garbage survives in remainder", func(t *testing.T) {
		// FAIL-BEFORE: if the parser hallucinates an area from random words, or
		// drops the unmatched words from the remainder, this fails.
		r := ParseThaiAddress("hello world")
		if !r.IsEmpty() {
			t.Errorf("IsEmpty() = false; want true")
		}
		if r.Province != nil || r.District != nil || r.Subdistrict != nil {
			t.Errorf("levels should all be nil; got %+v", r)
		}
		if r.Postcode != 0 {
			t.Errorf("Postcode = %d; want 0", r.Postcode)
		}
		low := strings.ToLower(r.Remainder)
		if !strings.Contains(low, "hello") || !strings.Contains(low, "world") {
			t.Errorf("Remainder %q should preserve the garbage words", r.Remainder)
		}
	})

	t.Run("bare repeated subdistrict name (no marker) => not guessed, IsEmpty", func(t *testing.T) {
		// dataset: subdistrict name หนองบัว appears in many provinces. A bare name
		// with no marker is not scanned; with no context it stays undecidable.
		// FAIL-BEFORE: if the parser picks an arbitrary หนองบัว match instead of
		// leaving it nil, this fails. Must never throw.
		r := ParseThaiAddress("หนองบัว")
		if r.Subdistrict != nil {
			t.Errorf("Subdistrict = %+v; want nil (ambiguous bare name)", r.Subdistrict)
		}
		if r.Province != nil || r.District != nil {
			t.Errorf("Province/District should be nil; got %+v", r)
		}
		if !r.IsEmpty() {
			t.Errorf("IsEmpty() = false; want true")
		}
	})
}

// ---------------------------------------------------------------------------
// 9. Consistency / internal pinning agreement.
// ---------------------------------------------------------------------------

func TestParseThaiAddress_ConsistentChain(t *testing.T) {
	// FAIL-BEFORE: if postcode and names are resolved independently and not
	// cross-checked, a consistent input could still drop a level or the resolved
	// subdistrict's own postcode could disagree with the parsed postcode.
	r := ParseThaiAddress("ต.สุเทพ อ.เมืองเชียงใหม่ จ.เชียงใหม่ 50200")
	if !r.IsComplete() {
		t.Errorf("IsComplete() = false; want true (got %+v)", r)
	}
	wantProvince(t, r.Province, 50)
	wantDistrict(t, r.District, 5001)
	wantSubdistrict(t, r.Subdistrict, 500108)
	// Resolved subdistrict's own postcode must agree with the parsed postcode.
	if r.Subdistrict != nil && r.Subdistrict.Postcode != r.Postcode {
		t.Errorf("Subdistrict.Postcode %d != parsed Postcode %d", r.Subdistrict.Postcode, r.Postcode)
	}
	// Chain must be internally consistent: sub -> district -> province.
	if r.Subdistrict != nil && r.District != nil && r.Subdistrict.DistrictCode != r.District.Code {
		t.Errorf("Subdistrict.DistrictCode %d != District.Code %d", r.Subdistrict.DistrictCode, r.District.Code)
	}
	if r.District != nil && r.Province != nil && r.District.ProvinceCode != r.Province.Code {
		t.Errorf("District.ProvinceCode %d != Province.Code %d", r.District.ProvinceCode, r.Province.Code)
	}
}

// ---------------------------------------------------------------------------
// JSON tags — Dart-matching lowerCamel wire shape.
// ---------------------------------------------------------------------------

func TestProvinceJSONShape(t *testing.T) {
	// Use the real dataset entry for province 10 (Bangkok).
	p, ok := ProvinceByCode(10)
	if !ok {
		t.Fatal("ProvinceByCode(10) not found in dataset")
	}
	// FAIL-BEFORE: without the lowerCamel JSON tags, Marshal emits Go field names
	// ("Code","NameTH","NameEN","Region"); these contains-checks fail.
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal(Province) error: %v", err)
	}
	js := string(b)
	for _, want := range []string{`"code":10`, `"nameTh":`, `"nameEn":`, `"region":`} {
		if !strings.Contains(js, want) {
			t.Errorf("Province JSON %s missing %q", js, want)
		}
	}

	t.Run("round-trip equality", func(t *testing.T) {
		// FAIL-BEFORE: if the tags on Marshal and Unmarshal disagree (or a field is
		// dropped), the round-tripped struct won't == the original.
		var back Province
		if err := json.Unmarshal(b, &back); err != nil {
			t.Fatalf("Unmarshal(Province) error: %v", err)
		}
		if back != p {
			t.Errorf("round-trip mismatch: got %+v; want %+v", back, p)
		}
	})
}

func TestDistrictJSONShape(t *testing.T) {
	// Real dataset entry for district 1001 (เขตพระนคร).
	d, ok := DistrictByCode(1001)
	if !ok {
		t.Fatal("DistrictByCode(1001) not found in dataset")
	}
	// FAIL-BEFORE: without tags, the owning-province key is "ProvinceCode", not
	// the Dart-matching "provinceCode".
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("Marshal(District) error: %v", err)
	}
	js := string(b)
	for _, want := range []string{`"code":1001`, `"provinceCode":`, `"nameTh":`, `"nameEn":`} {
		if !strings.Contains(js, want) {
			t.Errorf("District JSON %s missing %q", js, want)
		}
	}

	t.Run("round-trip equality", func(t *testing.T) {
		var back District
		if err := json.Unmarshal(b, &back); err != nil {
			t.Fatalf("Unmarshal(District) error: %v", err)
		}
		if back != d {
			t.Errorf("round-trip mismatch: got %+v; want %+v", back, d)
		}
	})
}

func TestSubdistrictJSONShape(t *testing.T) {
	// Real dataset entry for subdistrict 100101 (พระบรมมหาราชวัง, postcode 10200).
	s, ok := SubdistrictByCode(100101)
	if !ok {
		t.Fatal("SubdistrictByCode(100101) not found in dataset")
	}
	// FAIL-BEFORE: without tags, keys are "DistrictCode"/"Postcode", not the
	// Dart-matching "districtCode"/"postcode".
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Marshal(Subdistrict) error: %v", err)
	}
	js := string(b)
	for _, want := range []string{`"code":100101`, `"districtCode":`, `"postcode":`, `"nameTh":`, `"nameEn":`} {
		if !strings.Contains(js, want) {
			t.Errorf("Subdistrict JSON %s missing %q", js, want)
		}
	}

	t.Run("round-trip equality", func(t *testing.T) {
		var back Subdistrict
		if err := json.Unmarshal(b, &back); err != nil {
			t.Fatalf("Unmarshal(Subdistrict) error: %v", err)
		}
		if back != s {
			t.Errorf("round-trip mismatch: got %+v; want %+v", back, s)
		}
	})
}
