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
    <aside className="w-56 bg-white border-r border-slate-200 flex flex-col">
      <div className="px-3 py-2 flex items-center justify-between border-b border-slate-200">
        <span className="text-xs uppercase tracking-wide text-slate-500">Notebooks</span>
        <button
          onClick={onAdd}
          className="text-blue-600 hover:text-blue-800 text-lg leading-none"
          title="새 노트북"
        >
          +
        </button>
      </div>
      <ul className="flex-1 overflow-y-auto">
        {notebooks.length === 0 && (
          <li className="px-3 py-4 text-sm text-slate-400">노트북이 없습니다</li>
        )}
        {notebooks.map(nb => (
          <li
            key={nb.id}
            className={`px-3 py-2 text-sm cursor-pointer group flex items-center justify-between ${
              selectedId === nb.id ? 'bg-blue-50 text-blue-900' : 'hover:bg-slate-50'
            }`}
            onClick={() => onSelect(nb.id)}
          >
            <span className="truncate">{nb.title}</span>
            <span className="hidden group-hover:flex gap-1 text-xs text-slate-500">
              <button onClick={e => { e.stopPropagation(); onRename(nb) }} className="hover:text-slate-900">✎</button>
              <button onClick={e => { e.stopPropagation(); onDelete(nb) }} className="hover:text-red-600">✕</button>
            </span>
          </li>
        ))}
      </ul>
    </aside>
  )
}
