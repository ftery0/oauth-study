# oauth-study — oAuth 스타일 SSO 학습 프로젝트


## 무엇이 들어 있나

```
oauth-study/
├── server/                     IdP / DAuth 서버 (Go)
│   ├── handlers/               authorize, login, token, userinfo, JWKS, CSRF, redirect helper
│   ├── policy/                 정책 Resolver (그룹 단위 SSO 경계, 진리표 테스트)
│   ├── store/                  client / group / IdP session (인메모리)
│   └── models/                 ProjectGroup, Client, IdPSession, User
│
├── sdk/                        Go 클라이언트 SDK (다음 phase 에서 정리 예정)
│
├── examples/                   같은 IdP 를 쓰는 클라이언트 앱 3 개
│   ├── app1/                   group-a · Spring Boot + React  · :8011/:5181 · blue
│   ├── app2/                   group-a · Node Express + Vue   · :8012/:5182 · emerald
│   └── app3/                   group-b · FastAPI + Next.js    · :8013/:5183 · amber
│
└── docs/                       설계/계획 문서
    ├── master-plan.md          전체 로드맵 (Phase 1/2/3)
    ├── oauth-flow.md           기본 OAuth 흐름
    └── sso-phase1/             Phase 1 명세/설계/워크플로우/검토
```

---

## 핵심 모델 한 줄

```
같은 그룹 안 → silent SSO (폼 없이 자동 로그인)
다른 그룹 가면 → 로그인 폼 다시 (Realm 경계)
```

| 그룹 | SSO 기본 | 멤버 |
|------|:---:|------|
| group-a | ON | app1, app2 |
| group-b | OFF | app3 |

---

## 사전 요구

- Go 1.22+
- Node 18+
- Java 21+ (app1 의 Spring Boot 용. 없으면 brew install openjdk@21 maven)
- Python 3.11+ (app3 의 FastAPI 용. macOS 는 보통 기본 있음)

---

## 띄우기 — 7 개 터미널

각 터미널을 하나씩 열고 다음 명령을 실행한다. 총 7 개:

### 1. IdP 서버 (`server/`)

```bash
cd server
go run .
# → http://localhost:8080
```

### 2~3. app1 (Spring Boot + React)

```bash
# 터미널 2 — backend
cd examples/app1/backend
mvn spring-boot:run
# → http://localhost:8011
```

```bash
# 터미널 3 — frontend
cd examples/app1/frontend
npm install     # 첫 실행 시 한 번만
npm run dev
# → http://localhost:5181
```

### 4~5. app2 (Node Express + Vue)

```bash
# 터미널 4 — backend
cd examples/app2/backend
npm install     # 첫 실행 시 한 번만
npm start
# → http://localhost:8012
```

```bash
# 터미널 5 — frontend
cd examples/app2/frontend
npm install     # 첫 실행 시 한 번만
npm run dev
# → http://localhost:5182
```

### 6~7. app3 (FastAPI + Next.js)

```bash
# 터미널 6 — backend
cd examples/app3/backend
python3 -m venv .venv          # 첫 실행 시 한 번만
.venv/bin/pip install -r requirements.txt
.venv/bin/python main.py
# → http://localhost:8013
```

```bash
# 터미널 7 — frontend
cd examples/app3/frontend
npm install     # 첫 실행 시 한 번만
npm run dev
# → http://localhost:5183
```

> `.env` 파일은 backend 디렉토리에 이미 들어 있다. 시드 client_id/secret 이
> `server/store/clients.go` 와 정확히 일치하도록 미리 박혀 있음.

---

## 4 시나리오 재현 — silent SSO 가 진짜 동작하는지

### A. 첫 로그인

1. http://localhost:5181/ (app1, blue) 진입
2. **OAuth 로그인** 클릭
3. IdP 로그인 폼이 뜸 → `alice` / `password123` 입력
4. app1 으로 돌아옴, 카드에 `alice` 표시

### B. 같은 그룹 silent SSO ✨ (Phase 1 의 본질)

A 가 끝난 상태 (브라우저에 IdP 세션 쿠키 있음) 에서:

1. 새 탭으로 http://localhost:5182/ (app2, emerald) 진입
2. **OAuth 로그인** 클릭
3. **폼이 뜨지 않고 즉시 alice 로 로그인됨**
   - IdP 가 쿠키 받음 → group-a 매치 → silent code 발급 → app2 callback

### C. 다른 그룹 cross-group 차단 (Realm 경계)

A·B 가 끝난 상태 (alice 가 group-a 세션 보유) 에서:

1. 새 탭으로 http://localhost:5183/ (app3, amber) 진입
2. **OAuth 로그인** 클릭
3. **로그인 폼이 다시 표시됨**
   - IdP 쿠키는 있지만 group-b ≠ group-a 라 silent 차단

### D. cross-group 로그인 후 이전 그룹도 폼 다시

C 에서 app3 에 다시 로그인한 후:

1. 다시 http://localhost:5181/ (app1) 가서 로그아웃 후 OAuth 로그인 클릭
2. **폼이 다시 표시됨** (LastGroupID 가 group-b 로 바뀌어서)

---

## 표준 OIDC `prompt` 파라미터

silent 동작을 비대화식으로 확인하거나 강제 폼을 띄울 때 사용:

```bash
# 같은 그룹 silent 가능한 상태에서 비대화식 확인 → 302 + code
curl -i -b "<idp_session 쿠키>" \
  "http://localhost:8080/oauth/authorize?response_type=code&client_id=app1&redirect_uri=http%3A%2F%2Flocalhost%3A8011%2Fcallback&state=t&prompt=none"

# 세션 있어도 강제로 폼 → 200 (login.html)
curl -i -b "<idp_session 쿠키>" \
  "http://localhost:8080/oauth/authorize?response_type=code&client_id=app1&redirect_uri=http%3A%2F%2Flocalhost%3A8011%2Fcallback&state=t&prompt=login"

# 세션 없는 상태 + prompt=none → login_required 에러
curl -i "http://localhost:8080/oauth/authorize?response_type=code&client_id=app1&redirect_uri=http%3A%2F%2Flocalhost%3A8011%2Fcallback&state=t&prompt=none"
```

---

## 시드 사용자 (학습용 하드코딩)

| ID | 비밀번호 |
|-----|----------|
| alice | password123 |
| bob | password123 |
| carol | password123 |

다음 phase 에서 DB 의 `users` 테이블로 이전한다.

---

## 시드 클라이언트

| client_id | client_secret | 그룹 |
|-----------|---------------|------|
| app1 | app1-secret | group-a |
| app2 | app2-secret | group-a |
| app3 | app3-secret | group-b |

코드 시드 (`server/store/clients.go`). 다음 phase 의 어드민 UI 에서 등록으로 대체.

---

## 검증 안전망

- **단위 테스트** — 진리표 11 케이스: `go test ./server/policy/...`
- **통합 테스트** — 4 시나리오: `go test ./server/handlers/...`

```bash
cd server && go test ./...
```

---
