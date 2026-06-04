import { useEffect, useState } from 'react'
import { marked } from 'marked'
import { api } from './api'
import type { CurrentUser, Note, Notebook } from './types'

export default function App() {
  const [user, setUser] = useState<CurrentUser | null>(null)
  const [loading, setLoading] = useState(true)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const e = params.get('error')
    const loggedOut = params.get('logout')
    if (e) {
      setErrorMsg(e)
      window.history.replaceState({}, '', '/')
    }
    if (loggedOut) {
      window.history.replaceState({}, '', '/')
      setLoading(false)
      return
    }

    api.me()
      .then(data => {
        if (data) {
          setUser(data)
          setLoading(false)
        } else if (e) {
          setLoading(false)
        } else {
          window.location.href = '/login'
        }
      })
      .catch(() => setLoading(false))
  }, [])

  const logout = async (): Promise<void> => {
    await api.logout()
    window.location.href = '/?logout=1'
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-slate-50 flex items-center justify-center p-4">
        <p className="text-slate-500 text-sm">로딩 중...</p>
      </div>
    )
  }

  if (!user) return <LoginPage error={errorMsg} />
  return <NotebookApp user={user} onLogout={logout} />
}

// ─────────────────────────────────────────────
// 로그인 페이지
// ─────────────────────────────────────────────
function LoginPage({ error }: { error: string | null }) {
  return (
    <div className="min-h-screen bg-slate-50 flex items-center justify-center p-4 antialiased text-slate-900">
      <main className="w-full max-w-md bg-white rounded-2xl shadow-lg p-6 sm:p-8">
        <header className="mb-6">
          <div className="inline-flex items-center gap-2 mb-2">
            <span className="inline-block w-2 h-2 rounded-full bg-blue-500" />
            <span className="text-xs font-medium text-blue-700 uppercase tracking-wide">group-a</span>
          </div>
          <h1 className="text-2xl font-semibold">Notebook</h1>
          <p className="mt-1 text-sm text-slate-500">사내 노트 · Spring Boot + React</p>
        </header>

        {error && (
          <div className="mb-4 rounded-lg bg-red-50 border border-red-200 px-3 py-2 text-sm text-red-700" role="alert">
            오류: {error}
          </div>
        )}

        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-5 text-sm text-blue-900">
          <p className="font-medium mb-1">silent SSO 시연</p>
          <p>이미 같은 그룹(group-a)의 TaskBoard 에 로그인되어 있다면 폼 없이 자동 로그인됩니다.</p>
        </div>

        <a
          href="/login"
          className="block w-full text-center rounded-lg bg-blue-600 hover:bg-blue-700 active:bg-blue-800 text-white font-medium px-4 py-3 transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
        >
          OAuth 로그인
        </a>
      </main>
    </div>
  )
}

