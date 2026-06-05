'use client'

import { useEffect, useMemo, useState } from 'react'
import { api } from '@/lib/api'
import {
  TICKET_PRIORITIES,
  TICKET_STATUSES,
  type CurrentUser,
  type TicketDetail,
  type TicketPriority,
  type TicketStatus,
  type TicketSummary,
} from '@/lib/types'

export default function Page() {
  const [user, setUser] = useState<CurrentUser | null>(null)
  const [loading, setLoading] = useState(true)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const e = params.get('error')
    if (e) {
      setErrorMsg(e)
      window.history.replaceState({}, '', '/')
    }
    // 미로그인이어도 강제 redirect 안 함 — 공개 메인 표시. user 는 헤더 "OAuth 로그인" 버튼으로 명시적 진입.
    api.me()
      .then(data => {
        if (data) setUser(data)
        setLoading(false)
      })
      .catch(() => setLoading(false))
  }, [])

  // GET /api/logout 으로 navigate → 백엔드가 IdP RP-initiated logout 체인 트리거.
  const logout = (): void => {
    window.location.href = '/api/logout'
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-slate-50 flex items-center justify-center p-4">
        <p className="text-slate-500 text-sm">로딩 중...</p>
      </div>
    )
  }

  // ?error 케이스는 LoginPage 로. 그 외엔 HelpDeskApp 공개 모드.
  if (errorMsg) return <LoginPage error={errorMsg} />
  return <HelpDeskApp user={user} onLogout={logout} />
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
            <span className="inline-block w-2 h-2 rounded-full bg-amber-500" />
            <span className="text-xs font-medium text-amber-700 uppercase tracking-wide">group-b</span>
          </div>
          <h1 className="text-2xl font-semibold">HelpDesk</h1>
          <p className="mt-1 text-sm text-slate-500">외부 고객 지원 · FastAPI + Next.js</p>
        </header>

        {error && (
          <div className="mb-4 rounded-lg bg-red-50 border border-red-200 px-3 py-2 text-sm text-red-700" role="alert">
            오류: {error}
          </div>
        )}

        <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 mb-5 text-sm text-amber-900">
          <p className="font-medium mb-1">cross-group 차단 시연</p>
          <p>
            Notebook · TaskBoard (group-a) 에 이미 로그인되어 있어도, 이 앱은 다른 그룹(group-b) 이라{' '}
            <strong>로그인 폼이 다시 표시</strong>됩니다.
          </p>
        </div>

        <a
          href="/login"
          className="block w-full text-center rounded-lg bg-amber-600 hover:bg-amber-700 active:bg-amber-800 text-white font-medium px-4 py-3 transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500"
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
const PRIORITY_COLOR: Record<TicketPriority, string> = {
  low:    'bg-slate-100 text-slate-700',
  normal: 'bg-blue-100 text-blue-700',
  high:   'bg-orange-100 text-orange-700',
  urgent: 'bg-red-100 text-red-700',
}

const STATUS_COLOR: Record<TicketStatus, string> = {
  open:     'bg-emerald-100 text-emerald-800',
  pending:  'bg-amber-100 text-amber-800',
  resolved: 'bg-slate-100 text-slate-700',
  closed:   'bg-slate-200 text-slate-500',
}

