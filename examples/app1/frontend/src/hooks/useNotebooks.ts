import { useEffect, useState } from 'react'
import { api } from '../lib/api'
import type { Notebook } from '../lib/types'

// 노트북 목록 + 선택 + CRUD.
// 최초 1회 목록 로드, 첫 번째 노트북 자동 선택.
export function useNotebooks() {
  const [notebooks, setNotebooks] = useState<Notebook[]>([])
  const [selectedNotebookId, setSelectedNotebookId] = useState<number | null>(null)

  useEffect(() => {
    api.listNotebooks().then(list => {
      setNotebooks(list)
      if (list.length > 0) setSelectedNotebookId(list[0].id)
    })
  }, [])

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

  return {
    notebooks,
    selectedNotebookId,
    setSelectedNotebookId,
    addNotebook,
    renameNotebook,
    deleteNotebook,
  }
}
