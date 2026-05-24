package thaiaddress

import "strings"

// FindProvinces returns every province whose Thai or English name matches name
// exactly after normalization (see NormalizeName), ordered by code. An
// empty/whitespace query returns nil. Multiple results are possible only in
// theory for provinces (names are unique), but the slice form keeps the API
// uniform with the district and subdistrict finders.
func FindProvinces(name string) []Province {
	q := NormalizeName(name)
	if q == "" {
		return nil
	}
	var out []Province
	for _, p := range load().provinces {
		if NormalizeName(p.NameTH) == q || NormalizeName(p.NameEN) == q {
			out = append(out, p)
		}
	}
	return out
}

// FindDistricts returns every district whose Thai or English name matches name
// exactly after normalization, ordered by code. Because NormalizeName strips the
// "เขต"/"Khet " and "อำเภอ"/"Amphoe " prefixes, "เขตพระนคร", "พระนคร" and
// "อำเภอพระนคร" all match the same district. An empty/whitespace query returns
// nil. District names repeat across provinces, so several matches are common.
func FindDistricts(name string) []District {
	q := NormalizeName(name)
	if q == "" {
		return nil
	}
	var out []District
	for _, d := range load().districts {
		if NormalizeName(d.NameTH) == q || NormalizeName(d.NameEN) == q {
			out = append(out, d)
		}
	}
	return out
}

// FindSubdistricts returns every subdistrict whose Thai or English name matches
// name exactly after normalization, ordered by code. An empty/whitespace query
// returns nil. Subdistrict names are highly non-unique (e.g. "ในเมือง" occurs in
// 22 provinces), so this commonly returns many results; disambiguate with
// Resolve using province/district/postcode context.
func FindSubdistricts(name string) []Subdistrict {
	q := NormalizeName(name)
	if q == "" {
		return nil
	}
	var out []Subdistrict
	for _, s := range load().subdistricts {
		if NormalizeName(s.NameTH) == q || NormalizeName(s.NameEN) == q {
			out = append(out, s)
		}
	}
	return out
}

// SearchProvinces returns every province whose normalized Thai or English name
// begins with the normalized prefix, ordered by code. Intended for
// autocomplete. An empty/whitespace prefix returns nil (rather than every
// province).
func SearchProvinces(prefix string) []Province {
	q := NormalizeName(prefix)
	if q == "" {
		return nil
	}
	var out []Province
	for _, p := range load().provinces {
		if strings.HasPrefix(NormalizeName(p.NameTH), q) || strings.HasPrefix(NormalizeName(p.NameEN), q) {
			out = append(out, p)
		}
	}
	return out
}

// SearchDistricts returns every district whose normalized Thai or English name
// begins with the normalized prefix, ordered by code. An empty/whitespace
// prefix returns nil.
func SearchDistricts(prefix string) []District {
	q := NormalizeName(prefix)
	if q == "" {
		return nil
	}
	var out []District
	for _, d := range load().districts {
		if strings.HasPrefix(NormalizeName(d.NameTH), q) || strings.HasPrefix(NormalizeName(d.NameEN), q) {
			out = append(out, d)
		}
	}
	return out
}

// SearchSubdistricts returns every subdistrict whose normalized Thai or English
// name begins with the normalized prefix, ordered by code. An empty/whitespace
// prefix returns nil.
func SearchSubdistricts(prefix string) []Subdistrict {
	q := NormalizeName(prefix)
	if q == "" {
		return nil
	}
	var out []Subdistrict
	for _, s := range load().subdistricts {
		if strings.HasPrefix(NormalizeName(s.NameTH), q) || strings.HasPrefix(NormalizeName(s.NameEN), q) {
			out = append(out, s)
		}
	}
	return out
}
