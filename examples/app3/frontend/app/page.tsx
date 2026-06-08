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
import {
  STATUS_LABEL_KO,
  PRIORITY_LABEL_KO,
  absoluteKo,
  relativeKo,
} from '@/lib/format'

// ─────────────────────────────────────────────
// 진입점
// ─────────────────────────────────────────────
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
    api.me()
      .then(data => {
        if (data) setUser(data)
        setLoading(false)
      })
      .catch(() => setLoading(false))
  }, [])

  const logout = (): void => {
    window.location.href = '/api/logout'
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-50">
        <p className="text-slate-500 text-sm">워크스페이스 준비 중…</p>
      </div>
    )
  }

  if (errorMsg) return <LoginPage error={errorMsg} />
  return <HelpDeskApp user={user} onLogout={logout} />
}

// ─────────────────────────────────────────────
// 로그인 화면 (group-b cross-group 시연)
// ─────────────────────────────────────────────
function LoginPage({ error }: { error: string | null }) {
  return (
    <div className="min-h-screen flex bg-slate-50">
      <aside className="hidden md:flex w-1/2 bg-navy-900 text-slate-100 p-12 flex-col justify-between">
        <div>
          <div className="inline-flex items-center gap-2 mb-6">
            <span className="w-2 h-2 rounded-full bg-teal-400" />
            <span className="text-[11px] font-semibold uppercase tracking-[0.18em] text-teal-300">
              group · b
            </span>
          </div>
          <h1 className="text-4xl font-bold tracking-tight mb-3">HelpDesk</h1>
          <p className="text-slate-400 text-base leading-relaxed">
            고객의 문의를 한 곳에서 받아 응답을 빠르고 정확하게 정리합니다.
          </p>
        </div>
        <dl className="space-y-3 text-sm text-slate-400">
          <div className="flex gap-3">
            <dt className="text-teal-400 font-semibold w-24 shrink-0">상태 흐름</dt>
            <dd>응답 대기 → 처리 중 → 해결 → 종료</dd>
          </div>
          <div className="flex gap-3">
            <dt className="text-teal-400 font-semibold w-24 shrink-0">우선순위</dt>
            <dd>낮음 · 보통 · 높음 · 긴급</dd>
          </div>
          <div className="flex gap-3">
            <dt className="text-teal-400 font-semibold w-24 shrink-0">스택</dt>
            <dd>FastAPI · Next.js · PostgreSQL</dd>
          </div>
        </dl>
      </aside>

      <main className="flex-1 flex items-center justify-center p-6">
        <div className="w-full max-w-md">
          <h2 className="text-2xl font-semibold text-slate-900 mb-2">에이전트로 로그인</h2>
          <p className="text-slate-500 mb-8">
            이 워크스페이스는 별도 그룹(group-b)입니다.
            Notebook · TaskBoard 로그인 세션이 있어도 다시 인증해야 합니다.
          </p>

          {error && (
            <div className="mb-5 rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800" role="alert">
              <span className="font-semibold">오류 · </span>{error}
            </div>
          )}

          <div className="rounded-md border border-amber-200 bg-amber-50/70 px-4 py-3 mb-6 text-sm text-amber-900">
            <p className="font-semibold mb-1">cross-group 인증 분리</p>
            <p className="text-amber-800">
              다른 그룹의 SSO 세션은 자동 적용되지 않습니다.
              조직 정책에 따라 별도 자격증명이 필요합니다.
            </p>
          </div>

          <a
            href="/login"
            className="block w-full text-center rounded-md bg-teal-700 hover:bg-teal-800 text-white font-semibold px-4 py-3 transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-teal-600"
          >
            OAuth 로 인증 후 입장
          </a>
        </div>
      </main>
    </div>
  )
}

// ─────────────────────────────────────────────
// 색상 매핑
// ─────────────────────────────────────────────
const PRIORITY_STYLE: Record<TicketPriority, string> = {
  low:    'bg-slate-100 text-slate-700 ring-1 ring-slate-200',
  normal: 'bg-sky-100 text-sky-800 ring-1 ring-sky-200',
  high:   'bg-orange-100 text-orange-800 ring-1 ring-orange-200',
  urgent: 'bg-red-100 text-red-800 ring-1 ring-red-200',
}

