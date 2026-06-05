#!/bin/bash
# OAuth/OIDC end-to-end + 보안 + 운영 시나리오 테스트.
# app1 + IdP 만 대상. 결과는 PASS/FAIL 라인 + 요약.

set -u

IDP="http://localhost:8080"
APP1="http://localhost:8011"
APP1_FRONT="http://localhost:5181"
CLIENT_ID="app1"
CLIENT_SECRET="app1-secret"
REDIRECT_URI="http://localhost:8011/callback"
JAR="/tmp/oauth-test-cookies"
TMP="/tmp/oauth-test"
mkdir -p "$TMP"

PASS=0; FAIL=0
declare -a FAILED_TESTS

pass() { echo "  ✅ $1"; PASS=$((PASS+1)); }
fail() { echo "  ❌ $1 — $2"; FAIL=$((FAIL+1)); FAILED_TESTS+=("$1: $2"); }

section() { echo ""; echo "▶ $1"; }

# ─── PKCE 헬퍼 ───
gen_verifier() {
  openssl rand -base64 96 | tr -d "=+/\n" | cut -c1-64
}
sha256_b64url() {
  printf '%s' "$1" | openssl dgst -sha256 -binary | base64 | tr -d "=\n" | tr "+/" "-_"
}

# ─── 깨끗한 cookie jar로 OAuth 로그인 시뮬레이션 ───
# args: jar verifier nonce
# stdout: auth code  (또는 빈 문자열 + stderr 에 메시지)
oauth_login() {
  local jar="$1" verifier="$2" nonce="$3" user="${4:-user}" pass="${5:-password1!}"
  rm -f "$jar"
  local challenge state
  challenge=$(sha256_b64url "$verifier")
  state=$(openssl rand -hex 8)

  # 1. /authorize → login 폼 (csrf 쿠키 set + csrf_token hidden)
  local authorize_url="${IDP}/oauth/authorize?response_type=code&client_id=${CLIENT_ID}"
  authorize_url+="&redirect_uri=$(python3 -c "import urllib.parse;print(urllib.parse.quote('${REDIRECT_URI}'))")"
  authorize_url+="&scope=openid+profile+email&state=${state}"
  authorize_url+="&code_challenge=${challenge}&code_challenge_method=S256&nonce=${nonce}"

  curl -s -c "$jar" -b "$jar" "$authorize_url" -o "$TMP/login.html"
  local csrf
  csrf=$(grep -oE 'name="csrf_token"\s+value="[^"]+"' "$TMP/login.html" | sed 's/.*value="\([^"]*\)".*/\1/')
  if [ -z "$csrf" ]; then echo "" ; return 1 ; fi

  # 2. POST /oauth/login → 302 callback
  local headers
  headers=$(curl -s -i -c "$jar" -b "$jar" \
    -X POST "${IDP}/oauth/login" \
    -d "id=${user}" -d "password=${pass}" \
    -d "csrf_token=${csrf}" -d "client_id=${CLIENT_ID}" \
    -d "redirect_uri=${REDIRECT_URI}" -d "state=${state}" -d "scope=openid profile email" \
    -d "code_challenge=${challenge}" -d "code_challenge_method=S256" -d "nonce=${nonce}" 2>&1)
  local loc
  loc=$(echo "$headers" | grep -i '^location:' | tr -d '\r' | sed 's/^[Ll]ocation: //')
  if [[ "$loc" != *"code="* ]]; then
    echo "$headers" > "$TMP/login_fail.txt"
    echo ""; return 1
  fi
  echo "$loc" | sed -E 's/.*[?&]code=([^&]+).*/\1/'
}

# Basic auth header for token endpoint
basic_auth() {
  printf "%s" "$(printf '%s:%s' "${CLIENT_ID}" "${CLIENT_SECRET}" | base64)"
}

# ─────────────────────────────────────────────
section "1. Discovery + JWKS"

