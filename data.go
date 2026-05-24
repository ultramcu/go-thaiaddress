package thaiaddress

import (
	_ "embed"
	"encoding/json"
	"sort"
	"sync"
)

//go:embed data/areas.json
var rawAreas []byte

// store holds the parsed dataset and its lookup indexes. It is built once, on
// first use, by load.
type store struct {
	provinces    []Province
	districts    []District
	subdistricts []Subdistrict

	provinceByCode    map[int]Province
	districtByCode    map[int]District
	subdistrictByCode map[int]Subdistrict

	districtsByProvince    map[int][]District
	subdistrictsByDistrict map[int][]Subdistrict
	byPostcode             map[int][]Subdistrict
}

var (
	loadOnce sync.Once
	data     *store
)

// jsonAreas mirrors data/areas.json (compact keys to keep the asset small).
type jsonAreas struct {
	Provinces []struct {
		Code   int    `json:"code"`
		TH     string `json:"th"`
		EN     string `json:"en"`
		Region int    `json:"region"`
	} `json:"provinces"`
	Districts []struct {
		Code     int    `json:"code"`
		Province int    `json:"province"`
		TH       string `json:"th"`
		EN       string `json:"en"`
	} `json:"districts"`
	Subdistricts []struct {
		Code     int    `json:"code"`
		District int    `json:"district"`
		TH       string `json:"th"`
		EN       string `json:"en"`
		Zip      int    `json:"zip"`
	} `json:"subdistricts"`
}

func load() *store {
	loadOnce.Do(func() {
		var j jsonAreas
		if err := json.Unmarshal(rawAreas, &j); err != nil {
			// The asset is embedded and validated in tests; a failure here means
			// a corrupt build, which is unrecoverable.
			panic("thaiaddress: cannot parse embedded dataset: " + err.Error())
		}

		s := &store{
			provinceByCode:         make(map[int]Province, len(j.Provinces)),
			districtByCode:         make(map[int]District, len(j.Districts)),
			subdistrictByCode:      make(map[int]Subdistrict, len(j.Subdistricts)),
			districtsByProvince:    make(map[int][]District),
			subdistrictsByDistrict: make(map[int][]Subdistrict),
			byPostcode:             make(map[int][]Subdistrict),
		}

		s.provinces = make([]Province, len(j.Provinces))
		for i, p := range j.Provinces {
			pv := Province{Code: p.Code, NameTH: p.TH, NameEN: p.EN, Region: Region(p.Region)}
			s.provinces[i] = pv
			s.provinceByCode[pv.Code] = pv
		}

		s.districts = make([]District, len(j.Districts))
		for i, d := range j.Districts {
			dt := District{Code: d.Code, ProvinceCode: d.Province, NameTH: d.TH, NameEN: d.EN}
			s.districts[i] = dt
			s.districtByCode[dt.Code] = dt
			s.districtsByProvince[dt.ProvinceCode] = append(s.districtsByProvince[dt.ProvinceCode], dt)
		}

		s.subdistricts = make([]Subdistrict, len(j.Subdistricts))
		for i, sd := range j.Subdistricts {
			st := Subdistrict{Code: sd.Code, DistrictCode: sd.District, NameTH: sd.TH, NameEN: sd.EN, Postcode: sd.Zip}
			s.subdistricts[i] = st
			s.subdistrictByCode[st.Code] = st
			s.subdistrictsByDistrict[st.DistrictCode] = append(s.subdistrictsByDistrict[st.DistrictCode], st)
			s.byPostcode[st.Postcode] = append(s.byPostcode[st.Postcode], st)
		}

		// Keep every slice ordered by code for stable, deterministic output.
		sort.Slice(s.provinces, func(i, k int) bool { return s.provinces[i].Code < s.provinces[k].Code })
		sort.Slice(s.districts, func(i, k int) bool { return s.districts[i].Code < s.districts[k].Code })
		sort.Slice(s.subdistricts, func(i, k int) bool { return s.subdistricts[i].Code < s.subdistricts[k].Code })
		for k := range s.districtsByProvince {
			v := s.districtsByProvince[k]
			sort.Slice(v, func(i, k int) bool { return v[i].Code < v[k].Code })
		}
		for k := range s.subdistrictsByDistrict {
			v := s.subdistrictsByDistrict[k]
			sort.Slice(v, func(i, k int) bool { return v[i].Code < v[k].Code })
		}
		for k := range s.byPostcode {
			v := s.byPostcode[k]
			sort.Slice(v, func(i, k int) bool { return v[i].Code < v[k].Code })
		}

		data = s
	})
	return data
}

// Provinces returns all 77 provinces, ordered by code.
func Provinces() []Province { return clone(load().provinces) }

// Districts returns every district, ordered by code.
func Districts() []District { return clone(load().districts) }

// Subdistricts returns every subdistrict, ordered by code.
func Subdistricts() []Subdistrict { return clone(load().subdistricts) }

// ProvinceByCode looks up a province by its 2-digit code.
func ProvinceByCode(code int) (Province, bool) { p, ok := load().provinceByCode[code]; return p, ok }

// DistrictByCode looks up a district by its 4-digit code.
func DistrictByCode(code int) (District, bool) { d, ok := load().districtByCode[code]; return d, ok }

// SubdistrictByCode looks up a subdistrict by its 6-digit code.
func SubdistrictByCode(code int) (Subdistrict, bool) {
	s, ok := load().subdistrictByCode[code]
	return s, ok
}

// DistrictsOf returns the districts in a province (by 2-digit code), ordered by
// code; nil if the province is unknown.
func DistrictsOf(provinceCode int) []District { return clone(load().districtsByProvince[provinceCode]) }

// SubdistrictsOf returns the subdistricts in a district (by 4-digit code),
// ordered by code; nil if the district is unknown.
func SubdistrictsOf(districtCode int) []Subdistrict {
	return clone(load().subdistrictsByDistrict[districtCode])
}

// ByPostcode returns every subdistrict that uses the given 5-digit postal code,
// ordered by code. A postal code commonly maps to many subdistricts.
func ByPostcode(postcode int) []Subdistrict { return clone(load().byPostcode[postcode]) }

// PostcodesOf returns the distinct postal codes used within a district (by
// 4-digit code), ordered ascending; nil if the district is unknown.
func PostcodesOf(districtCode int) []int {
	subs := load().subdistrictsByDistrict[districtCode]
	if len(subs) == 0 {
		return nil
	}
	seen := make(map[int]struct{}, len(subs))
	var zips []int
	for _, s := range subs {
		if _, ok := seen[s.Postcode]; !ok {
			seen[s.Postcode] = struct{}{}
			zips = append(zips, s.Postcode)
		}
	}
	sort.Ints(zips)
	return zips
}

// clone returns a copy of s so callers cannot mutate the package's indexes.
func clone[T any](s []T) []T {
	if s == nil {
		return nil
	}
	out := make([]T, len(s))
	copy(out, s)
	return out
}
