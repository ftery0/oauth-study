#!/bin/bash
# App1 + IdP 통합 E2E. 실제 사용자 흐름을 curl 로 재현 (app1 /login 부터 시작):
#   1. app1 /login → IdP /authorize 로 redirect (state app1 세션에 저장)
#   2. 회원가입 폼 → 가입 → callback (state 검증) → 토큰 교환 → app1 세션
#   3. /api/me 인증 확인
#   4. Welcome seed 검증 (notebooks 2개 + notes 5개)
#   5. Notebook/Note CRUD + 검색
#   6. 로그아웃 체인 + 세션 폐기 확인

set -u

IDP="http://localhost:8080"
APP1="http://localhost:8011"
FRONT="http://localhost:5181"
JAR="/tmp/app1-e2e-jar"
TMP="/tmp/app1-e2e"
mkdir -p "$TMP"
rm -f "$JAR"

PASS=0; FAIL=0
declare -a FAILED_TESTS
pass() { echo "  ✅ $1"; PASS=$((PASS+1)); }
fail() { echo "  ❌ $1 — $2"; FAIL=$((FAIL+1)); FAILED_TESTS+=("$1"); }
section() { echo ""; echo "▶ $1"; }

USERNAME="e2e$(date +%s | tail -c 8)"
PASSWORD="testPass123!"
EMAIL="${USERNAME}@example.com"
echo "테스트 사용자: ${USERNAME}"

# helper: JSON 파싱 (인자 = 키)
jq_get() { python3 -c "import json,sys; d=json.loads(sys.stdin.read()); print(d.get('$1','') if isinstance(d,dict) else '')"; }
jq_count() { python3 -c "import json,sys; print(len(json.loads(sys.stdin.read())))"; }

# ─────────────────────────────────────────────
section "1. app1 /login → IdP /authorize (자동 follow, state 전파)"

# -L: follow redirects. 끝나면 IdP login form HTML 받음.
curl -s -L -c "$JAR" -b "$JAR" "${APP1}/login" -o "$TMP/login.html" \
  -w "최종 URL: %{url_effective}\n"

grep -q 'name="csrf_token"' "$TMP/login.html" && pass "app1/login → IdP authorize → 로그인 폼 렌더" || { fail "login form" "안 나옴"; exit 1; }

# ─────────────────────────────────────────────
section "2. 회원가입 폼 → 가입 → callback → 세션 발급"

# 로그인 폼의 회원가입 링크 URL 추출 (state 자동 포함됨)
REG_URL=$(grep -oE 'href="/oauth/register[^"]+"' "$TMP/login.html" | head -1 | sed 's/^href="\(.*\)"$/\1/')
[ -n "$REG_URL" ] && pass "회원가입 링크 추출" || fail "register link" "없음"

# 회원가입 폼 페이지 GET
curl -s -L -c "$JAR" -b "$JAR" "${IDP}${REG_URL}" -o "$TMP/register.html"
REG_CSRF=$(grep -oE 'name="csrf_token"\s+value="[^"]+"' "$TMP/register.html" | head -1 | sed 's/.*value="\([^"]*\)".*/\1/')
# 폼의 hidden state/client_id 추출
STATE=$(grep -oE 'name="state"\s+value="[^"]*"' "$TMP/register.html" | head -1 | sed 's/.*value="\([^"]*\)".*/\1/')
CLIENT_ID=$(grep -oE 'name="client_id"\s+value="[^"]*"' "$TMP/register.html" | head -1 | sed 's/.*value="\([^"]*\)".*/\1/')
REDIRECT_URI=$(grep -oE 'name="redirect_uri"\s+value="[^"]*"' "$TMP/register.html" | head -1 | sed 's/.*value="\([^"]*\)".*/\1/')
SCOPE=$(grep -oE 'name="scope"\s+value="[^"]*"' "$TMP/register.html" | head -1 | sed 's/.*value="\([^"]*\)".*/\1/')

[ -n "$REG_CSRF" ] && [ -n "$STATE" ] && [ "$CLIENT_ID" = "app1" ] \
  && pass "register 폼 hidden 값들 OK (state=$STATE)" \
  || fail "register form fields" "csrf=$REG_CSRF state=$STATE cid=$CLIENT_ID"