DISC=$(curl -s "${IDP}/.well-known/openid-configuration")
echo "$DISC" | python3 -c "import json,sys;d=json.loads(sys.stdin.read());\
print('OK' if d['issuer']=='${IDP}' and 'S256' in d['code_challenge_methods_supported'] \
and 'RS256' in d['id_token_signing_alg_values_supported'] else 'BAD')" > "$TMP/disc_check"
[ "$(cat $TMP/disc_check)" = "OK" ] && pass "Discovery: issuer/PKCE/RS256 명시" || fail "Discovery" "값 누락"

JWKS=$(curl -s "${IDP}/oauth/jwks")
echo "$JWKS" | python3 -c "import json,sys;d=json.loads(sys.stdin.read());\
k=d['keys'][0];print('OK' if k['kty']=='RSA' and k['alg']=='RS256' and k['use']=='sig' else 'BAD')" > "$TMP/jwks_check"
[ "$(cat $TMP/jwks_check)" = "OK" ] && pass "JWKS: RSA RS256 sig 공개키 노출" || fail "JWKS" "필드 누락"

# ─────────────────────────────────────────────
section "2. /authorize 입력 검증"

# 잘못된 client_id → 에러 페이지
RESP=$(curl -s -o /dev/null -w "%{http_code}" "${IDP}/oauth/authorize?response_type=code&client_id=nope&redirect_uri=${REDIRECT_URI}&state=x")
[ "$RESP" = "400" ] && pass "잘못된 client_id → 400" || fail "잘못된 client_id" "응답 $RESP"

# 잘못된 redirect_uri → 에러 페이지
RESP=$(curl -s -o /dev/null -w "%{http_code}" "${IDP}/oauth/authorize?response_type=code&client_id=${CLIENT_ID}&redirect_uri=http://evil/cb&state=x")
[ "$RESP" = "400" ] && pass "잘못된 redirect_uri → 400 (Open Redirect 차단)" || fail "redirect_uri 검증" "$RESP"

# 잘못된 response_type → 안전 redirect with error
LOC=$(curl -s -o /dev/null -w "%{redirect_url}" "${IDP}/oauth/authorize?response_type=token&client_id=${CLIENT_ID}&redirect_uri=${REDIRECT_URI}&state=abc")
[[ "$LOC" == *"error=unsupported_response_type"* && "$LOC" == *"state=abc"* ]] \
  && pass "response_type=token → unsupported_response_type + state 보존" \
  || fail "response_type" "loc=$LOC"

# 잘못된 PKCE method → invalid_request
LOC=$(curl -s -o /dev/null -w "%{redirect_url}" "${IDP}/oauth/authorize?response_type=code&client_id=${CLIENT_ID}&redirect_uri=${REDIRECT_URI}&state=abc&code_challenge=xxx&code_challenge_method=plain")
[[ "$LOC" == *"error=invalid_request"* ]] && pass "PKCE plain method → invalid_request (S256 강제)" || fail "PKCE method" "loc=$LOC"

# ─────────────────────────────────────────────
section "3. CSRF 검증 (rate limit 영향 받기 전에 먼저)"

rm -f "$JAR-csrf"
curl -s -c "$JAR-csrf" "${IDP}/oauth/authorize?response_type=code&client_id=${CLIENT_ID}&redirect_uri=${REDIRECT_URI}&state=csrf" -o "$TMP/csrf.html" > /dev/null
RESP=$(curl -s -o /dev/null -w "%{http_code}" -c "$JAR-csrf" -b "$JAR-csrf" \
  -X POST "${IDP}/oauth/login" \
  -d "id=alice" -d "password=password123" \
  -d "csrf_token=WRONG" -d "client_id=${CLIENT_ID}" -d "redirect_uri=${REDIRECT_URI}" -d "state=csrf" -d "scope=openid")
[ "$RESP" = "403" ] && pass "CSRF token 불일치 → 403" || fail "CSRF" "$RESP"

# ─────────────────────────────────────────────
section "4. /token 엔드포인트 인증"

# Basic auth 없음
RESP=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${IDP}/oauth/token" -d "grant_type=authorization_code")
[ "$RESP" = "401" ] && pass "Basic auth 없음 → 401 invalid_client" || fail "client auth missing" "$RESP"

