# nearby-gas-prices

한국석유공사 **오피넷(Opinet)** 무료 API를 이용해, 특정 위치(장소명/위경도) 기준 *주변 주유소 가격*과 *최저가 주유소*를 찾는 프로젝트입니다.

이 레포에는:
- GitHub Releases로 배포되는 **바이너리 CLI** (`nearby-gas-prices`)
- (선택) 대화형 에이전트에서 쓰고 싶을 때를 위한 Skill 소스 (`skill/nearby-gas-prices/`)

가 들어있습니다.

---

## 데이터 원천 (Data Sources)

### 1) 오피넷 (Opinet, 한국석유공사)
- 주유소 가격/주유소 목록 데이터는 **한국석유공사 오피넷(Opinet)** 에서 제공합니다.
- 본 프로젝트는 오피넷의 *무료 Open API* 중 아래 엔드포인트를 사용합니다.
  - **반경 내 주유소 검색**: `https://www.opinet.co.kr/api/aroundAll.do`
- 참고:
  - 오피넷 가격은 “주유소가 보고/갱신한 시점의 정보”라서 *현장 판매가와 차이*가 있을 수 있습니다.
- 공식 사이트: https://www.opinet.co.kr/

### 2) OpenStreetMap Nominatim
- 장소명(예: “소사역”)을 위경도(WGS84)로 바꾸는 지오코딩에 **OSM Nominatim**을 사용합니다.
- Nominatim은 Usage Policy를 준수해야 하며, **User-Agent에 contact 정보가 포함**되지 않으면 403이 날 수 있습니다.

---

## 설치 (CLI)

### 바이너리 CLI 설치 (Releases + install.sh)

릴리즈된 바이너리를 설치하면 Python 없이도 실행할 수 있습니다.

```bash
curl -fsSL https://raw.githubusercontent.com/rky0930/nearby-gas-prices/main/install.sh | sh

# 설치 확인
nearby-gas-prices --help
```

- Windows는 Releases에서 `nearby-gas-prices_windows_amd64.zip`을 내려받아 `nearby-gas-prices.exe`를 PATH에 두는 방식을 권장합니다.

*Skill도 결국 이 CLI를 실행하는 방식입니다.*

- *CLI가 설치되어 있어야 Skill이 동작합니다.*
- Skill은 실행 시점에 CLI 설치 여부를 확인하고, 미설치라면 *설치할지 먼저 물어본 뒤* 동의할 때만 설치를 진행하도록 안내합니다.

---

## 설정 (환경변수 / 설정 파일)

### 1) 환경변수 (기본 권장)

- `OPINET_KEY`: *(필수)* 오피넷 Open API 키
- `NOMINATIM_USER_AGENT`: *(조건부)* `--query`(장소명 검색) 사용 시 필요
  - OSM Nominatim 정책상 User-Agent에 앱/연락처(contact) 포함을 권장하며, 미설정/부적절 시 403이 날 수 있습니다.

예시:

```bash
export OPINET_KEY="<YOUR_KEY>"
export NOMINATIM_USER_AGENT="nearby-gas-prices/0.1 (contact: you@example.com)"
```

### 2) 설정 파일 (`~/.config/...`)

일부 에이전트/런타임 환경에서는 프로세스 환경변수 상속이 까다로울 수 있어, 설정 파일에서 키를 읽는 방식도 지원합니다.

- 경로: `~/.config/nearby-gas-prices/config.toml`
- 예시:

```toml
opinet_key = "YOUR_KEY"
nominatim_user_agent = "nearby-gas-prices/0.1 (contact: you@example.com)"
```

보안 권장:

```bash
chmod 600 ~/.config/nearby-gas-prices/config.toml
```

---

## 사용 예시

### 1) 지명으로 조회 (`--query`)