function HelpDeskApp({ user, onLogout }: { user: CurrentUser | null; onLogout: () => void }) {
  const [tickets, setTickets] = useState<TicketSummary[]>([])
  const [statusFilter, setStatusFilter] = useState<TicketStatus | ''>('')
  const [priorityFilter, setPriorityFilter] = useState<TicketPriority | ''>('')
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [detail, setDetail] = useState<TicketDetail | null>(null)
  const [showNewForm, setShowNewForm] = useState(false)

  const authed = user != null
  const displayName = user ? (user.display_name ?? user.sub) : ''

  const refresh = async (): Promise<void> => {
    const list = await api.listTickets({
      status: statusFilter || undefined,
      priority: priorityFilter || undefined,
    })
    setTickets(list)
    if (selectedId && !list.find(t => t.id === selectedId)) setSelectedId(null)
  }

  useEffect(() => {
    if (!authed) {
      setTickets([])
      setSelectedId(null)
      return
    }
    void refresh()
  }, [authed, statusFilter, priorityFilter])

  useEffect(() => {
    if (selectedId == null) {
      setDetail(null)
      return
    }
    void api.getTicket(selectedId).then(setDetail)
  }, [selectedId])

  const createTicket = async (
    subject: string,
    priority: TicketPriority,
    initialMessage: string,
  ): Promise<void> => {
    const created = await api.createTicket(subject, priority, initialMessage || undefined)
    setTickets(prev => [created, ...prev])
    setSelectedId(created.id)
    setShowNewForm(false)
  }

  const changeStatus = async (status: TicketStatus): Promise<void> => {
    if (!detail) return
    const updated = await api.updateTicket(detail.id, { status })
    setDetail({ ...detail, status: updated.status, updated_at: updated.updated_at })
    setTickets(prev => prev.map(t => (t.id === updated.id ? { ...t, status: updated.status } : t)))
  }

  const changePriority = async (priority: TicketPriority): Promise<void> => {
    if (!detail) return
    const updated = await api.updateTicket(detail.id, { priority })
    setDetail({ ...detail, priority: updated.priority, updated_at: updated.updated_at })
    setTickets(prev => prev.map(t => (t.id === updated.id ? { ...t, priority: updated.priority } : t)))
  }

  const addMessage = async (body: string): Promise<void> => {
    if (!detail) return
    const msg = await api.addMessage(detail.id, body)
    setDetail({ ...detail, messages: [...detail.messages, msg], updated_at: msg.created_at })
  }

  const deleteTicket = async (id: number): Promise<void> => {
    if (!window.confirm('티켓을 삭제할까요?')) return
    await api.deleteTicket(id)
    setTickets(prev => prev.filter(t => t.id !== id))
    if (selectedId === id) setSelectedId(null)
  }

  return (
    <div className="min-h-screen bg-slate-50 text-slate-900 antialiased flex flex-col">
      {/* Header */}
      <header className="bg-white border-b border-slate-200 px-4 py-3 flex items-center gap-3">
        <span className="inline-flex items-center gap-2">
          <span className="w-2 h-2 rounded-full bg-amber-500" />
          <span className="text-xs font-medium text-amber-700 uppercase tracking-wide">group-b</span>
        </span>
        <h1 className="font-semibold">HelpDesk</h1>
        <div className="flex-1" />
        {authed ? (
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-full bg-amber-600 text-white flex items-center justify-center font-semibold uppercase text-sm">
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
        ) : (
          <a
            href="/login"
            className="text-sm font-medium text-white bg-amber-600 hover:bg-amber-700 active:bg-amber-800 rounded px-4 py-1.5 transition-colors"
          >
            OAuth 로그인
          </a>
        )}
      </header>

      {/* Body */}
      <div className="flex-1 flex overflow-hidden">
        {/* Ticket list + filters */}
        <section className="w-96 bg-white border-r border-slate-200 flex flex-col">
          <div className="px-3 py-2 border-b border-slate-200 flex items-center gap-2">
            <select
              value={statusFilter}
              onChange={e => setStatusFilter(e.target.value as TicketStatus | '')}
              className="text-xs rounded border border-slate-300 px-2 py-1"
            >
              <option value="">모든 상태</option>
              {TICKET_STATUSES.map(s => <option key={s} value={s}>{s}</option>)}
            </select>
            <select
              value={priorityFilter}
              onChange={e => setPriorityFilter(e.target.value as TicketPriority | '')}
              className="text-xs rounded border border-slate-300 px-2 py-1"
            >
              <option value="">모든 우선순위</option>
              {TICKET_PRIORITIES.map(p => <option key={p} value={p}>{p}</option>)}
            </select>
            {authed && (
              <button
                onClick={() => setShowNewForm(true)}
                className="ml-auto text-xs rounded bg-amber-600 hover:bg-amber-700 text-white px-3 py-1"
              >
                + 새 티켓
              </button>
            )}
          </div>
          <ul className="flex-1 overflow-y-auto">
            {tickets.length === 0 && (
              <li className="px-4 py-6 text-sm text-slate-400">
                {authed ? '티켓이 없습니다' : '로그인하면 본인 티켓을 볼 수 있어요'}
              </li>
            )}
            {tickets.map(t => (
              <li
                key={t.id}
                className={`px-4 py-3 cursor-pointer border-b border-slate-100 group ${
                  selectedId === t.id ? 'bg-amber-50' : 'hover:bg-slate-50'
                }`}
                onClick={() => setSelectedId(t.id)}
              >
                <div className="flex items-center gap-2 mb-1">
                  <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${STATUS_COLOR[t.status]}`}>
                    {t.status}
                  </span>
                  <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${PRIORITY_COLOR[t.priority]}`}>
                    {t.priority}
                  </span>
                  <span className="ml-auto text-xs text-slate-400">#{t.id}</span>
                </div>
                <div className="text-sm font-medium truncate">{t.subject}</div>
                <div className="text-xs text-slate-500 mt-1 flex items-center justify-between">
                  <span>{t.message_count} 메시지</span>
                  <span>{new Date(t.updated_at).toLocaleString('ko-KR')}</span>
                </div>
              </li>
            ))}
          </ul>
        </section>

        {/* Detail */}
        <section className="flex-1 flex flex-col overflow-hidden">
          {showNewForm && authed ? (
            <NewTicketForm onCancel={() => setShowNewForm(false)} onCreate={createTicket} />
          ) : detail && user ? (
            <TicketDetailPane
              detail={detail}
              currentSub={user.sub}
              onChangeStatus={changeStatus}
              onChangePriority={changePriority}
              onAddMessage={addMessage}
              onDelete={() => deleteTicket(detail.id)}
            />
          ) : (
            <div className="flex-1 flex items-center justify-center text-slate-400 text-sm text-center px-6">
              {authed
                ? '왼쪽에서 티켓을 선택하거나 새로 만드세요'
                : '오른쪽 상단 "OAuth 로그인" 으로 들어가면 티켓을 등록할 수 있어요'}
            </div>
          )}
        </section>
      </div>
    </div>
  )
}

