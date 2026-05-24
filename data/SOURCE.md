# Data Source & Provenance

This document describes the origin, schema, and regeneration procedure for the
embedded dataset `data/areas.json`, which is compiled into the package via
`go:embed`. The package performs no network or filesystem access at runtime.

## Origin

- **Upstream:** [`github.com/kongvut/thai-province-data`](https://github.com/kongvut/thai-province-data)
- **License:** MIT License, Copyright (c) 2025 Kongvut Sangkla (see
  `../THIRD_PARTY_LICENSES.md`)
- **Snapshot:** approximately 2025-09 (`api/latest`)

`areas.json` is a reshaped, slimmed snapshot of the upstream `api/latest` JSON
files. The underlying data is unchanged; only the field layout was adjusted and
official two-digit province codes were derived (see below).

## Record counts

| Collection   | Thai           | Count |
|--------------|----------------|-------|
| Provinces    | จังหวัด        | 77    |
| Districts    | อำเภอ / เขต    | 930   |
| Subdistricts | ตำบล / แขวง    | 7,452 |

## Field schema of `areas.json`

The top-level object has a `source` string plus three arrays:

```jsonc
{
  "source": "github.com/kongvut/thai-province-data (MIT, snapshot 2025-09)",

  "provinces": [
    { "code": 10, "th": "กรุงเทพมหานคร", "en": "Bangkok", "region": 2 }
    // code:   official 2-digit DOPA province geocode (10–96)
    // th/en:  Thai / Latin-transcribed name
    // region: integer region id (see Region in region.go)
  ],

  "districts": [
    { "code": 1001, "province": 10, "th": "เขตพระนคร", "en": "Khet Phra Nakhon" }
    // code:     official 4-digit DOPA district geocode
    // province: owning province code (== code / 100)
    // th/en:    Thai / Latin-transcribed name
  ],

  "subdistricts": [
    { "code": 100101, "district": 1001, "th": "พระบรมมหาราชวัง",
      "en": "Phra Borom Maha Ratchawang", "zip": 10200 }
    // code:     official 6-digit DOPA subdistrict geocode
    // district: owning district code (== code / 100)
    // th/en:    Thai / Latin-transcribed name
    // zip:      5-digit Thai postal code
  ]
}
```

### Geocode hierarchy

DOPA geocodes are hierarchical and the parent is derivable by integer division:

- `districtCode / 100 == provinceCode`
- `subdistrictCode / 100 == districtCode`

Postal codes (`zip`) attach at the **subdistrict** level. A single district can
span several postal codes, so postcodes are not 1:1 with districts.

## How official province codes were derived

The upstream `province.json` uses an internal sequential `id` (1..77) that is
**not** the official DOPA province geocode. The official two-digit province code
is recovered from the district geocodes instead:

```
provinceCode = district.id / 100   (integer division, official DOPA geocode)
```

Each upstream province is matched to its districts (by the upstream
`province_id` foreign key), and the district geocode's high two digits give the
official province code (e.g. district `1001` -> province `10` = Bangkok).
District and subdistrict `id` values in the upstream data already are the
official 4- and 6-digit DOPA geocodes and are used directly as `code`.

## Regeneration recipe

To rebuild `areas.json` from a fresh upstream snapshot:

1. Fetch the three upstream `api/latest` files:

   ```sh
   BASE=https://raw.githubusercontent.com/kongvut/thai-province-data/master/api/latest
   curl -fsSL "$BASE/province.json"     -o province.json
   curl -fsSL "$BASE/district.json"     -o district.json
   curl -fsSL "$BASE/sub_district.json" -o sub_district.json
   ```

2. Reshape into the slim schema above:
   - **Provinces:** for each province, derive the official `code` from any of
     its districts (`districtCode / 100`); keep `name_th` -> `th`,
     `name_en` -> `en`; assign `region` (region grouping, see `region.go`).
   - **Districts:** use the upstream district `id` as `code`; set `province` to
     `code / 100`; keep `name_th` -> `th`, `name_en` -> `en`.
   - **Subdistricts:** use the upstream subdistrict `id` as `code`; set
     `district` to `code / 100`; keep `name_th` -> `th`, `name_en` -> `en`,
     `zip_code` -> `zip`.
   - Drop all other upstream fields (timestamps, foreign-key ids, etc.).

3. Write the result to `data/areas.json` with a `source` line recording the
   upstream project, license, and snapshot date.

4. Sanity-check the counts (77 / 930 / 7,452) and that every district/subdistrict
   parent resolves via integer division.
