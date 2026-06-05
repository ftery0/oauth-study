# oauth-study

같은 IdP 를 쓰는 **3 개의 풀스택 서비스**로 SSO 가 실제로 어떻게 동작하는지 직접 띄워볼 수 있는 학습 프로젝트.

| 앱 | 역할 | 스택 | DB | URL |
|---|---|---|---|---|
| **Notebook** | 노트 / 위키 | Spring Boot + React (TS) | MariaDB | http://localhost:5181 |
| **TaskBoard** | 칸반 보드 | Node Express + Vue 3 (TS) | MongoDB | http://localhost:5182 |
| **HelpDesk** | 고객 지원 티켓 | FastAPI + Next.js (TS) | PostgreSQL | http://localhost:5183 |
| IdP / OAuth | 인가 서버 | Go | Postgres | http://localhost:8080 |

---

## 빠른 시작

Docker / Docker Compose v2 만 있으면 됩니다.

```bash
docker compose up -d --build
```

총 11 개 컨테이너 (IdP + 4 DB + 3 backend + 3 frontend) 가 한 번에 뜹니다. 첫 빌드는 의존성 다운로드 때문에 5 ~ 10 분.

상태 확인:
```bash
docker compose ps
```

데이터 초기화:
```bash
docker compose down -v
```

---

## 흐름

1. http://localhost:5181 (Notebook) 접속 → OAuth 로그인 화면으로 자동 이동
2. 처음이면 회원가입, 있으면 로그인
3. 다른 앱(http://localhost:5182, http://localhost:5183) 으로 가도 같은 계정으로 **자동 로그인** (silent SSO)
4. 각 앱은 자체 DB 를 가지고 있어 IdP `sub` 으로 user 행이 자동 생성되고 데이터는 격리됨

---

## 디렉토리

```
.
├── server/                IdP / OAuth (Go)
├── examples/
│   ├── app1/              Notebook  (Spring + React)
│   ├── app2/              TaskBoard (Node + Vue)
│   └── app3/              HelpDesk  (FastAPI + Next)
└── docker-compose.yml
```

---

## 호스트에서 직접 띄우기

DB 만 컨테이너로 쓰고 각 앱은 호스트에서 직접 실행하고 싶다면:

```bash
docker compose up -d postgres app1-mariadb app2-mongo app3-postgres
```

이후 각 앱 디렉토리에서:

| 앱 | 명령 |
|---|---|
| IdP | `cd server && go run .` |
| Notebook backend | `cd examples/app1/backend && mvn spring-boot:run` |
| Notebook frontend | `cd examples/app1/frontend && npm install && npm run dev` |
| TaskBoard backend | `cd examples/app2/backend && npm install && npm run dev` |
| TaskBoard frontend | `cd examples/app2/frontend && npm install && npm run dev` |
| HelpDesk backend | `cd examples/app3/backend && python3 -m venv .venv && .venv/bin/pip install -r requirements.txt && .venv/bin/python main.py` |
| HelpDesk frontend | `cd examples/app3/frontend && npm install && npm run dev` |