// ─────────────────────────────────────────────
// 새 티켓 폼
// ─────────────────────────────────────────────
function NewTicketForm({
  onCreate,
  onCancel,
}: {
  onCreate: (subject: string, priority: TicketPriority, message: string) => Promise<void>
  onCancel: () => void
}) {
  const [subject, setSubject] = useState('')
  const [priority, setPriority] = useState<TicketPriority>('normal')
  const [message, setMessage] = useState('')

  return (
    <div className="p-8 max-w-2xl">
      <h2 className="text-xl font-semibold mb-4">새 티켓</h2>
      <form
        onSubmit={async e => {
          e.preventDefault()
          if (!subject.trim()) return
          await onCreate(subject.trim(), priority, message.trim())
        }}
        className="space-y-4"
      >
        <div>
          <label className="block text-sm font-medium mb-1">제목</label>
          <input
            value={subject}
            onChange={e => setSubject(e.target.value)}
            required
            className="w-full rounded border border-slate-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-amber-500"
            placeholder="어떤 도움이 필요하신가요?"
          />
        </div>
        <div>
          <label className="block text-sm font-medium mb-1">우선순위</label>
          <select
            value={priority}
            onChange={e => setPriority(e.target.value as TicketPriority)}
            className="rounded border border-slate-300 px-3 py-2"
          >
            {TICKET_PRIORITIES.map(p => <option key={p} value={p}>{p}</option>)}
          </select>
        </div>
        <div>
          <label className="block text-sm font-medium mb-1">메시지 (선택)</label>
          <textarea
            value={message}
            onChange={e => setMessage(e.target.value)}
            rows={6}
            className="w-full rounded border border-slate-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-amber-500"
            placeholder="자세한 설명을 적어주세요"
          />
        </div>
        <div className="flex gap-2">
          <button
            type="submit"
            className="rounded bg-amber-600 hover:bg-amber-700 text-white px-4 py-2 text-sm font-medium"
          >
            생성
          </button>
          <button
            type="button"
            onClick={onCancel}
            className="rounded border border-slate-300 hover:bg-slate-100 text-slate-700 px-4 py-2 text-sm"
          >
            취소
          </button>
        </div>
      </form>
    </div>
  )
}

