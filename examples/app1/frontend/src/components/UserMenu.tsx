import { pickDisplayName } from '../lib/types'
import type { CurrentUser } from '../lib/types'

// 헤더 우측: 아바타+이름(프로필 진입) + 로그아웃.
// NotebookPage / ProfilePage 양쪽에서 재사용 (단, ProfilePage 헤더는 별도 디자인이라 거기선 안 씀).
export function UserMenu({
  user,
  onProfile,
  onLogout,
}: {
  user: CurrentUser
  onProfile: () => void
  onLogout: () => void
}) {
  const displayName = pickDisplayName(user)
  return (
    <div className="flex items-center gap-3">
      <button
        onClick={onProfile}
        title="프로필 보기"
        className="flex items-center gap-2 rounded-full pr-2 hover:bg-paper-200/60 focus:outline-none focus:ring-2 focus:ring-emerald-700"
      >
        <span className="w-8 h-8 rounded-full bg-emerald-800 text-paper-50 flex items-center justify-center font-semibold uppercase text-sm">
          {displayName[0]}
        </span>
        <span className="text-sm text-ink-700 max-w-[10rem] truncate">{displayName}</span>
      </button>
      <button
        onClick={onLogout}
        className="text-xs uppercase tracking-wider text-ink-500 hover:text-ink-900"
      >
        로그아웃
      </button>
    </div>
  )
}
