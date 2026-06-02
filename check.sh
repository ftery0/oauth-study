#!/bin/bash
# 살아있는 포트 한 줄 점검.

echo "▶ 포트 상태"
for entry in "5433:postgres" "8080:idp" "8011:app1-backend" "5181:app1-frontend" "8012:app2-backend" "5182:app2-frontend" "8013:app3-backend" "5183:app3-frontend"; do
    port="${entry%%:*}"
    name="${entry#*:}"
    if lsof -nP -iTCP:$port -sTCP:LISTEN > /dev/null 2>&1; then
        printf "  %-18s :%-5s  ✓\n" "$name" "$port"
    else
        printf "  %-18s :%-5s  ✗\n" "$name" "$port"
    fi
done
