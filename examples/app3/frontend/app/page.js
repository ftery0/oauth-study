'use client'

import { useEffect, useState } from 'react'

export default function Page() {
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(true)
  const [errorMsg, setErrorMsg] = useState(null)

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const e = params.get('error')
    if (e) {
      setErrorMsg(e)
      window.history.replaceState({}, '', '/')
    }

    // App Router 의 fetch 캐싱 회피
    fetch('/api/me', { cache: 'no-store' })
      .then((r) => (r.ok ? r.json() : null))
      .then((data) => {
        if (data) setUser(data)
        setLoading(false)
      })
      .catch(() => setLoading(false))
  }, [])

  const logout = async () => {
    await fetch('/api/logout', { method: 'POST', cache: 'no-store' })
    setUser(null)
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
            <span className="inline-block w-2 h-2 rounded-full bg-amber-500" />
            <span className="text-xs font-medium text-amber-700 uppercase tracking-wide">
              group-b
            </span>
          </div>
          <h1 className="text-2xl font-semibold">App 3</h1>
          <p className="mt-1 text-sm text-slate-500">Python FastAPI + Next.js</p>
        </header>

        {error && (
          <div
            className="mb-4 rounded-lg bg-red-50 border border-red-200 px-3 py-2 text-sm text-red-700"
            role="alert"
          >
            오류: {error}
          </div>
        )}

        <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 mb-5 text-sm text-amber-900">
          <p className="font-medium mb-1">cross-group 차단 시연</p>
          <p>
            app1/app2(group-a)에 이미 로그인되어 있어도, 이 앱은 다른 그룹(group-b)이라
            <strong> 로그인 폼이 다시 표시</strong>됩니다.
          </p>
        </div>

        <a
          href="/login"
          className="block w-full text-center rounded-lg bg-amber-600 hover:bg-amber-700 active:bg-amber-800 text-white font-medium px-4 py-3 transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500"
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
          <div className="w-12 h-12 rounded-full bg-amber-600 text-white flex items-center justify-center font-semibold uppercase text-lg">
            {user.sub[0]}
          </div>
          <div>
            <h1 className="text-xl font-semibold">{user.sub}</h1>
            <p className="text-sm text-slate-500">App 3 로그인 완료</p>
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
