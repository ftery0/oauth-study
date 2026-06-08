import { useEffect, useState } from 'react'
import { marked } from 'marked'
import type { Note } from '../lib/types'
import { relativeKo } from '../lib/format'

// 디바운스(1.2초) 자동 저장 에디터. 좌측: textarea(serif), 우측: marked preview.
export function NoteEditor({
  note,
  onSave,
}: {
  note: Note
  onSave: (patch: { title?: string; bodyMd?: string }) => Promise<void>
}) {
  const [title, setTitle] = useState(note.title)
  const [body, setBody] = useState(note.bodyMd)
  const [dirty, setDirty] = useState(false)

  useEffect(() => {
    setTitle(note.title)
    setBody(note.bodyMd)
    setDirty(false)
  }, [note.id])

  useEffect(() => {
    if (title === note.title && body === note.bodyMd) {
      setDirty(false)
      return
    }
    setDirty(true)
    const handle = setTimeout(async () => {
      const patch: { title?: string; bodyMd?: string } = {}
      if (title !== note.title) patch.title = title
      if (body !== note.bodyMd) patch.bodyMd = body
      await onSave(patch)
      setDirty(false)
    }, 1200)
    return () => clearTimeout(handle)
  }, [title, body])

  return (
    <div className="flex-1 flex flex-col overflow-hidden bg-paper-50">
      <div className="px-10 pt-10 pb-4 max-w-5xl w-full self-center">
        <input
          value={title}
          onChange={e => setTitle(e.target.value)}
          className="w-full font-serif text-3xl text-ink-900 focus:outline-none bg-transparent placeholder:text-ink-400"
          placeholder="제목을 적어주세요"
        />
        <p className="text-xs text-ink-400 italic mt-2">
          {dirty
            ? '저장 중…'
            : `마지막으로 저장한 시각 · ${relativeKo(note.updatedAt)}`}
        </p>
      </div>
      <div className="flex-1 flex overflow-hidden border-t border-paper-200">
        <textarea
          value={body}
          onChange={e => setBody(e.target.value)}
          className="w-1/2 px-10 py-8 resize-none focus:outline-none font-serif text-[15px] leading-7 bg-paper-50 text-ink-900 placeholder:italic placeholder:text-ink-400"
          placeholder="여기에 자유롭게 적어보세요. 마크다운 문법도 사용할 수 있어요."
        />
        <div
          className="w-1/2 px-10 py-8 overflow-y-auto bg-paper-100/60 border-l border-paper-200 prose prose-stone prose-sm max-w-none prose-headings:font-serif prose-headings:text-ink-900"
          dangerouslySetInnerHTML={{
            __html: marked.parse(body || '_미리보기는 이곳에 나타납니다._') as string,
          }}
        />
      </div>
    </div>
  )
}