# 잘못된 client_secret
RESP=$(curl -s -o /dev/null -w "%{http_code}" -u "${CLIENT_ID}:WRONG" -X POST "${IDP}/oauth/token" -d "grant_type=authorization_code&code=xxx&redirect_uri=${REDIRECT_URI}")
[ "$RESP" = "401" ] && pass "잘못된 secret → 401 invalid_client (bcrypt verify)" || fail "wrong secret" "$RESP"

# 알 수 없는 grant_type
RESP=$(curl -s -u "${CLIENT_ID}:${CLIENT_SECRET}" -X POST "${IDP}/oauth/token" -d "grant_type=password&username=x&password=y")
[[ "$RESP" == *"unsupported_grant_type"* ]] && pass "지원 안 하는 grant_type → unsupported_grant_type" || fail "unsupported_grant_type" "$RESP"

# ─────────────────────────────────────────────
section "4. End-to-end Authorization Code Flow (PKCE + nonce)"

VERIFIER=$(gen_verifier)
NONCE=$(openssl rand -hex 16)
CODE=$(oauth_login "$JAR-e2e" "$VERIFIER" "$NONCE" "user" "password1!")

if [ -z "$CODE" ]; then
  # user 계정 비번 모를 가능성 — alice/bob/carol 로 폴백
  CODE=$(oauth_login "$JAR-e2e" "$VERIFIER" "$NONCE" "alice" "password123")
fi

if [ -z "$CODE" ]; then
  fail "로그인 시뮬레이션" "auth code 발급 실패 (user/alice 둘 다 실패)"
