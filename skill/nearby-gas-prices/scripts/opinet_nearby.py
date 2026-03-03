#!/usr/bin/env python3
"""Opinet aroundAll(반경내 주유소) 조회 + 최저가 출력.

- 입력: 장소명(query) 또는 위경도(lat/lon)
- 처리: OSM Nominatim으로 지오코딩(옵션) -> WGS84 -> KATEC 변환 -> Opinet aroundAll 호출

주의:
- Nominatim은 테스트/프로토타입 용도 권장(Usage Policy 준수). 대량 트래픽에는 카카오/네이버 등으로 교체.
- OPINET_KEY는 환경변수로 주입.

이 스크립트는 '코드 예시/스킬 자원'이며, 환경에 따라 좌표변환 정확도 검증이 필요할 수 있습니다.
"""

import argparse
import json
import os
import sys
import urllib.parse
import urllib.request
from math import atan, atanh, cos, exp, log, pi, sin, sqrt, tan

# --- Minimal proj-like transform: WGS84 -> KATEC (Bessel TM) ---
# This is a pragmatic implementation intended for this skill.
# Parameters match common KATEC definition used in many examples:
# +proj=tmerc +lat_0=38 +lon_0=128 +k=0.9999 +x_0=400000 +y_0=600000
# +ellps=bessel +towgs84=-146.43,507.89,681.46
#
# NOTE: This ignores full 7-parameter datum shift rotations; many client implementations also omit.
# For best accuracy, replace with a proper CRS transform library (proj/pyproj) in production.

def wgs84_to_katec(lon_deg: float, lat_deg: float):
    # Bessel 1841
    a = 6377397.155
    f = 1 / 299.1528128
    e2 = 2 * f - f * f

    lat0 = 38.0 * pi / 180.0
    lon0 = 128.0 * pi / 180.0
    k0 = 0.9999
    x0 = 400000.0
    y0 = 600000.0

    lat = lat_deg * pi / 180.0
    lon = lon_deg * pi / 180.0

    def meridional_arc(phi):
        n = f / (2 - f)
        A = a / (1 + n) * (1 + n**2 / 4 + n**4 / 64)
        alpha = [
            None,
            n / 2 - 2 * n**2 / 3 + 5 * n**3 / 16 + 41 * n**4 / 180,
            13 * n**2 / 48 - 3 * n**3 / 5 + 557 * n**4 / 1440,
            61 * n**3 / 240 - 103 * n**4 / 140,
            49561 * n**4 / 161280,
        ]
        s = A * (
            phi
            + alpha[1] * sin(2 * phi)
            + alpha[2] * sin(4 * phi)
            + alpha[3] * sin(6 * phi)
            + alpha[4] * sin(8 * phi)
        )
        return s

    # Compute
    N = a / sqrt(1 - e2 * sin(lat) ** 2)
    T = tan(lat) ** 2
    C = (e2 / (1 - e2)) * cos(lat) ** 2
    A_ = (lon - lon0) * cos(lat)

    M = meridional_arc(lat)
    M0 = meridional_arc(lat0)

    x = x0 + k0 * N * (
        A_
        + (1 - T + C) * A_**3 / 6
        + (5 - 18 * T + T**2 + 72 * C - 58 * (e2 / (1 - e2))) * A_**5 / 120
    )

    y = y0 + k0 * (
        (M - M0)
        + N * tan(lat) * (
            A_**2 / 2
            + (5 - T + 9 * C + 4 * C**2) * A_**4 / 24
            + (61 - 58 * T + T**2 + 600 * C - 330 * (e2 / (1 - e2))) * A_**6 / 720
        )
    )

    return x, y