const PRIORITY_DOT: Record<TicketPriority, string> = {
  low: 'bg-slate-400',
  normal: 'bg-sky-500',
  high: 'bg-orange-500',
  urgent: 'bg-red-500',
}

const STATUS_STYLE: Record<TicketStatus, string> = {
  open:     'bg-teal-100 text-teal-800 ring-1 ring-teal-200',
  pending:  'bg-amber-100 text-amber-800 ring-1 ring-amber-200',
  resolved: 'bg-slate-100 text-slate-700 ring-1 ring-slate-200',
  closed:   'bg-slate-200 text-slate-500 ring-1 ring-slate-300',
}

// ─────────────────────────────────────────────
// 메인 워크스페이스
// ─────────────────────────────────────────────
function HelpDeskApp({ user, onLogout }: { user: CurrentUser | null; onLogout: () => void }) {
  const [tickets, setTickets] = useState<TicketSummary[]>([])
  const [statusFilter, setStatusFilter] = useState<TicketStatus | 'all'>('all')
  const [priorityFilter, setPriorityFilter] = useState<TicketPriority | ''>('')
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [detail, setDetail] = useState<TicketDetail | null>(null)
  const [showNewForm, setShowNewForm] = useState(false)

  const authed = user != null
  const displayName = user ? (user.display_name ?? user.sub) : ''

  const refresh = async (): Promise<void> => {
    const list = await api.listTickets({
      status: statusFilter === 'all' ? undefined : statusFilter,
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
    if (!window.confirm('이 티켓을 영구 삭제하시겠습니까?')) return
    await api.deleteTicket(id)
    setTickets(prev => prev.filter(t => t.id !== id))
    if (selectedId === id) setSelectedId(null)
  }

  const counts = useMemo(() => {
    const c: Record<TicketStatus | 'all', number> = {
      all: tickets.length, open: 0, pending: 0, resolved: 0, closed: 0,
    }
    for (const t of tickets) c[t.status]++
    return c
  }, [tickets])

  return (
    <div className="min-h-screen flex bg-slate-50 text-slate-900">
      {/* Left icon rail */}
      <nav className="w-14 bg-navy-900 flex flex-col items-center py-4 gap-1 shrink-0">
        <div className="w-9 h-9 rounded-md bg-teal-600 text-white flex items-center justify-center font-bold text-sm mb-3">
          HD
        </div>
        <RailButton active title="문의함">
          <InboxIcon />
        </RailButton>
        <RailButton title="고객" disabled>
          <UsersIcon />
        </RailButton>
        <RailButton title="리포트" disabled>
          <ChartIcon />
        </RailButton>
        <div className="flex-1" />
        <RailButton title="설정" disabled>
          <GearIcon />
        </RailButton>
      </nav>

      {/* Main */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Top bar */}
        <header className="bg-white border-b border-slate-200 px-5 h-12 flex items-center gap-3 shrink-0">
          <div className="flex items-center gap-2">
            <h1 className="font-semibold text-slate-900">HelpDesk</h1>
            <span className="text-[10px] font-semibold uppercase tracking-[0.18em] text-teal-700 bg-teal-50 px-1.5 py-0.5 rounded">
              group · b
            </span>
          </div>
          <span className="text-slate-300">›</span>
          <span className="text-slate-500 text-xs">문의함</span>
          <div className="flex-1" />
          {authed ? (
            <div className="flex items-center gap-2">
              <div className="w-7 h-7 rounded-full bg-teal-700 text-white flex items-center justify-center font-semibold uppercase text-xs">
                {displayName[0]}
              </div>
              <span className="text-slate-700 text-xs hidden sm:inline">{displayName}</span>
              <button
                onClick={onLogout}
                className="ml-2 text-xs text-slate-500 hover:text-slate-900 border border-slate-200 hover:border-slate-300 px-2.5 py-1 rounded"
              >
                로그아웃
              </button>
            </div>
          ) : (
            <a
              href="/login"
              className="text-xs font-semibold text-white bg-teal-700 hover:bg-teal-800 rounded px-3 py-1.5"
            >
              로그인
            </a>
          )}
        </header>

        {/* Status filter strip */}
        <div className="bg-white border-b border-slate-200 px-5 flex items-center gap-1 overflow-x-auto shrink-0">
          {(['all', ...TICKET_STATUSES] as const).map(s => {
            const active = statusFilter === s
            return (
              <button
                key={s}
                onClick={() => setStatusFilter(s)}
                className={`relative shrink-0 px-3 py-2.5 text-xs font-semibold transition-colors border-b-2 -mb-px ${
                  active
                    ? 'text-teal-700 border-teal-600'
                    : 'text-slate-500 hover:text-slate-900 border-transparent'
                }`}
              >
                {s === 'all' ? '전체' : STATUS_LABEL_KO[s]}
                <span className={`ml-2 inline-block min-w-[1.25rem] text-center px-1 rounded text-[10px] ${
                  active ? 'bg-teal-100 text-teal-700' : 'bg-slate-100 text-slate-500'
                }`}>
                  {counts[s]}
                </span>
              </button>
            )
          })}
          <div className="flex-1" />
          <select
            value={priorityFilter}
            onChange={e => setPriorityFilter(e.target.value as TicketPriority | '')}
            className="shrink-0 text-xs rounded border border-slate-200 px-2 py-1 text-slate-700"
          >
            <option value="">모든 우선순위</option>
            {TICKET_PRIORITIES.map(p => (
              <option key={p} value={p}>{PRIORITY_LABEL_KO[p]}</option>
            ))}
          </select>
          {authed && (
            <button
              onClick={() => setShowNewForm(true)}
              className="shrink-0 ml-2 text-xs font-semibold rounded bg-teal-700 hover:bg-teal-800 text-white px-3 py-1.5"
            >
              + 문의 등록
            </button>
          )}
        </div>

        {/* Two-pane body */}
        <div className="flex-1 flex overflow-hidden">
          {/* Ticket queue */}
          <section className="w-96 border-r border-slate-200 bg-white flex flex-col shrink-0">
            <ul className="flex-1 overflow-y-auto divide-y divide-slate-100">
              {tickets.length === 0 && (
                <li className="px-5 py-8 text-center text-slate-400 text-sm">
                  {authed ? '처리할 티켓이 없습니다.' : '에이전트로 로그인하면 본인 큐가 표시됩니다.'}
                </li>
              )}
              {tickets.map(t => {
                const sel = selectedId === t.id
                return (
                  <li
                    key={t.id}
                    onClick={() => setSelectedId(t.id)}
                    className={`relative px-4 py-3 cursor-pointer transition-colors ${
                      sel ? 'bg-teal-50/70' : 'hover:bg-slate-50'
                    }`}
                  >
                    {sel && <span className="absolute left-0 top-0 bottom-0 w-1 bg-teal-600" />}
                    <div className="flex items-center gap-2 mb-1.5">
                      <span className={`w-1.5 h-1.5 rounded-full ${PRIORITY_DOT[t.priority]}`} title={PRIORITY_LABEL_KO[t.priority]} />
                      <span className="font-mono text-[10px] text-slate-400">#{t.id.toString().padStart(4, '0')}</span>
                      <span className={`tag-status text-[10px] font-semibold px-1.5 py-0.5 rounded ${STATUS_STYLE[t.status]}`}>
                        {STATUS_LABEL_KO[t.status]}
                      </span>
                    </div>
                    <div className="text-sm font-medium text-slate-900 truncate">{t.subject}</div>
                    <div className="text-[11px] text-slate-500 mt-1 flex items-center justify-between">
                      <span>{t.message_count} 메시지</span>
                      <span>{relativeKo(t.updated_at)}</span>
                    </div>
                  </li>
                )
              })}
            </ul>
          </section>

          {/* Detail pane */}
          <section className="flex-1 flex flex-col overflow-hidden min-w-0">
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
              <div className="flex-1 flex items-center justify-center p-8">
                <div className="text-center text-slate-400 max-w-sm">
                  <div className="text-5xl mb-4 text-slate-200">✉</div>
                  <p className="text-sm">
                    {authed
                      ? '왼쪽 큐에서 티켓을 선택하거나, 우측 상단에서 새 문의를 등록하세요.'
                      : '에이전트로 로그인하면 티켓 큐와 상세 응대 화면이 열립니다.'}
                  </p>
                </div>
              </div>
            )}
          </section>
        </div>
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
    <div className="flex-1 overflow-y-auto">
      <div className="max-w-2xl mx-auto p-8">
        <h2 className="text-xl font-semibold text-slate-900 mb-1">새 문의 등록</h2>
        <p className="text-slate-500 mb-6">고객의 문의 내용을 정리해 큐에 추가합니다.</p>
        <form
          onSubmit={async e => {
            e.preventDefault()
            if (!subject.trim()) return
            await onCreate(subject.trim(), priority, message.trim())
          }}
          className="space-y-5 bg-white border border-slate-200 rounded-lg p-6"
        >
          <div>
            <label className="block text-xs font-semibold text-slate-600 uppercase tracking-wider mb-1.5">
              제목 <span className="text-red-500">*</span>
            </label>
            <input
              value={subject}
              onChange={e => setSubject(e.target.value)}
              required
              className="w-full rounded border border-slate-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-teal-600/30 focus:border-teal-600"
              placeholder="문의 제목을 적어주세요"
            />
          </div>
          <div>
            <label className="block text-xs font-semibold text-slate-600 uppercase tracking-wider mb-1.5">
              우선순위
            </label>
            <select
              value={priority}
              onChange={e => setPriority(e.target.value as TicketPriority)}
              className="rounded border border-slate-300 px-3 py-2 text-sm"
            >
              {TICKET_PRIORITIES.map(p => (
                <option key={p} value={p}>{PRIORITY_LABEL_KO[p]}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-xs font-semibold text-slate-600 uppercase tracking-wider mb-1.5">
              상황 설명 (선택)
            </label>
            <textarea
              value={message}
              onChange={e => setMessage(e.target.value)}
              rows={6}
              className="w-full rounded border border-slate-300 px-3 py-2 text-sm leading-relaxed focus:outline-none focus:ring-2 focus:ring-teal-600/30 focus:border-teal-600"
              placeholder="고객이 겪은 상황을 자세히 적어주세요"
            />
          </div>
          <div className="flex justify-end gap-2 pt-2 border-t border-slate-100">
            <button
              type="button"
              onClick={onCancel}
              className="rounded border border-slate-300 hover:bg-slate-50 text-slate-700 px-4 py-2 text-sm font-medium"
            >
              취소
            </button>
            <button
              type="submit"
              className="rounded bg-teal-700 hover:bg-teal-800 text-white px-4 py-2 text-sm font-semibold"
            >
              티켓 생성
            </button>
          </div>
        </form>
      </div>
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
    <div className="flex-1 flex flex-col overflow-hidden bg-slate-50">
      {/* Ticket header w/ meta */}
      <header className="bg-white border-b border-slate-200 px-6 py-4 shrink-0">
        <div className="flex items-center gap-2 mb-2">
          <span className="font-mono text-[11px] text-slate-400">
            티켓 #{detail.id.toString().padStart(4, '0')}
          </span>
          <span className="text-slate-300">·</span>
          <span className="text-[11px] text-slate-500">
            등록 {absoluteKo(detail.created_at)}
          </span>
          <span className="text-slate-300">·</span>
          <span className="text-[11px] text-slate-500">
            최근 활동 {relativeKo(detail.updated_at)}
          </span>
          <div className="flex-1" />
          <button
            onClick={onDelete}
            className="text-[11px] text-slate-400 hover:text-red-600 px-2 py-1"
          >
            삭제
          </button>
        </div>
        <h2 className="text-lg font-semibold text-slate-900 mb-3">{detail.subject}</h2>
        <div className="flex flex-wrap items-center gap-2">
          <MetaField label="상태">
            <select
              value={detail.status}
              onChange={e => onChangeStatus(e.target.value as TicketStatus)}
              className={`text-[11px] font-semibold px-2 py-1 rounded ${STATUS_STYLE[detail.status]} cursor-pointer focus:outline-none`}
            >
              {TICKET_STATUSES.map(s => (
                <option key={s} value={s}>{STATUS_LABEL_KO[s]}</option>
              ))}
            </select>
          </MetaField>
          <MetaField label="우선순위">
            <select
              value={detail.priority}
              onChange={e => onChangePriority(e.target.value as TicketPriority)}
              className={`text-[11px] font-semibold px-2 py-1 rounded ${PRIORITY_STYLE[detail.priority]} cursor-pointer focus:outline-none`}
            >
              {TICKET_PRIORITIES.map(p => (
                <option key={p} value={p}>{PRIORITY_LABEL_KO[p]}</option>
              ))}
            </select>
          </MetaField>
          <MetaField label="메시지">
            <span className="text-xs text-slate-700">{ordered.length} 건</span>
          </MetaField>
        </div>
      </header>

      {/* Chat */}
      <div className="flex-1 overflow-y-auto px-6 py-5 space-y-4">
        {ordered.length === 0 && (
          <div className="text-center text-slate-400 text-sm py-8">
            아직 응답이 없습니다. 첫 응답을 작성해 보세요.
          </div>
        )}
        {ordered.map(m => {
          const mine = m.author_sub === currentSub
          return (
            <div key={m.id} className={`flex ${mine ? 'justify-end' : 'justify-start'}`}>
              <div className={`max-w-2xl ${mine ? 'items-end' : 'items-start'} flex flex-col gap-1`}>
                <div className="text-[10px] text-slate-500 px-1">
                  {mine ? '나' : m.author_sub.slice(0, 12)} · {absoluteKo(m.created_at)}
                </div>
                <div
                  className={`px-4 py-2.5 text-sm leading-relaxed whitespace-pre-wrap shadow-sm ${
                    mine
                      ? 'bg-teal-700 text-white rounded-2xl rounded-tr-sm'
                      : 'bg-white text-slate-800 border border-slate-200 rounded-2xl rounded-tl-sm'
                  }`}
                >
                  {m.body}
                </div>
              </div>
            </div>
          )
        })}
      </div>

      {/* Reply box */}
      <form
        onSubmit={async e => {
          e.preventDefault()
          if (!draft.trim()) return
          await onAddMessage(draft.trim())
          setDraft('')
        }}
        className="border-t border-slate-200 bg-white px-6 py-4 shrink-0"
      >
        <div className="flex items-start gap-3">
          <textarea
            value={draft}
            onChange={e => setDraft(e.target.value)}
            placeholder="응답을 입력하세요…  (Cmd/Ctrl + Enter 전송)"
            rows={2}
            onKeyDown={async e => {
              if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
                e.preventDefault()
                if (!draft.trim()) return
                await onAddMessage(draft.trim())
                setDraft('')
              }
            }}
            className="flex-1 rounded border border-slate-300 px-3 py-2 text-sm leading-relaxed focus:outline-none focus:ring-2 focus:ring-teal-600/30 focus:border-teal-600 resize-none"
          />
          <button
            type="submit"
            className="rounded bg-teal-700 hover:bg-teal-800 disabled:bg-slate-300 text-white px-4 py-2 text-sm font-semibold self-stretch"
            disabled={!draft.trim()}
          >
            응답 전송
          </button>
        </div>
      </form>
    </div>
  )
}

// ─────────────────────────────────────────────
// 메타 필드 (라벨 + 값)
// ─────────────────────────────────────────────
function MetaField({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="inline-flex items-center gap-1.5">
      <span className="text-[10px] uppercase tracking-wider text-slate-400 font-semibold">{label}</span>
      {children}
    </div>
  )
}

// ─────────────────────────────────────────────
// Rail
// ─────────────────────────────────────────────
function RailButton({
  active,
  disabled,
  title,
  children,
}: {
  active?: boolean
  disabled?: boolean
  title: string
  children: React.ReactNode
}) {
  return (
    <button
      title={title}
      disabled={disabled}
      className={`w-10 h-10 flex items-center justify-center rounded-md transition-colors ${
        active
          ? 'bg-navy-800 text-teal-300'
          : disabled
            ? 'text-slate-600 cursor-default'
            : 'text-slate-400 hover:bg-navy-800 hover:text-white'
      }`}
    >
      {children}
    </button>
  )
}

function InboxIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <polyline points="22 12 16 12 14 15 10 15 8 12 2 12" />
      <path d="M5.45 5.11 2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z" />
    </svg>
  )
}

function UsersIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
      <circle cx="9" cy="7" r="4" />
      <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
      <path d="M16 3.13a4 4 0 0 1 0 7.75" />
    </svg>
  )
}

function ChartIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <line x1="18" y1="20" x2="18" y2="10" />
      <line x1="12" y1="20" x2="12" y2="4" />
      <line x1="6" y1="20" x2="6" y2="14" />
    </svg>
  )
}

function GearIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="3" />
      <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" />
    </svg>
  )
}