# POST 회원가입 — redirect 수동 따라가기 (POST→302→GET callback 보장)
REG_RESP=$(curl -s -i -c "$JAR" -b "$JAR" \
  -X POST "${IDP}/oauth/register" \
  --data-urlencode "username=${USERNAME}" \
  --data-urlencode "email=${EMAIL}" \
  --data-urlencode "password=${PASSWORD}" \
  --data-urlencode "password_confirm=${PASSWORD}" \
  --data-urlencode "csrf_token=${REG_CSRF}" \
  --data-urlencode "client_id=${CLIENT_ID}" \
  --data-urlencode "redirect_uri=${REDIRECT_URI}" \
  --data-urlencode "state=${STATE}" \
  --data-urlencode "scope=${SCOPE}")

CB_URL=$(echo "$REG_RESP" | grep -i '^location:' | tr -d '\r' | sed 's/^[Ll]ocation: //' | head -1)
[[ "$CB_URL" == *"code="* ]] && pass "register POST → callback URL ($(echo "$CB_URL" | sed 's/code=[^&]*/code=***/'))" || { fail "register redirect" "$REG_RESP"; exit 1; }

# GET callback — app1 가 state 검증 + 토큰 교환 + 세션 attribute 세팅
CB_RESP=$(curl -s -i -c "$JAR" -b "$JAR" "$CB_URL")
CB_LOC=$(echo "$CB_RESP" | grep -i '^location:' | tr -d '\r' | sed 's/^[Ll]ocation: //' | head -1)
[ "$CB_LOC" = "${FRONT}" ] && pass "app1 callback → frontend redirect" || fail "callback location" "$CB_LOC"

JSESSION=$(grep -i "JSESSIONID" "$JAR" 2>/dev/null | head -1)
[ -n "$JSESSION" ] && pass "app1 JSESSIONID 쿠키 발급" || { fail "JSESSIONID" "없음"; exit 1; }

# ─────────────────────────────────────────────
section "3. /api/me — 신규 사용자 인증 정보"

ME=$(curl -s -b "$JAR" "${APP1}/api/me")
SUB=$(echo "$ME" | jq_get sub)
PREFERRED=$(echo "$ME" | jq_get preferred_username)
NAME=$(echo "$ME" | jq_get name)

[ -n "$SUB" ] && pass "/api/me sub 노출 (${SUB:0:8}...)" || fail "/api/me sub" "$ME"
[ "$PREFERRED" = "$USERNAME" ] && pass "/api/me preferred_username=${USERNAME}" || fail "preferred_username" "got '$PREFERRED'"
[ "$NAME" = "$USERNAME" ] && pass "/api/me name 동기화 (display_name=${USERNAME})" || fail "name sync" "got '$NAME'"

# ─────────────────────────────────────────────
section "4. Welcome Seed 실제 트리거 검증"

NOTEBOOKS=$(curl -s -b "$JAR" "${APP1}/api/notebooks")
NB_COUNT=$(echo "$NOTEBOOKS" | jq_count)
NB_TITLES=$(echo "$NOTEBOOKS" | python3 -c "import json,sys; ns=json.loads(sys.stdin.read()); print('|'.join(sorted(n['title'] for n in ns)))" 2>/dev/null)

[ "$NB_COUNT" = "2" ] && pass "Welcome seed: notebooks 2 개" || fail "notebook count" "${NB_COUNT}개 ($NOTEBOOKS)"
[[ "$NB_TITLES" == *"시작하기"* && "$NB_TITLES" == *"프로젝트 메모"* ]] \
  && pass "Welcome seed: '시작하기' + '프로젝트 메모'" \
  || fail "notebook titles" "$NB_TITLES"