// ─────────────────────────────────────────────
// 티켓 상세
// ─────────────────────────────────────────────
function TicketDetailPane({
  detail,
  currentSub,
  onChangeStatus,
  onChangePriority,
  onAddMessage,
  onDelete,
}: {
  detail: TicketDetail
  currentSub: string
  onChangeStatus: (s: TicketStatus) => Promise<void>
  onChangePriority: (p: TicketPriority) => Promise<void>
  onAddMessage: (body: string) => Promise<void>
  onDelete: () => Promise<void>
}) {
  const [draft, setDraft] = useState('')

  const ordered = useMemo(
    () => [...detail.messages].sort((a, b) => a.created_at.localeCompare(b.created_at)),
    [detail.messages],
  )

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      <header className="px-6 py-4 border-b border-slate-200 bg-white">
        <div className="flex items-center gap-2 mb-2">
          <span className="text-xs text-slate-400">#{detail.id}</span>
          <select
            value={detail.status}
            onChange={e => onChangeStatus(e.target.value as TicketStatus)}
            className={`text-xs px-2 py-1 rounded-full font-medium ${STATUS_COLOR[detail.status]}`}
          >
            {TICKET_STATUSES.map(s => <option key={s} value={s}>{s}</option>)}
          </select>
          <select
            value={detail.priority}
            onChange={e => onChangePriority(e.target.value as TicketPriority)}
            className={`text-xs px-2 py-1 rounded-full font-medium ${PRIORITY_COLOR[detail.priority]}`}
          >
            {TICKET_PRIORITIES.map(p => <option key={p} value={p}>{p}</option>)}
          </select>
          <div className="flex-1" />
          <button
            onClick={onDelete}
            className="text-xs text-slate-500 hover:text-red-600 rounded border border-slate-300 hover:border-red-300 px-2 py-1"
          >
            삭제
          </button>
        </div>
        <h2 className="text-xl font-semibold">{detail.subject}</h2>
        <p className="text-xs text-slate-400 mt-1">
          생성 {new Date(detail.created_at).toLocaleString('ko-KR')} · 갱신{' '}
          {new Date(detail.updated_at).toLocaleString('ko-KR')}
        </p>
      </header>

      <div className="flex-1 overflow-y-auto px-6 py-4 space-y-3">
        {ordered.length === 0 && (
          <p className="text-sm text-slate-400">아직 메시지가 없습니다</p>
        )}
        {ordered.map(m => {
          const mine = m.author_sub === currentSub
          return (
            <div key={m.id} className={`flex ${mine ? 'justify-end' : 'justify-start'}`}>
              <div
                className={`max-w-2xl rounded-2xl px-4 py-2 ${
                  mine ? 'bg-amber-100 text-amber-900' : 'bg-white border border-slate-200'
                }`}
              >
                <div className="text-xs text-slate-500 mb-1">
                  {m.author_sub} · {new Date(m.created_at).toLocaleString('ko-KR')}
                </div>
                <div className="text-sm whitespace-pre-wrap">{m.body}</div>
              </div>
            </div>
          )
        })}
      </div>

      <form
        onSubmit={async e => {
          e.preventDefault()
          if (!draft.trim()) return
          await onAddMessage(draft.trim())
          setDraft('')
        }}
        className="border-t border-slate-200 bg-white px-4 py-3 flex gap-2"
      >
        <input
          value={draft}
          onChange={e => setDraft(e.target.value)}
          placeholder="메시지 추가..."
          className="flex-1 rounded border border-slate-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-amber-500"
        />
        <button
          type="submit"
          className="rounded bg-amber-600 hover:bg-amber-700 text-white px-4 py-2 text-sm font-medium"
        >
          보내기
        </button>
      </form>
    </div>
  )
}
