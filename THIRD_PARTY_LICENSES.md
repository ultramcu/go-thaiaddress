# Third-Party Licenses

`go-thaiaddress` embeds a dataset of Thailand's administrative areas. That data
is derived from a third-party project and is redistributed here under its
original license. The notice below is preserved as required by the MIT License.

---

## thai-province-data

- **Project:** https://github.com/kongvut/thai-province-data
- **License:** MIT License
- **Usage:** The embedded dataset in `data/areas.json` is a reshaped snapshot
  (provinces, districts, subdistricts, and postal codes) of this project's
  `api/latest` data. Field names were slimmed and official two-digit province
  codes were derived, but it is the same underlying data. See `data/SOURCE.md`
  for full provenance and the regeneration recipe.

```
MIT License

Copyright (c) 2025 Kongvut Sangkla

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
