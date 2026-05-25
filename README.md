# go-thaiaddress

[![Go Reference](https://pkg.go.dev/badge/github.com/ultramcu/go-thaiaddress.svg)](https://pkg.go.dev/github.com/ultramcu/go-thaiaddress)
[![Release](https://img.shields.io/github/v/release/ultramcu/go-thaiaddress?sort=semver)](https://github.com/ultramcu/go-thaiaddress/releases)
[![CI](https://github.com/ultramcu/go-thaiaddress/actions/workflows/ci.yml/badge.svg)](https://github.com/ultramcu/go-thaiaddress/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/ultramcu/go-thaiaddress/graph/badge.svg)](https://codecov.io/gh/ultramcu/go-thaiaddress)
[![Go Report Card](https://goreportcard.com/badge/github.com/ultramcu/go-thaiaddress)](https://goreportcard.com/report/github.com/ultramcu/go-thaiaddress)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

ข้อมูลเขตการปกครองของประเทศไทย (จังหวัด/อำเภอ/ตำบล พร้อมรหัสไปรษณีย์) สำหรับภาษา Go ฝังข้อมูลมาในตัว ค้นหา/เติมคำอัตโนมัติ/ตรวจสอบที่อยู่ได้ทันที ไม่ต้องต่อเน็ต

A pure-Go, MIT-licensed library of Thailand's administrative areas — 77 provinces
(จังหวัด), 928 districts (อำเภอ/เขต), and 7,452 subdistricts (ตำบล/แขวง) with
postal codes — fully embedded, with lookup, hierarchy navigation, autocomplete,
validation, and address resolution. No network or files at runtime.

## Why this exists

There is no maintained, properly licensed, importable Go package that bundles
the full Thai province/district/subdistrict dataset *together with* lookup,
autocomplete, and address-resolution helpers. `go-thaiaddress` does exactly that:
one `go get`, no data files to ship, no runtime fetches.

## Install

```sh
go get github.com/ultramcu/go-thaiaddress
```

Requires Go 1.21+.

## Features

- **Embedded dataset** via `go:embed` — zero network/filesystem access at runtime.
- **Complete coverage** — 77 provinces, 928 districts, 7,452 subdistricts, postcodes.
- **Official DOPA geocodes** — 2-digit provinces (10–96), 4-digit districts,
  6-digit subdistricts; the hierarchy is derivable (`districtCode/100 == provinceCode`,
  `subCode/100 == districtCode`).
- **Lookups by code** and **hierarchy navigation** (parent/child in both directions).
- **Postal-code lookups** — find subdistricts by postcode, or postcodes of a district.
- **Name search & autocomplete** — exact (normalized) find plus prefix search.
- **Validation** of a province/district/subdistrict triple.
- **Address resolution** — match free-form fragments (plus an optional postcode)
  to concrete `Province`/`District`/`Subdistrict` matches.
- **Thai + English names** for every area, plus six regions.

## Quick start

### Look up by code and read names

```go
package main

import (
	"fmt"

	thai "github.com/ultramcu/go-thaiaddress"
)

func main() {
	p, ok := thai.ProvinceByCode(10)
	if !ok {
		return
	}
	fmt.Println(p.NameTH, "/", p.NameEN) // กรุงเทพมหานคร / Bangkok
	fmt.Println(p.Region.NameEN())        // Central
}
```

### Navigate the hierarchy

```go
p, _ := thai.ProvinceByCode(50) // เชียงใหม่ / Chiang Mai

for _, d := range p.Districts() {
	fmt.Printf("%d %s\n", d.Code, d.NameTH)
	for _, s := range d.Subdistricts() {
		fmt.Printf("  %d %s (%05d)\n", s.Code, s.NameTH, s.Postcode)
	}
}

// Walk back up from a subdistrict:
s, _ := thai.SubdistrictByCode(100101)
fmt.Println(s.NameTH, "→", s.District().NameTH, "→", s.Province().NameTH)
```

### Find by postal code

```go
for _, s := range thai.ByPostcode(50200) {
	fmt.Printf("%s, %s, %s\n",
		s.NameTH, s.District().NameTH, s.Province().NameTH)
}

// All postcodes within a district:
fmt.Println(thai.PostcodesOf(5001)) // e.g. [50200 50300 ...]
```

### Autocomplete (prefix search)

```go
// SearchSubdistricts matches a normalized prefix — great for a type-ahead box.
for _, s := range thai.SearchSubdistricts("สุเทพ") {
	fmt.Printf("%s — %s, %s (%05d)\n",
		s.NameTH, s.District().NameTH, s.Province().NameTH, s.Postcode)
}
```

### Resolve a full address

```go
matches, err := thai.Resolve(thai.Query{
	Subdistrict: "สุเทพ",
	District:    "เมืองเชียงใหม่",
	Province:    "เชียงใหม่",
	Postcode:    50200,
})
if err != nil {
	// handle: no match / ambiguous input
}
for _, m := range matches {
	fmt.Printf("%s > %s > %s (%05d)\n",
		m.Province.NameTH, m.District.NameTH,
		m.Subdistrict.NameTH, m.Subdistrict.Postcode)
}
```

### Validate a code triple

```go
if err := thai.Validate(50, 5001, 500101); err != nil {
	fmt.Println("invalid address codes:", err)
}
```

## Caveat: duplicate names

795 subdistrict names repeat across different provinces (e.g. "ในเมือง"), so
name-based lookups (`FindSubdistricts`, `SearchSubdistricts`) return a **slice**,
not a single result. Disambiguate with district/province context or a postcode —
`Resolve(Query{...})` is built for exactly this.

Note also that Bangkok uses เขต/แขวง terminology (its district names carry the
"เขต" prefix), and that postal codes attach at the subdistrict level — a single
district can span several postcodes.

## Data source & license

This library is MIT-licensed (see [`LICENSE`](LICENSE)).

The embedded dataset is a reshaped snapshot (~2025-09) of
[`github.com/kongvut/thai-province-data`](https://github.com/kongvut/thai-province-data)
by Kongvut Sangkla, used under the MIT License. We derived official two-digit
province codes and slimmed the field layout, but it is the same underlying data.
See [`THIRD_PARTY_LICENSES.md`](THIRD_PARTY_LICENSES.md) for the upstream notice
and [`data/SOURCE.md`](data/SOURCE.md) for provenance and the regeneration recipe.
