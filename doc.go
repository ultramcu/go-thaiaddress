// Package thaiaddress provides Thailand's administrative-area data — every
// province (จังหวัด), district (อำเภอ/เขต) and subdistrict (ตำบล/แขวง) together
// with its postal code — plus lookup, hierarchy navigation, name search,
// autocomplete and validation, all served from data embedded in the binary
// (no network, no files at runtime).
//
// Codes are the official Department of Provincial Administration (DOPA)
// geocodes: a 2-digit province code (10–96), a 4-digit district code whose
// first two digits are the province, and a 6-digit subdistrict code whose
// first four digits are the district. This makes the hierarchy derivable from
// any code and lets callers store a single integer per level.
//
//	p, _ := thaiaddress.ProvinceByCode(10)        // Bangkok
//	for _, d := range p.Districts() { ... }
//	subs := thaiaddress.ByPostcode(10200)         // subdistricts using a zip
//
// Thai subdistrict names are not unique — 795 names repeat across provinces
// (e.g. "ในเมือง") — so name lookups return slices; resolve to a single place
// with the province/district context or a postal code.
//
// The embedded dataset is a snapshot of github.com/kongvut/thai-province-data
// (MIT); see data/SOURCE.md.
package thaiaddress
