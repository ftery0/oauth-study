import type { Note } from '../lib/types'

export function NoteList({
  notes,
  selectedNote,
  onSelect,
  onAdd,
  onDelete,
  searchQ,
  notebookSelected,
}: {
  notes: Note[]
  selectedNote: Note | null
  onSelect: (n: Note) => void
  onAdd: () => void
  onDelete: (n: Note) => void
  searchQ: string
  notebookSelected: boolean
}) {
  const addDisabled = !notebookSelected || !!searchQ
  return (
    <section className="w-72 bg-white border-r border-slate-200 flex flex-col">
      <div className="px-3 py-2 flex items-center justify-between border-b border-slate-200">
        <span className="text-xs uppercase tracking-wide text-slate-500">
          {searchQ ? `검색: ${searchQ}` : '노트'}
        </span>
        <button
          onClick={onAdd}
          disabled={addDisabled}
          className="text-blue-600 hover:text-blue-800 disabled:text-slate-300 text-lg leading-none"
          title="새 노트"
        >
          +
        </button>
      </div>
      <ul className="flex-1 overflow-y-auto">
        {notes.length === 0 && (
          <li className="px-3 py-4 text-sm text-slate-400">
            {!notebookSelected && !searchQ ? '노트북을 선택하세요' : '노트가 없습니다'}
          </li>
        )}
        {notes.map(n => (
          <li
            key={n.id}
            className={`px-3 py-2 cursor-pointer group ${
              selectedNote?.id === n.id ? 'bg-blue-50' : 'hover:bg-slate-50'
            }`}
            onClick={() => onSelect(n)}
          >
            <div className="flex items-center justify-between">
              <div className="text-sm font-medium truncate">{n.title}</div>
              <button
                onClick={e => { e.stopPropagation(); onDelete(n) }}
                className="hidden group-hover:inline text-xs text-slate-400 hover:text-red-600"
              >
                ✕
              </button>
            </div>
            <div className="text-xs text-slate-500 mt-0.5 line-clamp-2">
              {n.bodyMd.slice(0, 80) || <span className="italic text-slate-400">비어있음</span>}
            </div>
          </li>
        ))}
      </ul>
    </section>
  )
}
