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

- (참고) Windows는 Releases에서 `nearby-gas-prices_windows_amd64.zip`을 내려받아 `nearby-gas-prices.exe`를 PATH에 두는 방식을 권장합니다.

_Skill에서도 이 CLI를 기반으로 실행 됩니다._

- Skill은 실행 시점에 CLI 설치 여부를 확인하고, 설치되어 있지 않다면 _설치할지 먼저 물어본 뒤, 동의시_ 설치를 진행하도록 만들어져 있습니다.

---

## 설정 (환경변수 / 설정 파일)

### 1) 환경변수 (기본 권장)

- `OPINET_KEY`: *(필수)* 오피넷 Open API 키  → [OPINET_KEY 발급 방법](#opinet_key-발급-방법)
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

### 3) OPINET_KEY 발급 방법

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

### 4) (선택) Skill로 사용

원하는 경우 이 프로젝트를 *스킬로 추가*해서, 대화형 요청을 `nearby-gas-prices` CLI 실행으로 연결할 수 있습니다.

> 아래 *사용 예시*의 "Skill로 조회(대화형)" 항목도 참고하세요.

- GitHub Releases의 `.skill` asset을 설치/임포트해서 사용
- 또는 https://skills.sh/ 의 `npx skills`로 설치

```bash
npx skills add https://github.com/rky0930/nearby-gas-prices/tree/main/skill/nearby-gas-prices
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

### 3) Skill로 조회 (대화형)

Skill이 설치되어 있다면, CLI 옵션을 직접 기억하지 않아도 아래처럼 대화형으로 호출할 수 있습니다.

예시 1) 휘발유 최저가 Top 3

```text
사용자: 소사역 근처 휘발유 최저가 주유소 top 3 알려줘

(스킬 응답 예시)
1) (주)역곡주유소 - 1645원 (2244m)
2) (주)명연에너지 시흥IC훼미리주유소 - 1655원 (3290m)
3) ㈜삼표에너지 삼표주유소 - 1655원 (2030m)
```

예시 2) 경유 최저가 Top 3

```text
사용자: 소사역 근처 경유 최저가 주유소 top 3 알려줘

(스킬 응답 예시)
1) (주)명연에너지 시흥IC훼미리주유소 - 1567원 (3290m)
2) (주)역곡주유소 - 1570원 (2244m)
3) ㈜삼표에너지 삼표주유소 - 1578원 (2030m)
```

예시 3) 선호 브랜드만 (예: GS칼텍스)

```text
사용자: 소사역 근처에서 GS칼텍스 주유소 중 휘발유 최저가 top 3 알려줘

(스킬 응답 예시)
1) ㈜삼표에너지 삼표주유소 - 1655원 (2030m)
2) 지에스칼텍스(주)옥길주유소 - 1665원 (2936m)
3) 큰사랑주유소 - 1668원 (4083m)
```

예시 4) 평균 대비 + 상세정보

```text
사용자: 소사역 근처 최저가 주유소가 전국 평균보다 얼마나 싼지랑 상세정보도 알려줘

(스킬 응답 예시)
- (주)역곡주유소: 1645원, 전국평균(1730.83) 대비 -86원
- 주소: 경기 부천시 원미구 부일로 820
- 전화: 032-345-5002
- 부가서비스: 세차=true, 품질인증=true
```

> 참고: 위 예시는 실행 시점에 따라 가격/순서/거리 등이 달라질 수 있습니다.

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
