---
name: nearby-gas-prices
description: Find nearby Korean gas stations and cheapest fuel prices using KNOC Opinet free API (aroundAll.do). Use when a user asks in Korean/English for “주변 주유소”, “근처 휘발유/경유 최저가”, “OO역/OO동 근처 주유소 가격”, “within 5km gas price”, or wants Opinet API usage for location-based gas prices. Supports place-name search via OSM Nominatim (prototype) and calls Opinet with OPINET_KEY (code parameter), converts WGS84 lat/lon to KATEC x/y, returns Top-N with Naver/Google map links.
---

# Opinet 근처 주유소 최저가 조회

## 핵심 포인트(실수 방지)

- `aroundAll.do` 인증 파라미터 이름은 **`code`** (NOT `certkey`)
- `x`,`y`는 **KATEC 좌표** (WGS84 위경도 아님)
- `radius` 최대 5000(m)
- 휘발유: `prodcd=B027`

## 빠른 사용(로컬)

1) 환경변수 설정

- `OPINET_KEY` (오피넷 무료 API 키)

2) 스크립트 실행

- 장소명으로 (Nominatim은 User-Agent에 실제 연락처 포함 필요):

```bash
export NOMINATIM_USER_AGENT='myapp/1.0 (contact: me@domain.com)'
python3 scripts/opinet_nearby.py --query "소사역" --prodcd B027 --radius 5000 --top 5
```

- 위경도로:

```bash
python3 scripts/opinet_nearby.py --lat 37.48278 --lon 126.79565 --prodcd B027 --radius 5000 --top 5
```

## 출력 해석

스크립트는 JSON을 출력한다.
- `top[0]`가 최저가(가격→거리 순)
- 각 항목 필드: `OS_NM`(상호), `PRICE`(가격), `DISTANCE`(m), `GIS_X_COOR`,`GIS_Y_COOR`(KATEC), `UNI_ID`, `POLL_DIV_CD`

## 네이버/구글 지도 링크 만들기

- WGS84 위경도(`lat,lon`)가 있으면:
  - Google: `https://www.google.com/maps?q=<lat>,<lon>`
  - Naver(v5): `https://map.naver.com/v5/search/<검색어>?c=<lon>,<lat>,15,0,0,0,dh`

> aroundAll 응답은 KATEC만 주므로, 서비스에서 지도 링크를 안정적으로 만들려면 KATEC→WGS84 역변환(또는 별도 지오코딩)을 추가 구현한다.

## OPINET_KEY 발급 방법(요약)

- 오피넷 접속: https://www.opinet.co.kr/
- 오피넷 API 안내: https://www.opinet.co.kr/user/custapi/custApiInfo.do
- 로그인 후 **무료 API 이용신청** → **인증키/KEY 확인**
- 키는 코드/레포에 하드코딩하지 말고 환경변수로만 보관

## 참고 자료

- 상세 파라미터/다른 API 목록은 `references/opinet-free-api-notes.md` 참고 (원문: Opinet_API_Free.pdf 발췌)

## 확장 아이디어(v2)

- Top20/평균가/상호검색/상세정보 등 다른 무료 API endpoint 추가
- 브랜드/셀프/부가서비스 필터는 웹 UI 기준으로는 별도 API 조합이 필요할 수 있음
- 결과 캐싱(예: 5~15분) + “마지막 갱신시각” 표시
