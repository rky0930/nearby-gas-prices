# Opinet 레퍼런스 링크 모음 (PDF 대신)

## 오피넷 API 이용 안내
- 페이지: *오피넷 API 이용 안내*
- 링크: https://www.opinet.co.kr/user/custapi/custApiInfo.do

## 반경 내 주유소 검색 (aroundAll)
- 페이지: *오픈 API 정보 → 반경 내 주유소*
- 링크: https://www.opinet.co.kr/user/custapi/openApiInfoDtl.do?apiId=3

### 주요 포인트
- 엔드포인트: `https://www.opinet.co.kr/api/aroundAll.do`
- 인증 파라미터: `code`
- 출력: `out=xml|json`
- 좌표: `x`,`y` = KATEC
- 반경: `radius` 최대 5000(m)
- 유종: `prodcd` (예: `B027` 휘발유)
- 정렬: `sort` (1=가격순, 2=거리순)
