import type { Note } from '../lib/types'
import { relativeKo } from '../lib/format'

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
    <section className="w-80 bg-paper-50 border-r border-paper-200 flex flex-col">
      <div className="px-5 pt-6 pb-3 flex items-baseline justify-between">
        <span className="text-[11px] uppercase tracking-[0.18em] text-ink-400">
          {searchQ ? '검색 결과' : '노트'}
        </span>
        <button
          onClick={onAdd}
          disabled={addDisabled}
          className="text-sm text-ink-500 hover:text-emerald-800 disabled:text-paper-300 disabled:hover:text-paper-300"
          title="새 노트 작성"
        >
          + 새 노트
        </button>
      </div>
      {searchQ && (
        <div className="px-5 pb-2 text-xs italic text-ink-500">
          "{searchQ}" 에 대한 결과
        </div>
      )}
      <ul className="flex-1 overflow-y-auto px-3 pb-4 space-y-1">
        {notes.length === 0 && (
          <li className="px-3 py-10 text-sm text-ink-400 italic text-center">
            {!notebookSelected && !searchQ
              ? '왼쪽에서 노트북을 골라주세요.'
              : searchQ
                ? '일치하는 노트가 없습니다.'
                : '이 노트북은 아직 비어 있습니다.'}
          </li>
        )}
        {notes.map(n => {
          const active = selectedNote?.id === n.id
          const preview = n.bodyMd.replace(/[#>*_`\-]/g, '').trim().slice(0, 140)
          return (
            <li
              key={n.id}
              className={`group rounded-md px-3 py-3 cursor-pointer border transition-colors ${
                active
                  ? 'bg-paper-100 border-paper-300 shadow-inner'
                  : 'border-transparent hover:bg-paper-100/60'
              }`}
              onClick={() => onSelect(n)}
            >
              <div className="flex items-start justify-between gap-2">
                <h3 className="font-serif text-base text-ink-900 leading-snug truncate flex-1">
                  {n.title || <span className="italic text-ink-400">제목 없음</span>}
                </h3>
                <button
                  onClick={e => { e.stopPropagation(); onDelete(n) }}
                  className="hidden group-hover:inline text-xs text-ink-400 hover:text-red-700 mt-0.5"
                  title="삭제"
                >
                  ✕
                </button>
              </div>
              <p className="text-xs text-ink-500 mt-1 leading-relaxed line-clamp-3">
                {preview || <span className="italic text-ink-400">아직 비어 있는 노트</span>}
              </p>
              <div className="mt-2 text-[11px] text-ink-400 italic">
                마지막 수정 · {relativeKo(n.updatedAt)}
              </div>
            </li>
          )
        })}
      </ul>
    </section>
  )
}
