// 실제 Chromium 으로 사용자 흐름 검증.
//   1. localhost:5181 → IdP 로그인 폼 자동 표시
//   2. 신규 회원가입 → app1 노트북 UI 진입
//   3. Welcome seed 노트북 + 노트 보임
//   4. 노트 검색 동작
//   5. 프로필 페이지 진입 → 뒤로가기
//   6. 로그아웃 → IdP 폼 다시 표시 (silent SSO 차단 확인)

const { test, expect } = require('@playwright/test')

const APP1 = 'http://localhost:5181'
const USERNAME = `e2e${Date.now().toString().slice(-7)}`
const PASSWORD = 'browserTest!1'
const EMAIL = `${USERNAME}@example.com`

test.describe.configure({ mode: 'serial' })

test('전체 사용자 흐름', async ({ page, context }) => {
  test.setTimeout(60_000)

  // ─── 1. 메인 진입 → IdP 로그인 폼 자동 표시 ───
  await page.goto(APP1)
  await page.waitForURL(/oauth\/authorize/)
  await expect(page.getByRole('button', { name: /로그인/ })).toBeVisible()
  console.log('  ✅ IdP 로그인 폼 자동 진입')

  // ─── 2. 회원가입 링크 → 폼 작성 → 가입 ───
  await page.getByRole('link', { name: '회원가입' }).click()
  await expect(page.getByText(/회원가입/).first()).toBeVisible()

  await page.locator('#username').fill(USERNAME)
  await page.locator('#email').fill(EMAIL)
  await page.locator('#password').fill(PASSWORD)
  await page.locator('#password_confirm').fill(PASSWORD)
  await page.getByRole('button', { name: /가입하고 계속/ }).click()

  // ─── 3. app1 노트북 UI 진입 + Welcome seed 표시 ───
  await page.waitForURL(`${APP1}/`)
  await expect(page.getByRole('heading', { name: 'Notebook' })).toBeVisible()

  // 헤더에 사용자명 표시
  await expect(page.locator('header').getByText(USERNAME)).toBeVisible()
  console.log('  ✅ 로그인 후 노트북 UI 진입, 헤더에 username 표시')

  // 좌측 사이드바에 두 노트북
  await expect(page.getByText('시작하기')).toBeVisible()
  await expect(page.getByText('프로젝트 메모')).toBeVisible()
  console.log('  ✅ Welcome seed 노트북 2 개 보임')

  // 시작하기 노트북 클릭 → 노트 3 개 (제목으로 localized — text-sm font-medium 클래스)
  await page.getByText('시작하기').first().click()
  await expect(page.locator('.text-sm.font-medium', { hasText: '환영합니다' })).toBeVisible()
  await expect(page.locator('.text-sm.font-medium', { hasText: '마크다운 치트시트' })).toBeVisible()
  await expect(page.locator('.text-sm.font-medium', { hasText: '팁' })).toBeVisible()
  console.log('  ✅ 시작하기 노트 3 개 보임')

  // 환영합니다 노트 제목 클릭 → 우측에 마크다운 렌더
  await page.locator('.text-sm.font-medium', { hasText: '환영합니다' }).click()
  await expect(page.locator('.prose')).toContainText('환영합니다')
  await expect(page.locator('.prose strong').first()).toBeVisible() // **Markdown** 굵게
  await expect(page.locator('.prose h1').first()).toBeVisible() // # 환영합니다 heading
  console.log('  ✅ 마크다운 미리보기 렌더 (heading + 굵게)')

  // ─── 4. 검색 동작 ───
  await page.locator('input[type="search"]').fill('마크다운')
  await page.waitForTimeout(400)  // 200ms debounce + margin
  await expect(page.getByText('마크다운 치트시트')).toBeVisible()
  console.log('  ✅ 검색 디바운스 동작')

  // 검색 키워드 변경 — 'OAuth' 가 마크다운 치트시트 본문에 있음
  await page.locator('input[type="search"]').fill('OAuth')
  await page.waitForTimeout(400)
  await expect(page.locator('.text-sm.font-medium', { hasText: '마크다운 치트시트' })).toBeVisible()
  console.log('  ✅ 다른 키워드도 검색됨')

  // 검색 비우기
  await page.locator('input[type="search"]').fill('')

  // ─── 5. 프로필 페이지 진입 → 뒤로가기 ───
  await page.locator('header button[title="프로필 보기"]').click()
  await page.waitForURL(`${APP1}/profile`)
  await expect(page.getByRole('heading', { name: '프로필' })).toBeVisible()
  await expect(page.getByText(USERNAME).first()).toBeVisible()
  console.log('  ✅ 프로필 페이지 진입')

  await page.getByRole('button', { name: /← Notebook/ }).click()
  await page.waitForURL(`${APP1}/`)
  await expect(page.getByRole('heading', { name: 'Notebook' })).toBeVisible()
  console.log('  ✅ Notebook 으로 복귀')

  // ─── 6. 로그아웃 → IdP 폼 다시 표시 (silent SSO 차단) ───
  await page.getByRole('button', { name: '로그아웃' }).click()
  // 체인: /api/logout → IdP /oauth/logout → frontend → useCurrentUser → /login → IdP /authorize → 로그인 폼
  await page.waitForURL(/oauth\/authorize/, { timeout: 10_000 })
  await expect(page.getByRole('button', { name: /로그인/ })).toBeVisible()
  console.log('  ✅ 로그아웃 후 IdP 폼 (silent SSO 차단됨)')

  // 동일 cookie 로 다시 메인 가도 silent SSO 안 됨 (IdP 세션이 죽었으니)
  await page.goto(APP1)
  await page.waitForURL(/oauth\/authorize/)
  await expect(page.getByRole('button', { name: /로그인/ })).toBeVisible()
  console.log('  ✅ 메인 재진입해도 silent SSO 안 됨 → IdP 폼')
})

test('잘못된 비밀번호 → 에러', async ({ page }) => {
  await page.goto(APP1)
  await page.waitForURL(/oauth\/authorize/)

  await page.locator('#id').fill('alice')
  await page.locator('#password').fill('wrong-password-xxx')
  await page.getByRole('button', { name: /로그인/ }).click()

  // 폼이 다시 보이고 에러 메시지 표시
  await expect(page.getByText(/아이디 또는 비밀번호가 틀렸습니다/)).toBeVisible()
  console.log('  ✅ 잘못된 비번 → 폼 + 에러 메시지')
})

test('alice 로그인 후 데이터 격리', async ({ page }) => {
  await page.goto(APP1)
  await page.waitForURL(/oauth\/authorize/)

  await page.locator('#id').fill('alice')
  await page.locator('#password').fill('password123')
  await page.getByRole('button', { name: /로그인/ }).click()

  await page.waitForURL(`${APP1}/`)
  await expect(page.locator('header').getByText('Alice')).toBeVisible()
  console.log('  ✅ alice 로그인 — 헤더에 Alice 표시')

  // alice 는 본인 노트북만 보여야 함 (다른 사용자 e2e* 데이터는 안 보임)
  const sidebarText = await page.locator('aside').textContent()
  if (sidebarText && sidebarText.includes('e2e')) {
    throw new Error('alice 사이드바에 다른 사용자의 노트북이 보임 — 데이터 격리 실패')
  }
  console.log('  ✅ alice 노트북에 다른 사용자 데이터 안 보임 (격리)')

  await page.getByRole('button', { name: '로그아웃' }).click()
  await page.waitForURL(/oauth\/authorize/, { timeout: 10_000 })
  console.log('  ✅ alice 로그아웃 완료')
})
