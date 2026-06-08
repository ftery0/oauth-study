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
    <div className="min-h-screen bg-paper-100 text-ink-900 flex flex-col">
      <header className="bg-paper-50/80 backdrop-blur border-b border-paper-200 px-6 py-3 flex items-center gap-4">
        <button
          onClick={onBack}
          className="text-sm text-ink-500 hover:text-ink-900"
        >
          ← Notebook 으로
        </button>
        <h1 className="font-serif text-xl text-ink-900">프로필</h1>
        <div className="flex-1" />
        <button
          onClick={onLogout}
          className="text-xs uppercase tracking-wider text-ink-500 hover:text-ink-900"
        >
          로그아웃
        </button>
      </header>

      <main className="flex-1 px-6 py-12">
        <div className="max-w-2xl mx-auto">
          <div className="flex items-center gap-5 mb-8">
            <div className="w-16 h-16 rounded-full bg-emerald-800 text-paper-50 flex items-center justify-center font-semibold uppercase text-2xl">
              {displayName[0]}
            </div>
            <div>
              <div className="font-serif text-2xl text-ink-900">{displayName}</div>
              {user.preferred_username && user.preferred_username !== displayName && (
                <div className="text-sm italic text-ink-500">@{user.preferred_username}</div>
              )}
            </div>
          </div>

          <dl className="border-t border-paper-200 divide-y divide-paper-200 text-sm">
            <Row label="사용자 식별자" value={<span className="font-mono text-ink-700 break-all">{user.sub}</span>} />
            <Row label="로그인 ID" value={user.preferred_username ?? '—'} />
            <Row label="표시명" value={user.name ?? '—'} />
          </dl>

          <p className="mt-8 text-xs italic text-ink-400 leading-relaxed">
            이 정보는 같은 IdP 의 다른 앱(TaskBoard 등)과 공유됩니다.
            노트북 안의 내용은 이 앱에만 저장되며, 다른 앱과 분리되어 있어요.
          </p>
        </div>
      </main>
    </div>
  )
}

function Row({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="grid grid-cols-[max-content_1fr] gap-x-8 py-3">
      <dt className="text-[11px] uppercase tracking-[0.18em] text-ink-400 self-center">{label}</dt>
      <dd className="text-ink-700">{value}</dd>
    </div>
  )
}
