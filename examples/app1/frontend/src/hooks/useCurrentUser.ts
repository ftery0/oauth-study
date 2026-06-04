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
// URL 의 ?error / ?logout 파라미터를 한 번 흡수해서 깔끔한 / 로 정리.
// 미로그인 + 에러도 없으면 /login 으로 자동 리다이렉트 (silent SSO 시도).
export function useCurrentUser(): UseCurrentUserResult {
  const [user, setUser] = useState<CurrentUser | null>(null)
  const [loading, setLoading] = useState(true)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const e = params.get('error')
    const loggedOut = params.get('logout')

    if (e || loggedOut) {
      window.history.replaceState({}, '', '/')
    }
    if (loggedOut) {
      setLoading(false)
      return
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

  const logout = useCallback(async (): Promise<void> => {
    await api.logout()
    window.location.href = '/?logout=1'
  }, [])

  return { user, loading, errorMsg, logout }
}
