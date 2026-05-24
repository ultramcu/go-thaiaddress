package thaiaddress

import (
	"errors"
	"fmt"
)

// ErrNotFound is returned (wrapped) when a referenced code, or a free-text
// query, does not match any administrative area.
var ErrNotFound = errors.New("thaiaddress: not found")

// ErrInconsistent is returned (wrapped) when codes exist individually but do not
// form a valid hierarchy (a district not in the named province, or a
// subdistrict not in the named district).
var ErrInconsistent = errors.New("thaiaddress: inconsistent hierarchy")

// Validate reports whether the three codes name a real, internally consistent
// place. It returns nil only when all of the following hold:
//
//   - provinceCode exists,
//   - districtCode exists and districtCode/100 == provinceCode,
//   - subdistrictCode exists and subdistrictCode/100 == districtCode.
//
// All three codes are required; this is deliberately strict. To validate only a
// province+district pair, pass that district's code and any of its subdistrict
// codes, or check the pair yourself via DistrictByCode. A zero (or otherwise
// unknown) subdistrictCode is treated as not found, not as "skip".
//
// Failures wrap ErrNotFound (a code does not exist) or ErrInconsistent (codes
// exist but the hierarchy is wrong); the message names the offending code.
func Validate(provinceCode, districtCode, subdistrictCode int) error {
	if _, ok := ProvinceByCode(provinceCode); !ok {
		return fmt.Errorf("province %d: %w", provinceCode, ErrNotFound)
	}

	d, ok := DistrictByCode(districtCode)
	if !ok {
		return fmt.Errorf("district %d: %w", districtCode, ErrNotFound)
	}
	if d.ProvinceCode != provinceCode {
		return fmt.Errorf("district %d belongs to province %d, not %d: %w",
			districtCode, d.ProvinceCode, provinceCode, ErrInconsistent)
	}

	s, ok := SubdistrictByCode(subdistrictCode)
	if !ok {
		return fmt.Errorf("subdistrict %d: %w", subdistrictCode, ErrNotFound)
	}
	if s.DistrictCode != districtCode {
		return fmt.Errorf("subdistrict %d belongs to district %d, not %d: %w",
			subdistrictCode, s.DistrictCode, districtCode, ErrInconsistent)
	}

	return nil
}

// Query is a free-text address to resolve. Any combination of fields may be
// set; empty strings and a zero Postcode are treated as "unspecified". The name
// fields are matched after normalization (see NormalizeName), so admin prefixes
// such as "ตำบล", "อำเภอ", "เขต" and their English forms are tolerated.
type Query struct {
	Subdistrict string // ตำบล/แขวง name (Thai or English)
	District    string // อำเภอ/เขต name (Thai or English)
	Province    string // จังหวัด name (Thai or English)
	Postcode    int    // 5-digit postal code, or 0 if unknown
}

// Match is a fully-qualified place: a subdistrict together with the district and
// province it belongs to.
type Match struct {
	Province    Province
	District    District
	Subdistrict Subdistrict
}

// Resolve turns a free-text Query into zero or more fully-qualified Matches,
// ordered by subdistrict code.
//
// It starts from the most specific signal available: if Subdistrict is given it
// uses the matching subdistricts as candidates; otherwise, if Postcode is given,
// it uses every subdistrict with that postcode. Candidates are then filtered by
// any of District, Province and Postcode that were supplied (each compared by
// normalized name, or by exact value for the postcode). Every surviving
// candidate is expanded into a Match by walking its codes up to the owning
// district and province.
//
// Errors:
//   - if no usable field is set (no Subdistrict name and no Postcode), Resolve
//     returns an error;
//   - if usable fields are set but nothing matches, it returns ErrNotFound.
//
// District/Province alone (without a Subdistrict name or Postcode) are not
// enough to drive resolution, because Resolve always resolves down to a
// subdistrict; use SearchDistricts/DistrictsOf for province- or district-level
// queries.
func Resolve(q Query) ([]Match, error) {
	hasSub := NormalizeName(q.Subdistrict) != ""
	if !hasSub && q.Postcode == 0 {
		return nil, fmt.Errorf("resolve: need at least a subdistrict name or a postcode: %w", ErrNotFound)
	}

	// Build the candidate set from the most specific available signal.
	var candidates []Subdistrict
	if hasSub {
		candidates = FindSubdistricts(q.Subdistrict)
	} else {
		candidates = ByPostcode(q.Postcode)
	}

	wantDistrict := NormalizeName(q.District)
	wantProvince := NormalizeName(q.Province)

	var out []Match
	for _, s := range candidates {
		if q.Postcode != 0 && s.Postcode != q.Postcode {
			continue
		}

		d, ok := DistrictByCode(s.DistrictCode)
		if !ok {
			continue
		}
		if wantDistrict != "" &&
			NormalizeName(d.NameTH) != wantDistrict &&
			NormalizeName(d.NameEN) != wantDistrict {
			continue
		}

		p, ok := ProvinceByCode(d.ProvinceCode)
		if !ok {
			continue
		}
		if wantProvince != "" &&
			NormalizeName(p.NameTH) != wantProvince &&
			NormalizeName(p.NameEN) != wantProvince {
			continue
		}

		out = append(out, Match{Province: p, District: d, Subdistrict: s})
	}

	if len(out) == 0 {
		return nil, ErrNotFound
	}
	// candidates come from code-ordered indexes, so out is already ordered by
	// subdistrict code.
	return out, nil
}