// ─────────────────────────────────────────────
// 메인 앱
// ─────────────────────────────────────────────
function NotebookApp({ user, onLogout }: { user: CurrentUser; onLogout: () => void }) {
  const [notebooks, setNotebooks] = useState<Notebook[]>([])
  const [selectedNotebookId, setSelectedNotebookId] = useState<number | null>(null)
  const [notes, setNotes] = useState<Note[]>([])
  const [selectedNote, setSelectedNote] = useState<Note | null>(null)
  const [searchQ, setSearchQ] = useState('')

  // 최초 로드: notebook 목록
  useEffect(() => {
    api.listNotebooks().then(list => {
      setNotebooks(list)
      if (list.length > 0) setSelectedNotebookId(list[0].id)
    })
  }, [])

  // notebook 선택 변경 → 노트 목록
  useEffect(() => {
    if (selectedNotebookId == null) {
      setNotes([])
      setSelectedNote(null)
      return
    }
    api.listNotes(selectedNotebookId).then(list => {
      setNotes(list)
      setSelectedNote(list[0] ?? null)
    })
  }, [selectedNotebookId])

  // 검색 (디바운스 200ms)
  useEffect(() => {
    if (!searchQ.trim()) return
    const handle = setTimeout(() => {
      api.searchNotes(searchQ.trim()).then(list => {
        setNotes(list)
        setSelectedNote(list[0] ?? null)
      })
    }, 200)
    return () => clearTimeout(handle)
  }, [searchQ])

  const addNotebook = async (): Promise<void> => {
    const title = window.prompt('새 노트북 이름')?.trim()
    if (!title) return
    const created = await api.createNotebook(title)
    setNotebooks(prev => [created, ...prev])
    setSelectedNotebookId(created.id)
  }

  const renameNotebook = async (nb: Notebook): Promise<void> => {
    const title = window.prompt('노트북 이름 변경', nb.title)?.trim()
    if (!title || title === nb.title) return
    const updated = await api.renameNotebook(nb.id, title)
    setNotebooks(prev => prev.map(n => (n.id === nb.id ? updated : n)))
  }

  const deleteNotebook = async (nb: Notebook): Promise<void> => {
    if (!window.confirm(`"${nb.title}" 와 안의 노트를 모두 삭제할까요?`)) return
    await api.deleteNotebook(nb.id)
    setNotebooks(prev => prev.filter(n => n.id !== nb.id))
    if (selectedNotebookId === nb.id) setSelectedNotebookId(null)
  }

  const addNote = async (): Promise<void> => {
    if (selectedNotebookId == null) return
    const title = window.prompt('새 노트 제목', '제목 없는 노트')?.trim()
    if (!title) return
    const created = await api.createNote(selectedNotebookId, title, '')
    setNotes(prev => [created, ...prev])
    setSelectedNote(created)
  }

  const saveNote = async (patch: { title?: string; bodyMd?: string }): Promise<void> => {
    if (!selectedNote) return
    const updated = await api.updateNote(selectedNote.id, patch)
    setSelectedNote(updated)
    setNotes(prev => prev.map(n => (n.id === updated.id ? updated : n)))
  }

  const deleteNote = async (n: Note): Promise<void> => {
    if (!window.confirm(`노트 "${n.title}" 를 삭제할까요?`)) return
    await api.deleteNote(n.id)
    setNotes(prev => prev.filter(x => x.id !== n.id))
    if (selectedNote?.id === n.id) setSelectedNote(null)
  }

  const displayName = user.display_name ?? user.sub

  return (
    <div className="min-h-screen bg-slate-50 text-slate-900 antialiased flex flex-col">
      {/* Header */}
      <header className="bg-white border-b border-slate-200 px-4 py-3 flex items-center gap-3">
        <span className="inline-flex items-center gap-2">
          <span className="w-2 h-2 rounded-full bg-blue-500" />
          <span className="text-xs font-medium text-blue-700 uppercase tracking-wide">group-a</span>
        </span>
        <h1 className="font-semibold">Notebook</h1>
        <div className="flex-1" />
        <input
          type="search"
          value={searchQ}
          onChange={e => setSearchQ(e.target.value)}
          placeholder="모든 노트 검색"
          className="w-64 rounded-md border border-slate-300 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-full bg-blue-600 text-white flex items-center justify-center font-semibold uppercase text-sm">
            {displayName[0]}
          </div>
          <span className="text-sm text-slate-700">{displayName}</span>
          <button
            onClick={onLogout}
            className="ml-2 text-sm text-slate-600 hover:text-slate-900 rounded border border-slate-300 px-3 py-1.5 hover:bg-slate-100"
          >
            로그아웃
          </button>
        </div>
      </header>

      {/* Main 3-column */}
      <div className="flex-1 flex overflow-hidden">
        {/* Sidebar: notebooks */}
        <aside className="w-56 bg-white border-r border-slate-200 flex flex-col">
          <div className="px-3 py-2 flex items-center justify-between border-b border-slate-200">
            <span className="text-xs uppercase tracking-wide text-slate-500">Notebooks</span>
            <button
              onClick={addNotebook}
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
                  selectedNotebookId === nb.id ? 'bg-blue-50 text-blue-900' : 'hover:bg-slate-50'
                }`}
                onClick={() => {
                  setSelectedNotebookId(nb.id)
                  setSearchQ('')
                }}
              >
                <span className="truncate">{nb.title}</span>
                <span className="hidden group-hover:flex gap-1 text-xs text-slate-500">
                  <button onClick={e => { e.stopPropagation(); renameNotebook(nb) }} className="hover:text-slate-900">✎</button>
                  <button onClick={e => { e.stopPropagation(); deleteNotebook(nb) }} className="hover:text-red-600">✕</button>
                </span>
              </li>
            ))}
          </ul>
        </aside>

        {/* Notes list */}
        <section className="w-72 bg-white border-r border-slate-200 flex flex-col">
          <div className="px-3 py-2 flex items-center justify-between border-b border-slate-200">
            <span className="text-xs uppercase tracking-wide text-slate-500">
              {searchQ ? `검색: ${searchQ}` : '노트'}
            </span>
            <button
              onClick={addNote}
              disabled={selectedNotebookId == null || !!searchQ}
              className="text-blue-600 hover:text-blue-800 disabled:text-slate-300 text-lg leading-none"
              title="새 노트"
            >
              +
            </button>
          </div>
          <ul className="flex-1 overflow-y-auto">
            {notes.length === 0 && (
              <li className="px-3 py-4 text-sm text-slate-400">
                {selectedNotebookId == null && !searchQ ? '노트북을 선택하세요' : '노트가 없습니다'}
              </li>
            )}
            {notes.map(n => (
              <li
                key={n.id}
                className={`px-3 py-2 cursor-pointer group ${
                  selectedNote?.id === n.id ? 'bg-blue-50' : 'hover:bg-slate-50'
                }`}
                onClick={() => setSelectedNote(n)}
              >
                <div className="flex items-center justify-between">
                  <div className="text-sm font-medium truncate">{n.title}</div>
                  <button
                    onClick={e => { e.stopPropagation(); deleteNote(n) }}
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

        {/* Editor */}
        <section className="flex-1 flex flex-col bg-slate-50 overflow-hidden">
          {selectedNote ? (
            <NoteEditor key={selectedNote.id} note={selectedNote} onSave={saveNote} />
          ) : (
            <div className="flex-1 flex items-center justify-center text-slate-400 text-sm">
              왼쪽에서 노트를 선택하거나 새로 만드세요
            </div>
          )}
        </section>
      </div>
    </div>
  )
}

// ─────────────────────────────────────────────
// 노트 편집기 (debounced auto-save)
// ─────────────────────────────────────────────
function NoteEditor({
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

  // 자동 저장 (1.2초 디바운스)
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