if [ "$NB_COUNT" = "2" ]; then
  STARTED_ID=$(echo "$NOTEBOOKS" | python3 -c "import json,sys; ns=json.loads(sys.stdin.read()); print(next(n['id'] for n in ns if n['title']=='시작하기'))")
  PROJECT_ID=$(echo "$NOTEBOOKS" | python3 -c "import json,sys; ns=json.loads(sys.stdin.read()); print(next(n['id'] for n in ns if n['title']=='프로젝트 메모'))")

  STARTED=$(curl -s -b "$JAR" "${APP1}/api/notes?notebookId=${STARTED_ID}")
  PROJECT=$(curl -s -b "$JAR" "${APP1}/api/notes?notebookId=${PROJECT_ID}")
  SC=$(echo "$STARTED" | jq_count); PC=$(echo "$PROJECT" | jq_count)

  [ "$SC" = "3" ] && pass "'시작하기' 노트 3 개 (환영/마크다운/팁)" || fail "started notes" "$SC개"
  [ "$PC" = "2" ] && pass "'프로젝트 메모' 노트 2 개" || fail "project notes" "$PC개"

  WELCOME_BODY=$(echo "$STARTED" | python3 -c "import json,sys; ns=json.loads(sys.stdin.read()); print(next(n['bodyMd'] for n in ns if n['title']=='환영합니다'))")
  [[ "$WELCOME_BODY" == *"Markdown"* && "$WELCOME_BODY" == *"자동 저장"* ]] \
    && pass "'환영합니다' 노트 본문 마크다운 OK" \
    || fail "welcome content" "본문 검증 실패"
fi

# ─────────────────────────────────────────────
section "5. Notebook CRUD"

NEW_NB=$(curl -s -b "$JAR" -X POST "${APP1}/api/notebooks" \
  -H "Content-Type: application/json" -d '{"title":"E2E 테스트 노트북"}')
NEW_NB_ID=$(echo "$NEW_NB" | jq_get id)
[ -n "$NEW_NB_ID" ] && pass "POST /api/notebooks (id=${NEW_NB_ID})" || fail "create notebook" "$NEW_NB"

if [ -n "$NEW_NB_ID" ]; then
  RENAMED=$(curl -s -b "$JAR" -X PATCH "${APP1}/api/notebooks/${NEW_NB_ID}" \
    -H "Content-Type: application/json" -d '{"title":"이름 변경됨"}')
  RN_TITLE=$(echo "$RENAMED" | jq_get title)
  [ "$RN_TITLE" = "이름 변경됨" ] && pass "PATCH /api/notebooks/{id} 이름 변경" || fail "rename" "got '$RN_TITLE'"

  CODE=$(curl -s -o /dev/null -w "%{http_code}" -b "$JAR" -X DELETE "${APP1}/api/notebooks/${NEW_NB_ID}")
  [ "$CODE" = "204" ] && pass "DELETE /api/notebooks/{id} → 204" || fail "delete notebook" "$CODE"

  NBS_AFTER=$(curl -s -b "$JAR" "${APP1}/api/notebooks" | jq_count)
  [ "$NBS_AFTER" = "2" ] && pass "삭제 후 노트북 갯수 원복 (2개)" || fail "post-delete" "${NBS_AFTER}"
fi

# ─────────────────────────────────────────────
section "6. Note CRUD + 검색"

if [ -n "${STARTED_ID:-}" ]; then
  NEW_NOTE=$(curl -s -b "$JAR" -X POST "${APP1}/api/notes" \
    -H "Content-Type: application/json" \
    -d "{\"notebookId\":${STARTED_ID},\"title\":\"E2E 테스트 노트\",\"bodyMd\":\"specialword e2e content\"}")
  NOTE_ID=$(echo "$NEW_NOTE" | jq_get id)
  [ -n "$NOTE_ID" ] && pass "POST /api/notes (id=${NOTE_ID})" || fail "create note" "$NEW_NOTE"

  if [ -n "$NOTE_ID" ]; then
    UPDATED=$(curl -s -b "$JAR" -X PATCH "${APP1}/api/notes/${NOTE_ID}" \
      -H "Content-Type: application/json" -d '{"title":"제목 수정","bodyMd":"본문 수정 specialword"}')
    U_TITLE=$(echo "$UPDATED" | jq_get title)
    [ "$U_TITLE" = "제목 수정" ] && pass "PATCH /api/notes/{id} 수정" || fail "update note" "$U_TITLE"

    SEARCH=$(curl -s -b "$JAR" "${APP1}/api/notes/search?q=specialword")
    HITS=$(echo "$SEARCH" | jq_count)
    HAS_OURS=$(echo "$SEARCH" | python3 -c "import json,sys; rs=json.loads(sys.stdin.read()); print(any(r['id']==${NOTE_ID} for r in rs))")
    [ "$HITS" -ge "1" ] && [ "$HAS_OURS" = "True" ] && pass "검색 'specialword' → 우리 노트 포함 ($HITS 건)" || fail "search" "hits=$HITS has=$HAS_OURS"

    CODE=$(curl -s -o /dev/null -w "%{http_code}" -b "$JAR" -X DELETE "${APP1}/api/notes/${NOTE_ID}")
    [ "$CODE" = "204" ] && pass "DELETE /api/notes/{id} → 204" || fail "delete note" "$CODE"
  fi
