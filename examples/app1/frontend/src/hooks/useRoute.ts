import { useCallback, useEffect, useState } from 'react'

export type View = 'notebook' | 'profile'

function readView(): View {
  return window.location.pathname === '/profile' ? 'profile' : 'notebook'
}

// 2-뷰 라우터. react-router 안 쓰고 pushState/popstate 만으로 처리.
export function useRoute() {
  const [view, setView] = useState<View>(readView)

  useEffect(() => {
    const onPop = () => setView(readView())
    window.addEventListener('popstate', onPop)
    return () => window.removeEventListener('popstate', onPop)
  }, [])

  const goProfile = useCallback(() => {
    window.history.pushState({}, '', '/profile')
    setView('profile')
  }, [])
  const goNotebook = useCallback(() => {
    window.history.pushState({}, '', '/')
    setView('notebook')
  }, [])

  return { view, goProfile, goNotebook }
}
