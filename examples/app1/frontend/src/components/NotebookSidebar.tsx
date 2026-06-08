import type { Notebook } from '../lib/types'

export function NotebookSidebar({
  notebooks,
  selectedId,
  onSelect,
  onAdd,
  onRename,
  onDelete,
}: {
  notebooks: Notebook[]
  selectedId: number | null
  onSelect: (id: number) => void
  onAdd: () => void
  onRename: (nb: Notebook) => void
  onDelete: (nb: Notebook) => void
}) {
  return (
    <aside className="w-60 bg-paper-100 border-r border-paper-200 flex flex-col">
      <div className="px-5 pt-6 pb-3 flex items-baseline justify-between">
        <span className="text-[11px] uppercase tracking-[0.18em] text-ink-400">노트북</span>
        <button
          onClick={onAdd}
          className="text-ink-500 hover:text-emerald-800 text-sm"
          title="노트북 추가"
        >
          + 추가
        </button>
      </div>
      <ul className="flex-1 overflow-y-auto px-3 pb-4 space-y-0.5">
        {notebooks.length === 0 && (
          <li className="px-3 py-6 text-sm text-ink-400 italic">
            첫 노트북을 만들어볼까요?
          </li>
        )}
        {notebooks.map(nb => {
          const active = selectedId === nb.id
          return (
            <li key={nb.id}>
              <div
                className={`group flex items-center justify-between rounded-md px-3 py-1.5 text-sm cursor-pointer transition-colors ${
                  active
                    ? 'bg-paper-200/70 text-ink-900 font-medium'
                    : 'text-ink-700 hover:bg-paper-200/40'
                }`}
                onClick={() => onSelect(nb.id)}
              >
                <span className="truncate flex items-center gap-2">
                  <span className={`text-xs ${active ? 'text-emerald-800' : 'text-ink-400'}`}>❦</span>
                  {nb.title}
                </span>
                <span className="hidden group-hover:flex gap-2 text-xs text-ink-400">
                  <button
                    onClick={e => { e.stopPropagation(); onRename(nb) }}
                    className="hover:text-ink-900"
                    title="이름 변경"
                  >
                    ✎
                  </button>
                  <button
                    onClick={e => { e.stopPropagation(); onDelete(nb) }}
                    className="hover:text-red-700"
                    title="삭제"
                  >
                    ✕
                  </button>
                </span>
              </div>
            </li>
          )
        })}
      </ul>
    </aside>
  )
}
