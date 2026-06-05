#!/bin/bash
# 회귀용 통합 테스트 진입점. Go unit + IdP e2e + app1 통합 + 브라우저 E2E 다 실행.
#
# 전제:
#   docker compose up -d 로 11 컨테이너 다 healthy 상태
#   Node.js 18+ (브라우저 E2E 용)

set -e

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PASS_COUNT=0; FAIL_COUNT=0
declare -a FAILED

run() {
  local name="$1"; shift
  echo ""
  echo "════════════════════════════════════════"
  echo "▶ $name"
  echo "════════════════════════════════════════"
  if "$@"; then
    PASS_COUNT=$((PASS_COUNT+1))
  else
    FAIL_COUNT=$((FAIL_COUNT+1))
    FAILED+=("$name")
  fi
}

# 1. Go 단위/통합 테스트
run "Go unit/integration" bash -c "cd '$ROOT/server' && go test ./..."

# 2. IdP HTTP e2e (rate limit 카운터 다음 테스트에 영향 주므로 컨테이너 재시작)
docker restart oauth-idp >/dev/null 2>&1
sleep 6
run "IdP HTTP e2e" bash "$ROOT/scripts/test-idp.sh"

# 3. app1 backend 통합 — 회원가입/세션/CRUD/welcome seed/로그아웃 체인
docker restart oauth-idp >/dev/null 2>&1
sleep 6
run "App1 backend integration" bash "$ROOT/scripts/test-app1.sh"

# 4. 브라우저 E2E (Playwright Chromium headless)
docker restart oauth-idp >/dev/null 2>&1
sleep 6
if [ ! -d "$ROOT/scripts/browser-e2e/node_modules" ]; then
  echo "▶ Playwright 설치 중 (최초 1 회)"
  (cd "$ROOT/scripts/browser-e2e" && npm install >/dev/null 2>&1 && npx playwright install chromium >/dev/null 2>&1)
fi
run "Browser E2E" bash -c "cd '$ROOT/scripts/browser-e2e' && npx playwright test"

# ─────────────────────────────────────────────
echo ""
echo "════════════════════════════════════════"
echo "전체 ${PASS_COUNT} PASS / ${FAIL_COUNT} FAIL"
if [ "$FAIL_COUNT" -gt 0 ]; then
  echo "실패:"
  for f in "${FAILED[@]}"; do echo "  ✗ $f"; done
  exit 1
fi
echo "✅ 모든 테스트 통과"
