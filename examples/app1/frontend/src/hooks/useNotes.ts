import { useEffect, useState } from 'react'
import { api } from '../lib/api'
import type { Note } from '../lib/types'

// 선택된 노트북의 노트 목록 + 선택 + 검색(디바운스) + CRUD.
//
// 검색 / 노트북 전환의 race condition 을 한 effect 에 통합해서 처리한다:
//   - searchQ 비어있으면 selectedNotebookId 의 노트 목록을 즉시 fetch
//   - searchQ 있으면 200ms 디바운스 후 검색 fetch
//   - 어느 의존성이 바뀌어도 이전 in-flight 요청은 AbortController 로 취소 → 응답 순서 뒤바뀜 무시
export function useNotes(selectedNotebookId: number | null) {
  const [notes, setNotes] = useState<Note[]>([])
  const [selectedNote, setSelectedNote] = useState<Note | null>(null)
  const [searchQ, setSearchQ] = useState('')

  useEffect(() => {
    const q = searchQ.trim()

    if (!q && selectedNotebookId == null) {
      setNotes([])
      setSelectedNote(null)
      return
    }

    const ac = new AbortController()
    const fetcher = q
      ? () => api.searchNotes(q, ac.signal)
      : () => api.listNotes(selectedNotebookId!, ac.signal)

    const handle = setTimeout(() => {
      fetcher()
        .then(list => {
          setNotes(list)
          setSelectedNote(list[0] ?? null)
        })
        .catch(err => {
          if (err?.name !== 'AbortError') throw err
        })
    }, q ? 200 : 0)

    return () => {
      clearTimeout(handle)
      ac.abort()
    }
  }, [selectedNotebookId, searchQ])

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

  return {
    notes,
    selectedNote,
    setSelectedNote,
    searchQ,
    setSearchQ,
    addNote,
    saveNote,
    deleteNote,
  }
}
