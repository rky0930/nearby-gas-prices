# Opinet 무료 API 요약 메모 (발췌)

출처: Opinet_API_Free.pdf (2024.09)

## 공통
- Base: `https://www.opinet.co.kr/api/...`
- 인증 파라미터: `code` (공사에서 부여한 키)
- 출력 형식: `out=xml|json`

## ⑨ 반경 내 주유소 (aroundAll)
- URL: `https://www.opinet.co.kr/api/aroundAll.do`
- 입력 좌표: `x`, `y`는 **KATEC**
- `radius`: 최대 5000 (m)
- `prodcd`: `B027` 휘발유, `D047` 경유, `B034` 고급휘발유, `C004` 등유, `K015` 부탄
- `sort`: `1` 가격순, `2` 거리순
- 반환: `OS_NM`(상호), `PRICE`, `DISTANCE`, `GIS_X_COOR`,`GIS_Y_COOR` 등

## 다른 조건 조회 (주요)
- 전국 평균가격(현재): `avgAllPrice.do`
- 시도별 평균가격(현재): `avgSidoPrice.do` (+ `sido`, `prodcd`)
- 시군구별 평균가격(현재): `avgSigunPrice.do` (+ `sido` 필수, `sigun`, `prodcd`)
- 최저가 Top20: (문서 항목 ⑧) 별도 endpoint
- 상호로 주유소 검색: (문서 항목 ⑪)
- 주유소 상세정보(ID): `detailById.do`

> 정확한 endpoint/파라미터는 PDF 원문을 우선으로 한다.
