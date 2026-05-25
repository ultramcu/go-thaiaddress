# Reference data

## `dopa-tambon.xlsx`

The authoritative subdistrict (tambon) coordinate dataset published by Thailand's
**Department of Provincial Administration (กรมการปกครอง, DOPA)**.

It is kept here only as a **reference / validation source** — it is *not* the
data the library serves, and it is *not* embedded in the Go build. The library's
embedded dataset (`../data/areas.json`) is derived from
[`kongvut/thai-province-data`](https://github.com/kongvut/thai-province-data)
(MIT); see [`../data/SOURCE.md`](../data/SOURCE.md).

### What it is

A point dataset (one or more rows per place, each with coordinates), columns:

| Column | Meaning |
| --- | --- |
| `AD_LEVEL` | administrative level (4 = subdistrict/tambon) |
| `TA_ID` | 6-digit subdistrict (tambon) geocode |
| `TAMBON_T` / `TAMBON_E` | subdistrict name (Thai / English) |
| `AM_ID` | 4-digit district (amphoe) geocode |
| `AMPHOE_T` / `AMPHOE_E` | district name (Thai / English) |
| `CH_ID` | 2-digit province (changwat) geocode |
| `CHANGWAT_T` / `CHANGWAT_E` | province name (Thai / English) |
| `LAT` / `LONG` | latitude / longitude of the point |

Note it carries **no postal code**.

### How it was used

It was the ground truth for validating the embedded geocodes and hierarchy
(2026-05): all 77 province codes matched; two non-administrative pseudo-districts
were dropped and one district code (อ.เวียงเก่า, ขอนแก่น) was corrected to the
official DOPA value. Full notes in [`../data/SOURCE.md`](../data/SOURCE.md).

### Attribution

Source: Department of Provincial Administration (กรมการปกครอง), Thailand.
Government administrative data. Included here for reference and verification; the
factual geocodes it contains are not subject to copyright. If you redistribute
this file, credit DOPA as the source.
