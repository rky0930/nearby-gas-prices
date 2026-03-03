---
name: nearby-gas-prices
description: Find nearby Korean gas stations and cheapest fuel prices using KNOC Opinet free API (aroundAll.do). Use when a user asks in Korean/English for “주변 주유소”, “근처 휘발유/경유 최저가”, “OO역/OO동 근처 주유소 가격”, “within 5km gas price”. This skill is an optional OpenClaw integration layer: it triggers the `nearby-gas-prices` CLI and returns a human-readable Top-N summary (and optional JSON) using OPINET_KEY. Place-name search uses OSM Nominatim.
---

# nearby-gas-prices (OpenClaw Skill)

이 스킬은 “*자연어 요청 → CLI 실행 → 결과 요약 반환*”을 담당하는 *OpenClaw용 통합 레이어*야.

- 일반 사용자 보급은 **CLI(바이너리)** 가 메인
- OpenClaw를 쓰는 사람은 이 스킬로 “대화로” 편하게 호출하는 게 목적

## 핵심 포인트(실수 방지)

- `aroundAll.do` 인증 파라미터 이름은 **`code`** (NOT `certkey`)
- `x`,`y`는 **KATEC 좌표** (WGS84 위경도 아님)
- `radius` 최대 5000(m)
- 휘발유: `prodcd=B027`

## 빠른 사용

### 1) 필수/조건부 설정

- `OPINET_KEY`: *(필수)* 오피넷 무료 API 키
- `NOMINATIM_USER_AGENT`: *(조건부)* `--query`(지명 검색) 사용 시 필요 (OSM Nominatim 정책상 contact 포함 권장, 미설정 시 403 가능)

환경변수로 설정하거나,
`~/.config/nearby-gas-prices/config.toml`에 아래 키로 저장해도 된다:
- `opinet_key`
- `nominatim_user_agent`

### 2) CLI 실행 예시

- 지명으로:

```bash
nearby-gas-prices --query "부천 역곡역" --top 5
```

- 위경도로:

```bash
nearby-gas-prices --lat 37.48278 --lon 126.79565 --top 5
```

- 휘발유/경유/LPG 한 번에(섹션별 출력):

```bash
nearby-gas-prices --query "부천 역곡역" --prodcd all --top 5
```

- 선호 브랜드만 필터(예: S-OIL=SOL, SK에너지=SKE):

```bash
nearby-gas-prices --query "부천 역곡역" --brand SOL,SKE --top 5
```

- 평균 대비 얼마나 싼지(전국 평균가) + 상세정보(Top1):

```bash
nearby-gas-prices --query "부천 역곡역" --with-avg --detail --top 5
```

- JSON 출력:

```bash
nearby-gas-prices --lat 37.48278 --lon 126.79565 --top 5 --json
```

## 출력 해석

CLI는 기본적으로 사람이 읽기 좋은 텍스트를 출력하고, `--json`을 주면 JSON 배열을 출력한다.
- 각 항목 필드(요약): `name`, `price`, `distance_m`, `x`, `y`
- `x,y`는 오피넷 응답의 KATEC 좌표를 그대로 담는다.

## 네이버/구글 지도 링크 만들기

- WGS84 위경도(`lat,lon`)가 있으면:
  - Google: `https://www.google.com/maps?q=<lat>,<lon>`
  - Naver(v5): `https://map.naver.com/v5/search/<검색어>?c=<lon>,<lat>,15,0,0,0,dh`

> aroundAll 응답은 KATEC만 주므로, 서비스에서 지도 링크를 안정적으로 만들려면 KATEC→WGS84 역변환(또는 별도 지오코딩)을 추가 구현한다.

## OPINET_KEY 발급 방법(요약)

`OPINET_KEY`가 없으면 오피넷 API를 호출할 수 없어서 스킬/CLI가 동작하지 않습니다.

- 오피넷 접속: https://www.opinet.co.kr/
- 페이지 이동: *이용안내 → 오피넷 API(유가정보 API)*
  - 안내 링크: https://www.opinet.co.kr/user/custapi/custApiInfo.do
- 로그인/회원가입 후 **무료 API 이용신청** → **인증키/KEY 확인**
- 발급받은 키를 다음 중 하나로 설정
  - 환경변수: `OPINET_KEY`
  - 설정 파일: `~/.config/nearby-gas-prices/config.toml` 의 `opinet_key`

보안 주의:
- 키를 코드/레포/이슈/로그에 남기지 마세요.

## 참고 자료

- 상세 파라미터/다른 API 목록은 `references/opinet-free-api-notes.md` 참고

## 확장 아이디어(v2)

- Top20/평균가/상호검색/상세정보 등 다른 무료 API endpoint 추가
- 브랜드/셀프/부가서비스 필터는 웹 UI 기준으로는 별도 API 조합이 필요할 수 있음
- 결과 캐싱(예: 5~15분) + “마지막 갱신시각” 표시