```bash
# 예: 소사역 근처
nearby-gas-prices --query "소사역" --top 3

# "근처" 같은 표현도 동작할 수 있지만, 지오코딩 결과가 애매할 수 있어
# 보통은 "역곡역", "부천 역곡역"처럼 지명/역 이름을 권장합니다.
```

예시 출력:

```text
기준 좌표(WGS84): 37.482766, 126.795590
검색 반경: 5000m (오피넷 aroundAll.do 최대 5000m)

1) (주)역곡주유소
   가격: 1645원
   거리: 2244m
   브랜드: ETC
   지도(네이버): https://map.naver.com/v5/search/%28%EC%A3%BC%29%EC%97%AD%EA%B3%A1%EC%A3%BC%EC%9C%A0%EC%86%8C?c=126.819282,37.489991,15,0,0,0,dh

2) ㈜삼표에너지 삼표주유소
   가격: 1655원
   거리: 2030m
   브랜드: GSC
   지도(네이버): https://map.naver.com/v5/search/%E3%88%9C%EC%82%BC%ED%91%9C%EC%97%90%EB%84%88%EC%A7%80+%EC%82%BC%ED%91%9C%EC%A3%BC%EC%9C%A0%EC%86%8C?c=126.778113,37.494629,15,0,0,0,dh

3) (주)명연에너지 시흥IC훼미리주유소
   가격: 1655원
   거리: 3290m
   브랜드: HDO
   지도(네이버): https://map.naver.com/v5/search/%28%EC%A3%BC%29%EB%AA%85%EC%97%B0%EC%97%90%EB%84%88%EC%A7%80+%EC%8B%9C%ED%9D%A5IC%ED%9B%BC%EB%AF%B8%EB%A6%AC%EC%A3%BC%EC%9C%A0%EC%86%8C?c=126.795755,37.453109,15,0,0,0,dh
```

> 참고: `--query`는 OSM Nominatim을 사용하므로 `NOMINATIM_USER_AGENT`(또는 설정 파일의 `nominatim_user_agent`)가 필요합니다.

### 2) 위경도로 조회 (예: 소사역 근처)

```bash
nearby-gas-prices --lat 37.48278 --lon 126.79565 --top 3
```

예시 출력:

```text
기준 좌표(WGS84): 37.482780, 126.795650
검색 반경: 5000m (오피넷 aroundAll.do 최대 5000m)

1) (주)역곡주유소
   가격: 1645원
   거리: 2238m
   지도(네이버): https://map.naver.com/v5/search/%28%EC%A3%BC%29%EC%97%AD%EA%B3%A1%EC%A3%BC%EC%9C%A0%EC%86%8C?c=126.819282,37.489991,15,0,0,0,dh

2) (주)명연에너지 시흥IC훼미리주유소
   가격: 1655원
   거리: 3292m
   지도(네이버): https://map.naver.com/v5/search/%28%EC%A3%BC%29%EB%AA%85%EC%97%B0%EC%97%90%EB%84%88%EC%A7%80+%EC%8B%9C%ED%9D%A5IC%ED%9B%BC%EB%AF%B8%EB%A6%AC%EC%A3%BC%EC%9C%A0%EC%86%8C?c=126.795755,37.453109,15,0,0,0,dh

3) ㈜삼표에너지 삼표주유소
   가격: 1655원
   거리: 2033m
   지도(네이버): https://map.naver.com/v5/search/%E3%88%9C%EC%82%BC%ED%91%9C%EC%97%90%EB%84%88%EC%A7%80+%EC%82%BC%ED%91%9C%EC%A3%BC%EC%9C%A0%EC%86%8C?c=126.778113,37.494629,15,0,0,0,dh
```

### 3) JSON 출력

```bash
nearby-gas-prices --lat 37.48278 --lon 126.79565 --top 2 --json
```

예시 출력:

```json
[
  {
    "name": "(주)역곡주유소",
    "price": 1645,
    "distance_m": 2238.1,
    "x": 295782.1,
    "y": 543747.1
  },
  {
    "name": "㈜삼표에너지 삼표주유소",
    "price": 1655,
    "distance_m": 2032.8,
    "x": 292147.9736,
    "y": 544308.2397
  }
]
```

> 참고: 위 예시는 실행 시점에 따라 가격/순서/거리 등이 달라질 수 있습니다.

---

## OPINET_KEY 발급 방법

오피넷 무료 API를 호출하려면 API 키가 필요합니다. 오피넷 API 요청에서는 이 키를 **`code` 파라미터**로 전달합니다.

1. 오피넷 접속: https://www.opinet.co.kr/
2. 페이지 이동: *이용안내 → 오피넷 API(유가정보 API)*
   - 안내 페이지 링크: https://www.opinet.co.kr/user/custapi/custApiInfo.do
3. 로그인/회원가입
4. **무료 API 이용신청** 진행
5. 발급된 KEY(인증키)를 환경변수로 설정

```bash
export OPINET_KEY="<YOUR_KEY>"
```

보안 주의:
- 키를 코드/레포/이슈/로그에 남기지 마세요.
- 환경변수 사용을 권장합니다(또는 `.env`를 사용하되 git에는 커밋하지 않기).

---

## (선택) Skill로 사용

원하는 경우 이 프로젝트를 *스킬로 추가*해서, 대화형 요청을 `nearby-gas-prices` CLI 실행으로 연결할 수 있습니다.

### 스킬로 할 수 있는 것

- "소사역 근처 최저가 주유소 알려줘" 같은 요청을 *자연어 → CLI 실행*으로 연결
- `--query`/`--lat`/`--lon`, `--prodcd all`, `--brand`, `--with-avg`, `--detail` 같은 옵션을 AI가 상황에 맞게 조합
- 에이전트 환경에서 환경변수 상속이 애매할 때 `~/.config/nearby-gas-prices/config.toml` 방식 안내

### 설치 방법(권장: Releases에서 가져오기)

사용자가 다시 패키징할 필요 없이, *Releases에 업로드된 `.skill` 파일*을 그대로 가져가서 설치/임포트하면 됩니다.

- 스킬 소스(개발/커스터마이즈용): `skill/nearby-gas-prices/`
- 스킬 배포(사용자 설치용): GitHub Releases의 `.skill` asset

추가로, Agent Skills 디렉토리(https://skills.sh/)에서 쓰는 `npx skills`로도 설치할 수 있습니다:

```bash
npx skills add https://github.com/rky0930/nearby-gas-prices/tree/main/skill/nearby-gas-prices
```

> 참고: 이 레포에서는 *일반 사용자 보급*을 위해 CLI 설치를 우선 안내합니다.

---

## API 제한 / 주의사항

- 오피넷 `aroundAll.do`의 키 파라미터는 **`code`** 입니다 (`certkey` 아님).
- `aroundAll.do`의 `x`,`y` 좌표는 **KATEC** 입니다(WGS84 위경도 아님).
- `radius`는 오피넷 API 문서에 **최대 5000m**로 명시되어 있어 더 크게 요청할 수 없습니다.
  - 참고(페이지명: *오픈 API 정보 → 반경 내 주유소*): https://www.opinet.co.kr/user/custapi/openApiInfoDtl.do?apiId=3
- Nominatim은 **User-Agent에 contact 포함**이 사실상 필수이며(403 방지), Usage Policy를 준수해야 합니다.

---

## 레퍼런스 (PDF 대신 링크)

- 오피넷 API 이용 안내: https://www.opinet.co.kr/user/custapi/custApiInfo.do
- 오피넷 API 상세(반경 내 주유소): https://www.opinet.co.kr/user/custapi/openApiInfoDtl.do?apiId=3
  - 페이지명: *오픈 API 정보 → 반경 내 주유소*
- 오피넷 공식 사이트: https://www.opinet.co.kr/

---

## 라이선스

MIT