fi

# ─────────────────────────────────────────────
section "7. 다른 사용자 데이터 접근 차단"

# 존재 안 하는 notebook id → 404 (404 면 정상. 403 도 ok)
RESP=$(curl -s -o /dev/null -w "%{http_code}" -b "$JAR" "${APP1}/api/notes?notebookId=999999")
[[ "$RESP" = "404" || "$RESP" = "403" ]] && pass "존재 안 하는 notebook id → ${RESP}" || fail "isolation" "$RESP"

# ─────────────────────────────────────────────
section "8. 로그아웃 체인 — 세션 폐기"

L1=$(curl -s -o /dev/null -w "%{http_code}|%{redirect_url}" -b "$JAR" "${APP1}/api/logout")
L1_CODE="${L1%%|*}"; L1_LOC="${L1##*|}"
[ "$L1_CODE" = "302" ] && [[ "$L1_LOC" == *"/oauth/logout"* ]] && pass "GET /api/logout → 302 IdP /oauth/logout" || fail "app1 logout" "$L1"

L2_RESP=$(curl -s -i -c "$JAR" -b "$JAR" "$L1_LOC")
L2_CODE=$(echo "$L2_RESP" | head -1 | awk '{print $2}')
L2_LOC=$(echo "$L2_RESP" | grep -i '^location:' | tr -d '\r' | sed 's/^[Ll]ocation: //')
L2_COOKIE=$(echo "$L2_RESP" | grep -i 'set-cookie: idp_session')

[ "$L2_CODE" = "302" ] && pass "IdP /oauth/logout → 302" || fail "idp logout code" "$L2_CODE"
[[ "$L2_COOKIE" == *"Max-Age=0"* ]] && pass "IdP idp_session 쿠키 폐기 (Max-Age=0)" || fail "idp cookie" "$L2_COOKIE"
[[ "$L2_LOC" == "${FRONT}"* ]] && pass "IdP → frontend redirect" || fail "idp redirect" "$L2_LOC"

sleep 1
ME_AFTER=$(curl -s -o /dev/null -w "%{http_code}" -b "$JAR" "${APP1}/api/me")
[ "$ME_AFTER" = "401" ] && pass "로그아웃 후 /api/me → 401" || fail "post-logout me" "$ME_AFTER"

# silent SSO 차단 확인: 이전 cookie 로 /authorize 호출 → 로그인 폼 나와야
RECHECK=$(curl -s -b "$JAR" "${IDP}/oauth/authorize?response_type=code&client_id=app1&redirect_uri=http%3A%2F%2Flocalhost%3A8011%2Fcallback&state=recheck&scope=openid")
echo "$RECHECK" | grep -q 'name="csrf_token"' && pass "로그아웃 후 /authorize → 로그인 폼 (silent SSO 차단)" || fail "silent SSO" "재로그인됨"

# ─────────────────────────────────────────────
echo ""
echo "════════════════════════════════════════════"
echo "총 ${PASS} PASS / ${FAIL} FAIL"
if [ "$FAIL" -gt 0 ]; then
  echo ""
  echo "실패:"
  for t in "${FAILED_TESTS[@]}"; do echo "  - $t"; done
fi
exit $FAIL
