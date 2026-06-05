import { useNotebooks } from '../hooks/useNotebooks'
import { useNotes } from '../hooks/useNotes'
import { NotebookSidebar } from '../components/NotebookSidebar'
import { NoteList } from '../components/NoteList'
import { NoteEditor } from '../components/NoteEditor'
import { UserMenu } from '../components/UserMenu'
import type { CurrentUser } from '../lib/types'

export function NotebookPage({
  user,
  onProfile,
  onLogout,
}: {
  user: CurrentUser
  onProfile: () => void
  onLogout: () => void
}) {
  const {
    notebooks,
    selectedNotebookId,
    setSelectedNotebookId,
    addNotebook,
    renameNotebook,
    deleteNotebook,
  } = useNotebooks()

  const {
    notes,
    selectedNote,
    setSelectedNote,
    searchQ,
    setSearchQ,
    addNote,
    saveNote,
    deleteNote,
  } = useNotes(selectedNotebookId)

  const selectNotebook = (id: number) => {
    setSelectedNotebookId(id)
    setSearchQ('')
  }

  return (
    <div className="min-h-screen bg-slate-50 text-slate-900 antialiased flex flex-col">
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
        <UserMenu user={user} onProfile={onProfile} onLogout={onLogout} />
      </header>

      <div className="flex-1 flex overflow-hidden">
        <NotebookSidebar
          notebooks={notebooks}
          selectedId={selectedNotebookId}
          onSelect={selectNotebook}
          onAdd={addNotebook}
          onRename={renameNotebook}
          onDelete={deleteNotebook}
        />
        <NoteList
          notes={notes}
          selectedNote={selectedNote}
          onSelect={setSelectedNote}
          onAdd={addNote}
          onDelete={deleteNote}
          searchQ={searchQ}
          notebookSelected={selectedNotebookId != null}
        />
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
