import { useCallback, useEffect, useState } from 'react'
import { api } from '../lib/api'
import type { CurrentUser } from '../lib/types'

interface UseCurrentUserResult {
  user: CurrentUser | null
  loading: boolean
  errorMsg: string | null
  logout: () => Promise<void>
}

// /api/me 로 현재 사용자 로드.
// URL 의 ?error 파라미터를 한 번 흡수해서 깔끔한 / 로 정리.
// 미로그인 + 에러도 없으면 /login 으로 자동 리다이렉트 (silent SSO 시도).
export function useCurrentUser(): UseCurrentUserResult {
  const [user, setUser] = useState<CurrentUser | null>(null)
  const [loading, setLoading] = useState(true)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const e = params.get('error')

    if (e) {
      window.history.replaceState({}, '', '/')
    }

    api.me()
      .then(data => {
        if (data) {
          setUser(data)
          setLoading(false)
        } else if (e) {
          setErrorMsg(e)
          setLoading(false)
        } else {
          window.location.href = '/login'
        }
      })
      .catch(() => setLoading(false))
  }, [])

  // 백엔드 /api/logout 으로 navigate → 백엔드가 세션 무효화 + IdP /oauth/logout 으로 redirect 체인.
  // IdP 가 자기 세션도 끊고 post_logout_redirect_uri (?logout=1) 로 돌려보냄.
  const logout = useCallback(async (): Promise<void> => {
    window.location.href = '/api/logout'
  }, [])

  return { user, loading, errorMsg, logout }
}
