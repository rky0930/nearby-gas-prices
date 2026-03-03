# nearby-gas-prices

Find nearby gas stations and cheapest fuel prices in Korea using **KNOC Opinet** free APIs.

This repo contains:
- An **OpenClaw Skill** under `skill/nearby-gas-prices/`
- A small CLI script (bundled inside the skill)

## Data Sources

### Opinet (KNOC / 한국석유공사)
- Gas station price data comes from **Opinet** (operated by KNOC).
- We use the **free Open API** endpoint:
  - `https://www.opinet.co.kr/api/aroundAll.do` (nearby gas stations within radius)
- Note: prices are based on reported/updated snapshots and may differ from on-site prices.
- Official site: https://www.opinet.co.kr/

### OpenStreetMap Nominatim
- Place name → coordinates (WGS84) geocoding uses **OSM Nominatim**.
- You must follow the Nominatim usage policy.
- In practice you should set a proper User-Agent with contact info; otherwise you may get HTTP 403.

## Getting an OPINET_KEY

Opinet free API calls require an API key. In requests, this key is passed via the `code` parameter.

1. Visit https://www.opinet.co.kr/
2. Go to the Opinet API page: https://www.opinet.co.kr/user/custapi/custApiInfo.do
3. Sign up / log in
4. Apply for **free API** usage (무료 API 이용신청)
5. Get your API key (KEY/인증키) and set it as an environment variable:

```bash
export OPINET_KEY="<YOUR_KEY>"
```

Security note:
- Do **not** commit keys to git.
- Prefer environment variables (or `.env` ignored by git).

## Quickstart (CLI)

The CLI is bundled in the skill folder:

```bash
cd skill/nearby-gas-prices

# required
export OPINET_KEY="<YOUR_KEY>"

# required for Nominatim calls
export NOMINATIM_USER_AGENT='nearby-gas-prices/0.1 (contact: you@example.com)'

# search by place name
python3 scripts/opinet_nearby.py --query "소사역" --prodcd B027 --radius 5000 --top 5

# or use coordinates
python3 scripts/opinet_nearby.py --lat 37.48278 --lon 126.79565 --prodcd B027 --radius 5000 --top 5
```

## OpenClaw Skill

Skill source lives at:
- `skill/nearby-gas-prices/`

To distribute:
- Package into a `.skill` file (see OpenClaw docs / your environment tooling)
- Upload the `.skill` file to GitHub Releases

## Notes / Gotchas

- Opinet `aroundAll.do` API key parameter is **`code`** (not `certkey`).
- Opinet `x`,`y` are **KATEC** coordinates (not WGS84 lat/lon).

## License

MIT
