import { useEffect, useState } from 'react'
import { marked } from 'marked'
import type { Note } from '../lib/types'

// 디바운스(1.2초) 자동 저장 에디터. 좌측: textarea, 우측: marked preview.
export function NoteEditor({
  note,
  onSave,
}: {
  note: Note
  onSave: (patch: { title?: string; bodyMd?: string }) => Promise<void>
}) {
  const [title, setTitle] = useState(note.title)
  const [body, setBody] = useState(note.bodyMd)

  useEffect(() => {
    setTitle(note.title)
    setBody(note.bodyMd)
  }, [note.id])

  useEffect(() => {
    if (title === note.title && body === note.bodyMd) return
    const handle = setTimeout(() => {
      const patch: { title?: string; bodyMd?: string } = {}
      if (title !== note.title) patch.title = title
      if (body !== note.bodyMd) patch.bodyMd = body
      void onSave(patch)
    }, 1200)
    return () => clearTimeout(handle)
  }, [title, body])

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      <div className="px-6 py-3 border-b border-slate-200 bg-white">
        <input
          value={title}
          onChange={e => setTitle(e.target.value)}
          className="w-full text-xl font-semibold focus:outline-none bg-transparent"
          placeholder="제목"
        />
        <p className="text-xs text-slate-400 mt-1">
          최근 저장 {new Date(note.updatedAt).toLocaleString('ko-KR')}
        </p>
      </div>
      <div className="flex-1 flex overflow-hidden">
        <textarea
          value={body}
          onChange={e => setBody(e.target.value)}
          className="w-1/2 p-6 resize-none focus:outline-none font-mono text-sm bg-slate-50"
          placeholder="마크다운으로 작성하세요..."
        />
        <div
          className="w-1/2 p-6 overflow-y-auto bg-white border-l border-slate-200 prose prose-sm max-w-none"
          dangerouslySetInnerHTML={{ __html: marked.parse(body || '_미리보기_') as string }}
        />
      </div>
    </div>
  )
}
