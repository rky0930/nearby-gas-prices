---
name: nearby-gas-prices
description: "Find nearby Korean gas stations and cheapest fuel prices using KNOC Opinet free API (aroundAll.do). Use when a user asks (KR/EN): nearby gas prices, cheapest gas within 5km, or queries like '주변 주유소', '근처 휘발유/경유 최저가', 'OO역/OO동 근처 주유소 가격'. This is an optional OpenClaw integration layer that runs the nearby-gas-prices Go binary CLI and summarizes results."
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

### 0) CLI 설치 확인 (중요)

이 스킬은 내부적으로 `nearby-gas-prices` CLI를 호출하는 방식이라, 먼저 CLI가 설치되어 있어야 한다.

- 설치 여부 확인:

```bash
command -v nearby-gas-prices
```

- 설치가 안 되어 있다면:
  - 먼저 사용자에게 “CLI를 설치할까요?”라고 물어보고, 동의할 때만 설치를 진행한다.
  - 설치 명령(예시):

```bash
curl -fsSL https://raw.githubusercontent.com/rky0930/nearby-gas-prices/main/install.sh | sh
```

### 1) 필수/조건부 설정

이 스킬은 내부적으로 `nearby-gas-prices` CLI를 호출하며, CLI 설정은 아래 2가지 방식 중 하나로 줄 수 있다.

*설정 파일(권장)*
- 경로: `~/.config/nearby-gas-prices/config.toml`
- 키:
  - `opinet_key` *(필수)* 오피넷 무료 API 키
  - `nominatim_user_agent` *(조건부)* `--query`(지명 검색) 사용 시 필요 (OSM Nominatim 정책상 contact 포함 권장, 미설정 시 403 가능)

*환경변수(옵션 / override 용)*
- `OPINET_KEY`
- `NOMINATIM_USER_AGENT`

> 둘 다 설정되어 있다면 현재 CLI는 **환경변수를 우선** 사용한다.

*설정 파일 템플릿 만들기(추천)*

```bash
mkdir -p ~/.config/nearby-gas-prices
cat > ~/.config/nearby-gas-prices/config.toml <<'TOML'
opinet_key = "YOUR_KEY"
# --query(지명 검색) 쓸 때만 필요 (contact 포함 권장)
# nominatim_user_agent = "nearby-gas-prices/0.1 (contact: you@example.com)"
TOML
chmod 600 ~/.config/nearby-gas-prices/config.toml
```

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

## API Key가 없을 때 (중요)

이 스킬/CLI는 기본적으로 오피넷 *Open API*를 호출합니다. 따라서 `OPINET_KEY`가 없으면 *API 호출은 할 수 없습니다*.

- 이 경우에는 오류만 내고 끝내기보다,
  - 사용자에게 “API Key 없이도 수동으로 확인할 수 있는 오피넷 페이지”를 안내하고
  - (OpenClaw 환경이라면) 내장 브라우저로 해당 페이지를 열어 탐색하도록 제안합니다.

수동 확인 링크:
- https://www.opinet.co.kr/searRgSelect.do

> 참고: Playwright/Selenium 등 브라우저 자동화로도 시도할 수는 있지만, 캡차/차단/레이아웃 변경에 취약해 *best effort*입니다.

## OPINET_KEY 발급 방법(요약)

`OPINET_KEY`는 자동/반복 조회(=API 호출)를 위해 필요합니다.

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
