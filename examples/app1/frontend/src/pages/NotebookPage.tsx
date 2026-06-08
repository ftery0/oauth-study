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
    <div className="min-h-screen bg-paper-100 text-ink-900 flex flex-col">
      <header className="bg-paper-50/80 backdrop-blur border-b border-paper-200 px-6 py-3 flex items-center gap-4">
        <div className="flex items-baseline gap-3">
          <h1 className="font-serif text-xl text-ink-900">Notebook</h1>
          <span className="text-[10px] uppercase tracking-[0.18em] text-emerald-800/70 hidden sm:inline">
            · group a
          </span>
        </div>
        <div className="flex-1" />
        <div className="relative">
          <span className="absolute left-3 top-1/2 -translate-y-1/2 text-ink-400 text-sm">⌕</span>
          <input
            type="search"
            value={searchQ}
            onChange={e => setSearchQ(e.target.value)}
            placeholder="모든 노트에서 찾기"
            className="w-72 rounded-md bg-paper-100 border border-paper-200 pl-8 pr-3 py-1.5 text-sm italic placeholder:text-ink-400 focus:outline-none focus:ring-2 focus:ring-emerald-700/40 focus:border-emerald-700/40"
          />
        </div>
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
        <section className="flex-1 flex flex-col bg-paper-50 overflow-hidden">
          {selectedNote ? (
            <NoteEditor key={selectedNote.id} note={selectedNote} onSave={saveNote} />
          ) : (
            <div className="flex-1 flex items-center justify-center text-ink-400 text-sm italic px-8 text-center">
              왼쪽 목록에서 노트를 골라 펴거나, 새 노트를 적어보세요.
            </div>
          )}
        </section>
      </div>
    </div>
  )
}
