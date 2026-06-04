import { useState, useEffect } from 'react'

export default function App() {
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(true)
  const [errorMsg, setErrorMsg] = useState(null)

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const e = params.get('error')
    const loggedOut = params.get('logout')
    if (e) {
      setErrorMsg(e)
      window.history.replaceState({}, '', '/')
    }
    if (loggedOut) {
      // 로그아웃 직후 자동 재로그인 루프 방지 — 한 번은 LoginPage 보여줌
      window.history.replaceState({}, '', '/')
      setLoading(false)
      return
    }

    fetch('/api/me')
      .then((r) => (r.ok ? r.json() : null))
      .then((data) => {
        if (data) {
          setUser(data)
          setLoading(false)
        } else if (e) {
          // 에러가 있으면 자동 재시도 안 함
          setLoading(false)
        } else {
          // Keycloak 패턴: 401 + 에러 없음 → 즉시 OAuth 시작
          // IdP 세션이 있으면 silent SSO 로 통과, 없으면 폼 표시
          window.location.href = '/login'
        }
      })
      .catch(() => setLoading(false))
  }, [])

  const logout = async () => {
    await fetch('/api/logout', { method: 'POST' })
    // 로그아웃 후 즉시 재로그인 막기 위해 marker 동봉
    window.location.href = '/?logout=1'
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-slate-50 flex items-center justify-center p-4">
        <p className="text-slate-500 text-sm">로딩 중...</p>
      </div>
    )
  }

  if (!user) return <LoginPage error={errorMsg} />
  return <UserPage user={user} onLogout={logout} />
}

function LoginPage({ error }) {
  return (
    <div className="min-h-screen bg-slate-50 flex items-center justify-center p-4 antialiased text-slate-900">
      <main className="w-full max-w-md bg-white rounded-2xl shadow-lg p-6 sm:p-8">
        <header className="mb-6">
          <div className="inline-flex items-center gap-2 mb-2">
            <span className="inline-block w-2 h-2 rounded-full bg-blue-500" />
            <span className="text-xs font-medium text-blue-700 uppercase tracking-wide">group-a</span>
          </div>
          <h1 className="text-2xl font-semibold">App 1</h1>
          <p className="mt-1 text-sm text-slate-500">Spring Boot + React</p>
        </header>

        {error && (
          <div
            className="mb-4 rounded-lg bg-red-50 border border-red-200 px-3 py-2 text-sm text-red-700"
            role="alert"
          >
            오류: {error}
          </div>
        )}

        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-5 text-sm text-blue-900">
          <p className="font-medium mb-1">silent SSO 시연</p>
          <p>이미 같은 그룹(group-a)의 다른 앱(app2)에 로그인되어 있다면 폼 없이 자동 로그인됩니다.</p>
        </div>

        <a
          href="/login"
          className="block w-full text-center rounded-lg bg-blue-600 hover:bg-blue-700 active:bg-blue-800 text-white font-medium px-4 py-3 transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
        >
          OAuth 로그인
        </a>
      </main>
    </div>
  )
}

function UserPage({ user, onLogout }) {
  return (
    <div className="min-h-screen bg-slate-50 flex items-center justify-center p-4 antialiased text-slate-900">
      <main className="w-full max-w-md bg-white rounded-2xl shadow-lg p-6 sm:p-8">
        <div className="flex items-center gap-4 mb-6">
          <div className="w-12 h-12 rounded-full bg-blue-600 text-white flex items-center justify-center font-semibold uppercase text-lg">
            {user.sub[0]}
          </div>
          <div>
            <h1 className="text-xl font-semibold">{user.sub}</h1>
            <p className="text-sm text-slate-500">App 1 로그인 완료</p>
          </div>
        </div>

        <div className="space-y-2 mb-6 text-sm">
          <Row label="sub" value={user.sub} />
          <Row label="client_id" value={user.client_id} />
          <Row label="scope" value={user.scope || '(none)'} />
        </div>

        <button
          onClick={onLogout}
          className="w-full rounded-lg border border-slate-300 hover:bg-slate-100 active:bg-slate-200 text-slate-700 font-medium px-4 py-3 transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-slate-400"
        >
          로그아웃
        </button>
      </main>
    </div>
  )
}

function Row({ label, value }) {
  return (
    <div className="flex justify-between border-b last:border-b-0 pb-2">
      <span className="text-slate-500">{label}</span>
      <span className="font-mono text-slate-900">{value}</span>
    </div>
  )
}
