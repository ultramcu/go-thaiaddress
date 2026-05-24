# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-05-25

### Added

- Initial release of `go-thaiaddress`, a pure-Go, MIT-licensed library of
  Thailand's administrative areas.
- Embedded dataset (via `go:embed`, no network or filesystem access at runtime):
  77 provinces, 930 districts, and 7,452 subdistricts with postal codes. Data is
  a reshaped snapshot (~2025-09) of `github.com/kongvut/thai-province-data` (MIT).
- Core types: `Province`, `District`, `Subdistrict`, and `Region` (six regions
  with `.NameTH()` / `.NameEN()`).
- Code lookups: `Provinces`, `Districts`, `Subdistricts`, `ProvinceByCode`,
  `DistrictByCode`, `SubdistrictByCode`.
- Hierarchy navigation: `DistrictsOf`, `SubdistrictsOf`, plus methods
  `Province.Districts`, `District.Province` / `Subdistricts` / `Postcodes`, and
  `Subdistrict.District` / `Province`. Parent links use official DOPA geocodes.
- Postal-code lookups: `ByPostcode` and `PostcodesOf`.
- Name search: `NormalizeName`, `FindProvinces` / `FindDistricts` /
  `FindSubdistricts` (exact, normalized) and `SearchProvinces` /
  `SearchDistricts` / `SearchSubdistricts` (prefix autocomplete).
- Validation: `Validate(provinceCode, districtCode, subdistrictCode)` checks the
  geocode hierarchy is internally consistent.
- Address resolution: `Resolve(Query)` matches free-form
  subdistrict/district/province/postcode fragments to `[]Match`.

[0.1.0]: https://github.com/ultramcu/go-thaiaddress/releases/tag/v0.1.0
