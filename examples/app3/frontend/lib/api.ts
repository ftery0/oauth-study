import type {
  CurrentUser,
  TicketDetail,
  TicketMessage,
  TicketPriority,
  TicketStatus,
  TicketSummary,
} from './types'

async function json<T>(input: string, init?: RequestInit): Promise<T> {
  const r = await fetch(input, {
    cache: 'no-store',
    ...init,
    headers: { 'Content-Type': 'application/json', ...(init?.headers ?? {}) },
  })
  if (!r.ok) throw new Error(`${r.status} ${r.statusText}`)
  if (r.status === 204) return undefined as T
  return r.json() as Promise<T>
}

export const api = {
  me: async (): Promise<CurrentUser | null> => {
    const r = await fetch('/api/me', { cache: 'no-store' })
    return r.ok ? ((await r.json()) as CurrentUser) : null
  },
  logout: () => fetch('/api/logout', { method: 'POST', cache: 'no-store' }),

  listTickets: (filters: { status?: TicketStatus; priority?: TicketPriority } = {}) => {
    const qs = new URLSearchParams()
    if (filters.status) qs.set('status', filters.status)
    if (filters.priority) qs.set('priority', filters.priority)
    return json<TicketSummary[]>(`/api/tickets${qs.toString() ? `?${qs}` : ''}`)
  },
  getTicket: (id: number) => json<TicketDetail>(`/api/tickets/${id}`),
  createTicket: (subject: string, priority: TicketPriority, initial_message?: string) =>
    json<TicketSummary>('/api/tickets', {
      method: 'POST',
      body: JSON.stringify({ subject, priority, initial_message }),
    }),
  updateTicket: (id: number, patch: { status?: TicketStatus; priority?: TicketPriority }) =>
    json<TicketSummary>(`/api/tickets/${id}`, { method: 'PATCH', body: JSON.stringify(patch) }),
  deleteTicket: (id: number) =>
    json<void>(`/api/tickets/${id}`, { method: 'DELETE' }),
  addMessage: (id: number, body: string) =>
    json<TicketMessage>(`/api/tickets/${id}/messages`, {
      method: 'POST',
      body: JSON.stringify({ body }),
    }),
}
