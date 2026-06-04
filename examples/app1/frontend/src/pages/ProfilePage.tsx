import { pickDisplayName } from '../lib/types'
import type { CurrentUser } from '../lib/types'

export function ProfilePage({
  user,
  onBack,
  onLogout,
}: {
  user: CurrentUser
  onBack: () => void
  onLogout: () => void
}) {
  const displayName = pickDisplayName(user)
  return (
    <div className="min-h-screen bg-slate-50 text-slate-900 antialiased flex flex-col">
      <header className="bg-white border-b border-slate-200 px-4 py-3 flex items-center gap-3">
        <button
          onClick={onBack}
          className="text-sm text-slate-600 hover:text-slate-900 rounded border border-slate-300 px-3 py-1.5 hover:bg-slate-100"
        >
          ← Notebook
        </button>
        <h1 className="font-semibold">프로필</h1>
        <div className="flex-1" />
        <button
          onClick={onLogout}
          className="text-sm text-slate-600 hover:text-slate-900 rounded border border-slate-300 px-3 py-1.5 hover:bg-slate-100"
        >
          로그아웃
        </button>
      </header>

      <main className="flex-1 p-6 sm:p-10">
        <div className="max-w-2xl mx-auto bg-white rounded-2xl shadow-sm border border-slate-200 p-6 sm:p-8">
          <div className="flex items-center gap-4 mb-6">
            <div className="w-16 h-16 rounded-full bg-blue-600 text-white flex items-center justify-center font-semibold uppercase text-2xl">
              {displayName[0]}
            </div>
            <div>
              <div className="text-xl font-semibold">{displayName}</div>
              {user.preferred_username && user.preferred_username !== displayName && (
                <div className="text-sm text-slate-500">@{user.preferred_username}</div>
              )}
            </div>
          </div>

          <dl className="grid grid-cols-1 sm:grid-cols-[max-content_1fr] gap-x-6 gap-y-3 text-sm">
            <dt className="text-slate-500">사용자 ID</dt>
            <dd className="font-mono text-slate-700 break-all">{user.sub}</dd>

            <dt className="text-slate-500">로그인 ID</dt>
            <dd className="text-slate-700">{user.preferred_username ?? '—'}</dd>

            <dt className="text-slate-500">표시명</dt>
            <dd className="text-slate-700">{user.name ?? '—'}</dd>
          </dl>
        </div>
      </main>
    </div>
  )
}
