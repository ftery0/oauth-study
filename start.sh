#!/bin/bash
# 한 번 실행 = 모든 서버 백그라운드 기동.
# 사용:   ./start.sh
# 종료:   ./stop.sh
# 로그:   tail -f logs/*.log

set -e

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="$ROOT/logs"
mkdir -p "$LOG_DIR"

# Spring Boot 용 Java PATH (brew openjdk@21)
export JAVA_HOME="/opt/homebrew/opt/openjdk@21/libexec/openjdk.jdk/Contents/Home"
export PATH="/opt/homebrew/opt/openjdk@21/bin:$PATH"

# ───────── Postgres ─────────
echo "▶ Postgres (docker compose)"
docker compose -f "$ROOT/docker-compose.yml" up -d > "$LOG_DIR/postgres.log" 2>&1

# 헬퍼: 디렉토리 + 명령 → 백그라운드 + 로그 파일
run() {
    local name="$1"
    local dir="$2"
    shift 2
    (
        cd "$dir"
        nohup "$@" > "$LOG_DIR/$name.log" 2>&1 &
        echo $! > "$LOG_DIR/$name.pid"
    )
    echo "  → $name (pid $(cat "$LOG_DIR/$name.pid"))   log: logs/$name.log"
}

echo "▶ IdP / 앱 백그라운드 기동"
run idp           "$ROOT/server"                  go run .
run app1-backend  "$ROOT/examples/app1/backend"   mvn -q spring-boot:run
run app1-frontend "$ROOT/examples/app1/frontend"  npm run dev
run app2-backend  "$ROOT/examples/app2/backend"   node index.js
run app2-frontend "$ROOT/examples/app2/frontend"  npm run dev
run app3-backend  "$ROOT/examples/app3/backend"   .venv/bin/python main.py
run app3-frontend "$ROOT/examples/app3/frontend"  npm run dev

cat <<EOF

▶ 부팅 중 — Spring Boot 는 30초~1분 걸려요.
▶ 진행 상황:  tail -f $LOG_DIR/*.log
▶ 한 줄 점검:  ./check.sh    (선택, 살아있는 포트 확인)
▶ 종료:        ./stop.sh

▶ 접속
   어드민:  http://localhost:8080/admin/login   (비밀번호: admin)
   app1 :  http://localhost:5181     (Spring + React, group-a)
   app2 :  http://localhost:5182     (Node + Vue,    group-a)
   app3 :  http://localhost:5183     (Python + Next, group-b)

EOF
