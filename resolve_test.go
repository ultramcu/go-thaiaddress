package thaiaddress

import (
	"errors"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name            string
		prov, dist, sub int
		wantErr         error // nil, ErrNotFound or ErrInconsistent
	}{
		// Happy path: Bangkok / Khet Phra Nakhon / Phra Borom Maha Ratchawang.
		{"valid bangkok", 10, 1001, 100101, nil},
		// Happy path: Chiang Mai / Mueang Chiang Mai / (its first subdistrict).
		{"valid chiang mai", 50, 5001, 500101, nil},

		// Province does not exist.
		{"province not found", 99, 1001, 100101, ErrNotFound},
		// District does not exist.
		{"district not found", 10, 9999, 100101, ErrNotFound},
		// Subdistrict does not exist.
		{"subdistrict not found", 10, 1001, 999999, ErrNotFound},
		// Zero subdistrict is treated as not found (strict: all three required).
		{"zero subdistrict", 10, 1001, 0, ErrNotFound},
		{"zero district", 10, 0, 100101, ErrNotFound},
		{"zero province", 0, 1001, 100101, ErrNotFound},

		// District exists but is in another province.
		{"district wrong province", 50, 1001, 100101, ErrInconsistent},
		// Subdistrict exists but is in another district.
		{"subdistrict wrong district", 10, 1002, 100101, ErrInconsistent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.prov, tt.dist, tt.sub)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate(%d,%d,%d) = %v, want nil", tt.prov, tt.dist, tt.sub, err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate(%d,%d,%d) = %v, want errors.Is %v", tt.prov, tt.dist, tt.sub, err, tt.wantErr)
			}
		})
	}
}

func TestValidateSubdistrict500101Exists(t *testing.T) {
	// Guard: the "valid chiang mai" case above assumes subdistrict 500101 exists
	// and belongs to district 5001. If the dataset changes, fail loudly here.
	s, ok := SubdistrictByCode(500101)
	if !ok || s.DistrictCode != 5001 {
		t.Fatalf("fixture drift: subdistrict 500101 ok=%v district=%d", ok, s.DistrictCode)
	}
}

func TestResolveSubdistrictDistrictProvince(t *testing.T) {
	// "ในเมือง" in อำเภอเมืองขอนแก่น, จังหวัดขอนแก่น resolves to exactly one place:
	// subdistrict 400101.
	got, err := Resolve(Query{
		Subdistrict: "ในเมือง",
		District:    "เมืองขอนแก่น",
		Province:    "ขอนแก่น",
	})
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("Resolve returned %d matches, want 1: %+v", len(got), got)
	}
	m := got[0]
	if m.Subdistrict.Code != 400101 || m.District.Code != 4001 || m.Province.Code != 40 {
		t.Errorf("Resolve = sub %d / dist %d / prov %d, want 400101/4001/40",
			m.Subdistrict.Code, m.District.Code, m.Province.Code)
	}

	// Same query with admin prefixes and English names must resolve identically.
	got2, err := Resolve(Query{
		Subdistrict: "ตำบลในเมือง",
		District:    "Amphoe Mueang Khon Kaen",
		Province:    "Changwat Khon Kaen",
	})
	if err != nil || len(got2) != 1 || got2[0].Subdistrict.Code != 400101 {
		t.Errorf("Resolve with prefixes/EN = %+v err=%v, want single 400101", got2, err)
	}
}

func TestResolvePostcodePlusSubdistrict(t *testing.T) {
	// Postcode 40000 + ในเมือง → only the Khon Kaen one (400101).
	got, err := Resolve(Query{Subdistrict: "ในเมือง", Postcode: 40000})
	if err != nil {
		t.Fatalf("Resolve error: %v", err)
	}
	if len(got) != 1 || got[0].Subdistrict.Code != 400101 {
		t.Fatalf("Resolve(ในเมือง, 40000) = %+v, want single 400101", got)
	}
}

func TestResolvePostcodeOnly(t *testing.T) {
	// Postcode 10200 alone → all 12 subdistricts that use it, ordered by code.
	got, err := Resolve(Query{Postcode: 10200})
	if err != nil {
		t.Fatalf("Resolve(postcode only) error: %v", err)
	}
	want := ByPostcode(10200)
	if len(got) != len(want) {
		t.Fatalf("Resolve(10200) = %d matches, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i].Subdistrict.Code != want[i].Code {
			t.Errorf("Resolve(10200)[%d] sub = %d, want %d", i, got[i].Subdistrict.Code, want[i].Code)
		}
	}
	// Ordered by subdistrict code.
	for i := 1; i < len(got); i++ {
		if got[i-1].Subdistrict.Code > got[i].Subdistrict.Code {
			t.Errorf("Resolve results not ordered by subdistrict code")
		}
	}
}

func TestResolveAmbiguous(t *testing.T) {
	// "ในเมือง" with no other context → many matches (one per province that has it).
	got, err := Resolve(Query{Subdistrict: "ในเมือง"})
	if err != nil {
		t.Fatalf("Resolve(ambiguous) error: %v", err)
	}
	if len(got) != 22 {
		t.Errorf("Resolve(ในเมือง) = %d matches, want 22", len(got))
	}
	// Filtering by province narrows it down.
	narrowed, err := Resolve(Query{Subdistrict: "ในเมือง", Province: "ขอนแก่น"})
	if err != nil {
		t.Fatalf("Resolve(narrowed) error: %v", err)
	}
	if len(narrowed) >= len(got) {
		t.Errorf("province filter did not narrow: %d vs %d", len(narrowed), len(got))
	}
}

func TestResolveNotFound(t *testing.T) {
	// Subdistrict name that does not exist.
	if _, err := Resolve(Query{Subdistrict: "ไม่มีจริง"}); !errors.Is(err, ErrNotFound) {
		t.Errorf("Resolve(nonexistent sub) err = %v, want ErrNotFound", err)
	}
	// Real subdistrict but wrong province → filtered out → not found.
	if _, err := Resolve(Query{Subdistrict: "ในเมือง", Province: "เชียงใหม่"}); !errors.Is(err, ErrNotFound) {
		t.Errorf("Resolve(ในเมือง in เชียงใหม่) err = %v, want ErrNotFound", err)
	}
	// Real subdistrict but mismatched postcode.
	if _, err := Resolve(Query{Subdistrict: "ในเมือง", Postcode: 99999}); !errors.Is(err, ErrNotFound) {
		t.Errorf("Resolve(ในเมือง, 99999) err = %v, want ErrNotFound", err)
	}
	// Postcode that does not exist.
	if _, err := Resolve(Query{Postcode: 99999}); !errors.Is(err, ErrNotFound) {
		t.Errorf("Resolve(postcode 99999) err = %v, want ErrNotFound", err)
	}
}

func TestResolveNoUsableField(t *testing.T) {
	// No subdistrict name and no postcode → error (district/province alone is not
	// enough to drive resolution).
	for _, q := range []Query{
		{},
		{Province: "ขอนแก่น"},
		{District: "เมืองขอนแก่น", Province: "ขอนแก่น"},
		{Subdistrict: "   "},
	} {
		if _, err := Resolve(q); err == nil {
			t.Errorf("Resolve(%+v) = nil error, want error", q)
		}
	}
}
