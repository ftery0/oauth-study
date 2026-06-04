#!/bin/bash
# 한 번 실행 = 모든 서비스(11 컨테이너) 백그라운드 기동.
# 사용:   ./start.sh
# 종료:   ./stop.sh
# 로그:   docker compose logs -f [서비스명]

set -e

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if ! docker info >/dev/null 2>&1; then
    echo "✗ Docker 데몬이 꺼져 있어. Docker Desktop 켜고 다시 실행."
    exit 1
fi

echo "▶ docker compose up -d --build"
docker compose -f "$ROOT/docker-compose.yml" up -d --build

cat <<EOF

▶ 부팅 중 — 첫 빌드면 5~10 분, 이후엔 30 초 안.
▶ 상태:   docker compose ps
▶ 로그:   docker compose logs -f                (전체)
          docker compose logs -f oauth-idp      (특정 서비스)
▶ 종료:   ./stop.sh                             (데이터 보존)
          docker compose down -v                (데이터까지 초기화)

▶ 접속
   어드민:  http://localhost:8080/admin/login   (비밀번호: admin)
   app1 :  http://localhost:5181     (Notebook,  Spring + React)
   app2 :  http://localhost:5182     (TaskBoard, Node + Vue)
   app3 :  http://localhost:5183     (HelpDesk,  FastAPI + Next)

EOF
