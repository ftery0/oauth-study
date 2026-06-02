#!/bin/bash
# 한 번 실행 = start.sh 가 띄운 모든 서버 종료.
# 사용:   ./stop.sh
# Postgres 도 끄려면:  docker compose down

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="$ROOT/logs"

echo "▶ 종료 중..."

# 1. start.sh 가 남긴 pid 파일로 먼저 정리
if [ -d "$LOG_DIR" ]; then
    for pidfile in "$LOG_DIR"/*.pid; do
        [ -f "$pidfile" ] || continue
        name=$(basename "$pidfile" .pid)
        pid=$(cat "$pidfile")
        if kill "$pid" 2>/dev/null; then
            echo "  $name (pid $pid) ✓"
        fi
        rm -f "$pidfile"
    done
fi

# 2. 자식 프로세스(mvn → java, npm → vite/next 등) 까지 포트로 정리
sleep 0.5
for P in 8080 8011 5181 8012 5182 8013 5183; do
    pid=$(lsof -nP -iTCP:$P -sTCP:LISTEN -t 2>/dev/null)
    if [ -n "$pid" ]; then
        kill $pid 2>/dev/null && echo "  port $P (pid $pid) ✓"
    fi
done

echo "▶ Postgres 도 끄려면:  docker compose down"
