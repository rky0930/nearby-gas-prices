# nearby-gas-prices

한국석유공사 **오피넷(Opinet)** 무료 API를 이용해, 특정 위치(장소명/위경도) 기준 *주변 주유소 가격*과 *최저가 주유소*를 찾는 프로젝트입니다.

이 레포에는:
- OpenClaw용 **Skill** (`skill/nearby-gas-prices/`)
- (스킬에 포함된) 간단한 CLI 스크립트

이 들어있습니다.

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

## 빠른 시작 (CLI)

CLI 스크립트는 스킬 폴더에 포함되어 있습니다.

```bash
cd skill/nearby-gas-prices

# 필수
export OPINET_KEY="<YOUR_KEY>"

# Nominatim 사용 시 필수 (contact 포함 권장)
export NOMINATIM_USER_AGENT='nearby-gas-prices/0.1 (contact: you@example.com)'

# 1) 장소명으로 조회
python3 scripts/opinet_nearby.py --query "소사역" --prodcd B027 --radius 5000 --top 5

# 2) 위경도로 조회
python3 scripts/opinet_nearby.py --lat 37.48278 --lon 126.79565 --prodcd B027 --radius 5000 --top 5
```

---

## OpenClaw Skill

- 스킬 소스: `skill/nearby-gas-prices/`
- 배포: `.skill` 파일로 패키징 후 GitHub Releases에 업로드하는 방식을 권장합니다.

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
