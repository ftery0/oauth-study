#!/bin/bash
# 한 번 실행 = start.sh 가 띄운 모든 컨테이너 종료. 데이터는 볼륨에 그대로.
# 사용:           ./stop.sh
# 데이터 초기화:  docker compose down -v

set -e

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if ! docker info >/dev/null 2>&1; then
    echo "✗ Docker 데몬이 꺼져 있어. 이미 stop 된 것과 같아."
    exit 0
fi

echo "▶ docker compose down"
docker compose -f "$ROOT/docker-compose.yml" down

echo "▶ 데이터까지 지우려면:  docker compose down -v"