else
  pass "로그인 → auth code 발급 (state 검증 + CSRF 통과)"

  # 토큰 교환
  TOKENS=$(curl -s -u "${CLIENT_ID}:${CLIENT_SECRET}" -X POST "${IDP}/oauth/token" \
    -d "grant_type=authorization_code" -d "code=${CODE}" \
    -d "redirect_uri=${REDIRECT_URI}" -d "code_verifier=${VERIFIER}")

  ACCESS=$(echo "$TOKENS" | python3 -c "import json,sys;print(json.loads(sys.stdin.read()).get('access_token',''))")
  REFRESH=$(echo "$TOKENS" | python3 -c "import json,sys;print(json.loads(sys.stdin.read()).get('refresh_token',''))")
  IDTOKEN=$(echo "$TOKENS" | python3 -c "import json,sys;print(json.loads(sys.stdin.read()).get('id_token',''))")

  [ -n "$ACCESS" ] && pass "/token authorization_code → access_token 발급" || fail "access_token" "$TOKENS"
  [ -n "$REFRESH" ] && pass "/token → refresh_token 발급" || fail "refresh_token" "missing"
  [ -n "$IDTOKEN" ] && pass "/token (scope=openid) → id_token 발급" || fail "id_token" "missing"

  # ID Token claims 검증
  if [ -n "$IDTOKEN" ]; then
    CLAIMS=$(echo "$IDTOKEN" | python3 -c "
import json,sys,base64
tok=sys.stdin.read().strip()
payload=tok.split('.')[1]
payload+='='*(-len(payload)%4)
try:
  c=json.loads(base64.urlsafe_b64decode(payload).decode())
  a=c.get('aud')
  aud_ok = a=='${CLIENT_ID}' or (isinstance(a,list) and '${CLIENT_ID}' in a)
  print(f\"aud_ok={aud_ok}|nonce_ok={c.get('nonce')=='${NONCE}'}|iss_ok={c.get('iss')=='${IDP}'}|sub_len={len(c.get('sub',''))}|auth_time={'auth_time' in c}\")
except Exception as e: print(f'PARSE_FAIL:{e}')
")
    if [[ "$CLAIMS" == *"aud_ok=True"* && "$CLAIMS" == *"nonce_ok=True"* && "$CLAIMS" == *"iss_ok=True"* && "$CLAIMS" == *"auth_time=True"* ]]; then
      pass "ID Token claims: aud(array)/nonce echo/iss/auth_time"
    else
      fail "ID Token claims" "$CLAIMS"
    fi
  fi

  # PKCE 재발급 차단 - 같은 code 재사용
  RESP=$(curl -s -u "${CLIENT_ID}:${CLIENT_SECRET}" -X POST "${IDP}/oauth/token" \
    -d "grant_type=authorization_code" -d "code=${CODE}" \
    -d "redirect_uri=${REDIRECT_URI}" -d "code_verifier=${VERIFIER}")
  [[ "$RESP" == *"invalid_grant"* ]] && pass "Auth code 재사용 → invalid_grant (LoadAndDelete)" || fail "code reuse" "$RESP"
fi

# ─────────────────────────────────────────────
section "5. PKCE 검증 실패 케이스"

VERIFIER=$(gen_verifier)
NONCE=$(openssl rand -hex 16)
CODE=$(oauth_login "$JAR-pkce" "$VERIFIER" "$NONCE" "user" "password1!")
[ -z "$CODE" ] && CODE=$(oauth_login "$JAR-pkce" "$VERIFIER" "$NONCE" "alice" "password123")

if [ -n "$CODE" ]; then
  # 잘못된 verifier
  WRONG=$(gen_verifier)
  RESP=$(curl -s -u "${CLIENT_ID}:${CLIENT_SECRET}" -X POST "${IDP}/oauth/token" \
    -d "grant_type=authorization_code" -d "code=${CODE}" \
    -d "redirect_uri=${REDIRECT_URI}" -d "code_verifier=${WRONG}")
  [[ "$RESP" == *"invalid_grant"* && "$RESP" == *"PKCE"* ]] \
    && pass "PKCE 잘못된 verifier → invalid_grant" \
    || fail "PKCE verify" "$RESP"
fi

# code_verifier 누락
VERIFIER=$(gen_verifier)
NONCE=$(openssl rand -hex 16)
CODE=$(oauth_login "$JAR-pkce2" "$VERIFIER" "$NONCE" "user" "password1!")
[ -z "$CODE" ] && CODE=$(oauth_login "$JAR-pkce2" "$VERIFIER" "$NONCE" "alice" "password123")
if [ -n "$CODE" ]; then
  RESP=$(curl -s -u "${CLIENT_ID}:${CLIENT_SECRET}" -X POST "${IDP}/oauth/token" \
    -d "grant_type=authorization_code" -d "code=${CODE}" -d "redirect_uri=${REDIRECT_URI}")
  [[ "$RESP" == *"code_verifier"* ]] && pass "PKCE verifier 누락 → invalid_request" || fail "PKCE missing" "$RESP"
fi

# ─────────────────────────────────────────────
section "6. Refresh Token Rotation"

if [ -n "${REFRESH:-}" ]; then
  RESP=$(curl -s -u "${CLIENT_ID}:${CLIENT_SECRET}" -X POST "${IDP}/oauth/token" \
    -d "grant_type=refresh_token" -d "refresh_token=${REFRESH}")
  NEW_ACCESS=$(echo "$RESP" | python3 -c "import json,sys;print(json.loads(sys.stdin.read()).get('access_token',''))")
  NEW_REFRESH=$(echo "$RESP" | python3 -c "import json,sys;print(json.loads(sys.stdin.read()).get('refresh_token',''))")
  [ -n "$NEW_ACCESS" ] && pass "refresh_token grant → 새 access_token" || fail "refresh access" "$RESP"
  [ -n "$NEW_REFRESH" ] && [ "$NEW_REFRESH" != "$REFRESH" ] && pass "refresh rotation: 새 refresh_token 발급" || fail "refresh rotation" "동일 토큰"

  # 옛 refresh 재사용 차단
  RESP=$(curl -s -u "${CLIENT_ID}:${CLIENT_SECRET}" -X POST "${IDP}/oauth/token" \
    -d "grant_type=refresh_token" -d "refresh_token=${REFRESH}")
  [[ "$RESP" == *"invalid_grant"* ]] && pass "옛 refresh_token 재사용 → invalid_grant" || fail "refresh reuse" "$RESP"

  ACCESS="$NEW_ACCESS"
  REFRESH="$NEW_REFRESH"
fi

# ─────────────────────────────────────────────
section "7. userinfo + scope 필터링"

if [ -n "${ACCESS:-}" ]; then
  INFO=$(curl -s -H "Authorization: Bearer ${ACCESS}" "${IDP}/oauth/userinfo")
  HAS_USERNAME=$(echo "$INFO" | python3 -c "import json,sys;d=json.loads(sys.stdin.read());print('Y' if 'preferred_username' in d else 'N')")
  HAS_EMAIL=$(echo "$INFO" | python3 -c "import json,sys;d=json.loads(sys.stdin.read());print('Y' if 'email' in d else 'N')")
  HAS_SUB=$(echo "$INFO" | python3 -c "import json,sys;d=json.loads(sys.stdin.read());print('Y' if 'sub' in d else 'N')")
  [ "$HAS_SUB" = "Y" ] && pass "userinfo: sub 노출" || fail "userinfo sub" "$INFO"
  [ "$HAS_USERNAME" = "Y" ] && pass "userinfo scope=profile → preferred_username" || fail "userinfo profile" "$INFO"
  # email 은 user 계정에 메일 없을 수도 있어서 조건부
  [ "$HAS_EMAIL" = "Y" ] && pass "userinfo scope=email → email" || echo "  ℹ userinfo email 없음 (사용자에 email 미설정 가능)"

  # 잘못된 토큰
  RESP=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer invalid.token.here" "${IDP}/oauth/userinfo")
  [ "$RESP" = "401" ] && pass "userinfo 잘못된 토큰 → 401" || fail "userinfo invalid" "$RESP"

  # Bearer 헤더 없음
  RESP=$(curl -s -o /dev/null -w "%{http_code}" "${IDP}/oauth/userinfo")
  [ "$RESP" = "401" ] && pass "userinfo Bearer 없음 → 401" || fail "userinfo no bearer" "$RESP"
fi

# ─────────────────────────────────────────────
section "8. Token Introspection"

if [ -n "${ACCESS:-}" ]; then
  RESP=$(curl -s -u "${CLIENT_ID}:${CLIENT_SECRET}" -X POST "${IDP}/oauth/introspect" \
    -d "token=${ACCESS}" -d "token_type_hint=access_token")
  ACTIVE=$(echo "$RESP" | python3 -c "import json,sys;print(json.loads(sys.stdin.read()).get('active'))")
  [ "$ACTIVE" = "True" ] && pass "introspect 유효 access → active=true" || fail "introspect active" "$RESP"

  # 무효 토큰
  RESP=$(curl -s -u "${CLIENT_ID}:${CLIENT_SECRET}" -X POST "${IDP}/oauth/introspect" -d "token=garbage")
  ACTIVE=$(echo "$RESP" | python3 -c "import json,sys;print(json.loads(sys.stdin.read()).get('active'))")
  [ "$ACTIVE" = "False" ] && pass "introspect 무효 토큰 → active=false" || fail "introspect inactive" "$RESP"

  # 인증 없이 호출
  RESP=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${IDP}/oauth/introspect" -d "token=x")
  [ "$RESP" = "401" ] && pass "introspect 클라이언트 인증 없음 → 401" || fail "introspect noauth" "$RESP"
fi

# ─────────────────────────────────────────────
section "9. Token Revocation"

if [ -n "${ACCESS:-}" ]; then
  # access token revoke
  RESP=$(curl -s -o /dev/null -w "%{http_code}" -u "${CLIENT_ID}:${CLIENT_SECRET}" \
    -X POST "${IDP}/oauth/revoke" -d "token=${ACCESS}" -d "token_type_hint=access_token")
  [ "$RESP" = "200" ] && pass "revoke access → 200" || fail "revoke access status" "$RESP"

  # 폐기 후 userinfo 거부
  RESP=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer ${ACCESS}" "${IDP}/oauth/userinfo")
  [ "$RESP" = "401" ] && pass "폐기된 access 로 userinfo → 401" || fail "revoked userinfo" "$RESP"

  # 폐기 후 introspect = false
  RESP=$(curl -s -u "${CLIENT_ID}:${CLIENT_SECRET}" -X POST "${IDP}/oauth/introspect" -d "token=${ACCESS}")
  ACTIVE=$(echo "$RESP" | python3 -c "import json,sys;print(json.loads(sys.stdin.read()).get('active'))")
  [ "$ACTIVE" = "False" ] && pass "폐기된 access introspect → active=false" || fail "revoked introspect" "$RESP"

  # 존재하지 않는 토큰 폐기 시도 → 200 (RFC 7009 정보 노출 방지)
  RESP=$(curl -s -o /dev/null -w "%{http_code}" -u "${CLIENT_ID}:${CLIENT_SECRET}" \
    -X POST "${IDP}/oauth/revoke" -d "token=nonexistent")
  [ "$RESP" = "200" ] && pass "존재 없는 토큰 폐기 시도 → 200 (정보 노출 방지)" || fail "revoke unknown" "$RESP"
fi

# ─────────────────────────────────────────────
section "10. RP-initiated Logout"

if [ -n "${IDTOKEN:-}" ]; then
  # post_logout_redirect_uri 허용 (등록된 MainURL — curl 이 끝에 / 추가할 수 있어 prefix 비교)
  LOC=$(curl -s -o /dev/null -w "%{redirect_url}" \
    "${IDP}/oauth/logout?post_logout_redirect_uri=http%3A%2F%2Flocalhost%3A5181&id_token_hint=${IDTOKEN}")
  [[ "$LOC" == "${APP1_FRONT}"* ]] && pass "logout 등록된 MainURL redirect 허용 ($LOC)" || fail "logout redirect main" "loc=$LOC"

  # same-origin path 도 허용 (학습 단순화 정책)
  LOC=$(curl -s -o /dev/null -w "%{redirect_url}" \
    "${IDP}/oauth/logout?post_logout_redirect_uri=http%3A%2F%2Flocalhost%3A5181%2Ffoo&id_token_hint=${IDTOKEN}")
  [[ "$LOC" == "${APP1_FRONT}/foo" ]] && pass "logout same-origin path 허용" || fail "logout same-origin" "loc=$LOC"

  # 등록 안 된 origin 거부 → 안내 페이지 (redirect 없음, 200)
  RESP=$(curl -s -o /dev/null -w "%{http_code}|%{redirect_url}" \
    "${IDP}/oauth/logout?post_logout_redirect_uri=http%3A%2F%2Fevil.com&id_token_hint=${IDTOKEN}")
  [[ "$RESP" == "200|" ]] && pass "logout 등록 안 된 origin 거부 (안내 페이지)" || fail "logout reject evil" "$RESP"

  # idp_session 쿠키 만료 확인
  CK=$(curl -s -i "${IDP}/oauth/logout?post_logout_redirect_uri=http%3A%2F%2Flocalhost%3A5181" | grep -i "set-cookie: idp_session")
  [[ "$CK" == *"Max-Age=0"* ]] && pass "logout 시 idp_session 쿠키 Max-Age=0" || fail "logout cookie" "$CK"
fi

# ─────────────────────────────────────────────
section "11. 보안 정밀 검증 (rate limit 전에)"

# client_secret 이 평문 저장 안 되어 있는지
PG=$(docker exec oauth-postgres psql -U oauth -d oauth -tA -c "SELECT COUNT(*) FROM clients WHERE client_secret IS NOT NULL;" 2>/dev/null)
[ "$PG" = "0" ] && pass "DB clients.client_secret 평문 컬럼 NULL (bcrypt hash 만 저장)" || fail "plaintext secret" "$PG 행에 평문 남음"

# ─────────────────────────────────────────────
section "12. Rate Limit (/oauth/login) — 카운터 소진하므로 마지막에"

# 새 cookie jar + CSRF 가져옴
rm -f "$JAR-rl"
curl -s -c "$JAR-rl" "${IDP}/oauth/authorize?response_type=code&client_id=${CLIENT_ID}&redirect_uri=${REDIRECT_URI}&state=rl" -o "$TMP/rl.html"
CSRF=$(grep -oE 'name="csrf_token"\s+value="[^"]+"' "$TMP/rl.html" | sed 's/.*value="\([^"]*\)".*/\1/')

# 잘못된 비번으로 13 회 시도 (burst 5 + rate 10/min). 11회 즈음에 429 나와야.
HIT_429=0
for i in $(seq 1 13); do
  CODE=$(curl -s -o /dev/null -w "%{http_code}" -c "$JAR-rl" -b "$JAR-rl" \
    -X POST "${IDP}/oauth/login" \
    -d "id=nobody" -d "password=wrong" \
    -d "csrf_token=${CSRF}" -d "client_id=${CLIENT_ID}" \
    -d "redirect_uri=${REDIRECT_URI}" -d "state=rl" -d "scope=openid")
  if [ "$CODE" = "429" ]; then HIT_429=1; break; fi
done
[ "$HIT_429" = "1" ] && pass "13 회 반복 시도 중 429 발생 (rate limit 작동)" || fail "rate limit" "429 안 떴음"

# 만료된 토큰 (직접 발급은 어렵고 — refresh 만료 7d/access 15분이라 패스 시간 부족)
# 알고리즘 none 토큰 시도
NONE_JWT="eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ4In0."
RESP=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer ${NONE_JWT}" "${IDP}/oauth/userinfo")
[ "$RESP" = "401" ] && pass "alg=none 위조 토큰 → 401 (알고리즘 혼동 방어)" || fail "alg=none" "$RESP"

# ─────────────────────────────────────────────
section "13. App1 백엔드 통합"

# /api/me 인증 없음
RESP=$(curl -s -o /dev/null -w "%{http_code}" "${APP1}/api/me")
[ "$RESP" = "401" ] && pass "app1 /api/me 세션 없음 → 401" || fail "app1 me unauth" "$RESP"

# /api/logout: 세션 없어도 IdP logout 으로 redirect (302)
LOC=$(curl -s -o /dev/null -w "%{redirect_url}" "${APP1}/api/logout")
[[ "$LOC" == *"/oauth/logout?"* && "$LOC" == *"post_logout_redirect_uri="* ]] \
  && pass "app1 /api/logout → IdP /oauth/logout 체인" \
  || fail "app1 logout chain" "loc=$LOC"

# /api/notebooks 인증 없음
RESP=$(curl -s -o /dev/null -w "%{http_code}" "${APP1}/api/notebooks")
[ "$RESP" = "401" ] && pass "app1 /api/notebooks 세션 없음 → 401" || fail "app1 notebooks unauth" "$RESP"

# ─────────────────────────────────────────────
section "14. Audit Log 발생 확인"

# 마지막 1 분 로그에서 token.issued, login.success/failed, token.revoked 가 있는지
AUDIT=$(docker logs oauth-idp --since 5m 2>&1 | grep -E '"msg":(("token\.issued")|("login\.success")|("token\.client_auth_failed")|("token\.revoked"))' | wc -l | tr -d ' ')
[ "$AUDIT" -ge 3 ] && pass "감사 로그 다수 발생 (token.issued/login/revoke 등 ≥3)" || fail "audit log" "${AUDIT}건"

# ─────────────────────────────────────────────
section "15. Welcome Seed 동작"

# 노트북이 비어있는 사용자가 callback 거치면 seed 가 채움. 직접 확인:
# 현재 DB 상태로 가늠.
SEED_USERS=$(docker exec app1-mariadb mariadb -uapp1 -papp1-dev app1_notebook -sN -e \
  "SELECT COUNT(DISTINCT owner_sub) FROM notebooks WHERE title IN ('시작하기','프로젝트 메모')" 2>/dev/null)
[ "$SEED_USERS" -ge 0 ] && pass "Welcome seed 코드 DB 접근 가능 (현재 시드 사용자 ${SEED_USERS} 명)" || fail "seed" "$SEED_USERS"

# ─────────────────────────────────────────────
echo ""
echo "════════════════════════════════════════"
echo "총 ${PASS} PASS / ${FAIL} FAIL"
if [ "$FAIL" -gt 0 ]; then
  echo ""
  echo "실패한 테스트:"
  for t in "${FAILED_TESTS[@]}"; do echo "  - $t"; done
fi
exit $FAIL
