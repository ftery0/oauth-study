import { useCurrentUser } from './hooks/useCurrentUser'
import { useRoute } from './hooks/useRoute'
import { LoginPage } from './pages/LoginPage'
import { NotebookPage } from './pages/NotebookPage'
import { ProfilePage } from './pages/ProfilePage'

export default function App() {
  const { user, loading, errorMsg, logout } = useCurrentUser()
  const { view, goProfile, goNotebook } = useRoute()

  if (loading) {
    return (
      <div className="min-h-screen bg-paper-100 flex items-center justify-center p-4">
        <p className="font-serif italic text-ink-500">노트를 펼치는 중…</p>
      </div>
    )
  }

  if (!user) return <LoginPage error={errorMsg} />
  if (view === 'profile') return <ProfilePage user={user} onBack={goNotebook} onLogout={logout} />
  return <NotebookPage user={user} onProfile={goProfile} onLogout={logout} />
}
