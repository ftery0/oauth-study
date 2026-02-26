import { useState, useEffect } from 'react'

export default function App() {
  const [user, setUser]       = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError]     = useState(null)

  useEffect(() => {
    // URL에 ?error= 파라미터가 있으면 표시 (OAuth 서버가 에러 리턴한 경우)
    const params = new URLSearchParams(window.location.search)
    const errParam = params.get('error')
    if (errParam) {
      setError(errParam)
      window.history.replaceState({}, '', '/')
    }

    // 이미 로그인된 상태인지 확인
    fetch('/api/me')
      .then(r => r.ok ? r.json() : null)
      .then(data => {
        setUser(data)
        setLoading(false)
      })
      .catch(() => setLoading(false))
  }, [])

  const logout = async () => {
    await fetch('/api/logout', { method: 'POST' })
    setUser(null)
  }

  if (loading) {
    return (
      <div className="container">
        <div className="card">
          <p className="loading">로딩 중...</p>
        </div>
      </div>
    )
  }

  if (!user) {
    return <LoginPage error={error} />
  }

  return <UserPage user={user} onLogout={logout} />
}

// ──────────────────────────────────────────
// 로그인 전 페이지
// ──────────────────────────────────────────
function LoginPage({ error }) {
  return (
    <div className="container">
      <div className="card">
        <div className="login-header">
          <h1>OAuth 2.0 Demo</h1>
          <p>Go로 직접 만든 OAuth 서버를 Express + React에서 연동하는 예제입니다.</p>
        </div>

        {error && (
          <div className="error-msg">오류: {error}</div>
        )}

        <div className="flow-box">
          <h3>Authorization Code Flow</h3>
          <ol className="flow-steps">
            <li>아래 버튼 클릭 → Express가 OAuth 서버로 redirect</li>
            <li>OAuth 서버 로그인 페이지 (alice / password123)</li>
            <li>로그인 성공 → OAuth 서버가 auth code 발급</li>
            <li>Express가 code를 access token + refresh token으로 교환</li>
            <li>토큰은 서버 세션에 저장 (브라우저에는 세션 쿠키만 전달)</li>
            <li>이 페이지에서 /api/me 호출 → 유저 정보 표시</li>
          </ol>
        </div>

        <a href="/login" className="btn-login">
          Login with OAuth
        </a>
      </div>
    </div>
  )
}

// ──────────────────────────────────────────
// 로그인 후 유저 정보 페이지
// ──────────────────────────────────────────
function UserPage({ user, onLogout }) {
  return (
    <div className="container">
      <div className="card">
        <div className="user-header">
          <div className="avatar">{user.sub[0].toUpperCase()}</div>
          <div>
            <h1>{user.sub}</h1>
            <p>로그인 성공</p>
          </div>
        </div>

        <div className="token-section">
          <p className="section-label">유저 정보 (/oauth/userinfo)</p>
          <div className="info-table">
            <div className="info-row">
              <span className="info-key">sub (사용자 ID)</span>
              <span className="info-val">{user.sub}</span>
            </div>
            <div className="info-row">
              <span className="info-key">client_id</span>
              <span className="info-val">{user.client_id}</span>
            </div>
            <div className="info-row">
              <span className="info-key">scope</span>
              <span className="info-val">
                {user.scope.split(' ').map(s => (
                  <span key={s} className="scope-badge" style={{ marginLeft: 4 }}>{s}</span>
                ))}
              </span>
            </div>
          </div>
        </div>

        <div className="token-note">
          Access token은 <strong>15분</strong> 후 만료됩니다.
          만료되면 Express가 <strong>refresh token</strong>으로 자동 갱신하므로
          사용자는 다시 로그인할 필요가 없습니다.
        </div>

        <button onClick={onLogout} className="btn-logout">
          Logout
        </button>
      </div>
    </div>
  )
}