def nominatim_geocode(query: str, user_agent: str):
    url = (
        "https://nominatim.openstreetmap.org/search?format=json&limit=1&q="
        + urllib.parse.quote(query)
    )
    req = urllib.request.Request(url, headers={"User-Agent": user_agent})
    with urllib.request.urlopen(req, timeout=20) as r:
        data = json.load(r)
    if not data:
        raise RuntimeError(f"No geocode result for query: {query}")
    return float(data[0]["lat"]), float(data[0]["lon"]), data[0].get("display_name")


def opinet_aroundall(code: str, x: float, y: float, radius: int, prodcd: str, sort: int, out: str = "json"):
    base = "https://www.opinet.co.kr/api/aroundAll.do"
    params = {
        "code": code,
        "out": out,
        "x": f"{x:.1f}",
        "y": f"{y:.1f}",
        "radius": str(radius),
        "prodcd": prodcd,
        "sort": str(sort),
    }
    url = base + "?" + urllib.parse.urlencode(params)
    req = urllib.request.Request(url, headers={"User-Agent": "openclaw-opinet-skill/1.0"})
    with urllib.request.urlopen(req, timeout=20) as r:
        body = r.read().decode("utf-8")
    j = json.loads(body)
    oil = j.get("RESULT", {}).get("OIL", [])
    # normalize
    for o in oil:
        for k in ["PRICE", "DISTANCE", "GIS_X_COOR", "GIS_Y_COOR"]:
            if k in o and isinstance(o[k], str):
                try:
                    o[k] = float(o[k]) if "." in o[k] else int(o[k])
                except Exception:
                    pass
    return oil


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--query", help="장소명(예: 소사역)")
    ap.add_argument("--lat", type=float, help="위도(WGS84)")
    ap.add_argument("--lon", type=float, help="경도(WGS84)")
    ap.add_argument("--radius", type=int, default=5000, help="반경(m), 최대 5000")
    ap.add_argument("--prodcd", default="B027", help="유종코드 (기본: B027=휘발유)")
    ap.add_argument("--sort", type=int, default=1, help="1=가격순, 2=거리순")
    ap.add_argument("--top", type=int, default=5, help="상위 N개 출력")
    ap.add_argument(
        "--user-agent",
        default=os.getenv("NOMINATIM_USER_AGENT", ""),
        help="Nominatim 호출용 User-Agent (예: 'myapp/1.0 (contact: me@domain.com)'; env NOMINATIM_USER_AGENT 사용 가능)",
    )

    args = ap.parse_args()

    code = os.getenv("OPINET_KEY")
    if not code:
        print("ERROR: OPINET_KEY env var not set", file=sys.stderr)
        sys.exit(2)

    if args.radius > 5000:
        print("ERROR: radius max is 5000", file=sys.stderr)
        sys.exit(2)

    if args.query:
        if not args.user_agent:
            print(
                "ERROR: Nominatim 사용 시 User-Agent에 실제 연락처가 포함되어야 합니다.\n"
                "  예) export NOMINATIM_USER_AGENT='myapp/1.0 (contact: me@domain.com)'\n"
                "  또는: --user-agent 'myapp/1.0 (contact: me@domain.com)'",
                file=sys.stderr,
            )
            sys.exit(2)
        lat, lon, disp = nominatim_geocode(args.query, args.user_agent)
    else:
        if args.lat is None or args.lon is None:
            print("ERROR: provide --query or both --lat and --lon", file=sys.stderr)
            sys.exit(2)
        lat, lon, disp = args.lat, args.lon, None

    x, y = wgs84_to_katec(lon, lat)

    oil = opinet_aroundall(code=code, x=x, y=y, radius=args.radius, prodcd=args.prodcd, sort=args.sort)
    oil_sorted = sorted(oil, key=lambda o: (o.get("PRICE", 10**9), o.get("DISTANCE", 10**9)))

    out = {
        "input": {"query": args.query, "lat": lat, "lon": lon, "display_name": disp},
        "katec": {"x": x, "y": y},
        "count": len(oil_sorted),
        "top": oil_sorted[: args.top],
    }

    print(json.dumps(out, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
