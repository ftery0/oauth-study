export interface CurrentUser {
  sub: string
  client_id: string
  scope?: string
  display_name?: string
}

export type TicketStatus = 'open' | 'pending' | 'resolved' | 'closed'
export type TicketPriority = 'low' | 'normal' | 'high' | 'urgent'

export const TICKET_STATUSES: TicketStatus[] = ['open', 'pending', 'resolved', 'closed']
export const TICKET_PRIORITIES: TicketPriority[] = ['low', 'normal', 'high', 'urgent']

export interface TicketSummary {
  id: number
  subject: string
  status: TicketStatus
  priority: TicketPriority
  created_at: string
  updated_at: string
  message_count: number
}

export interface TicketMessage {
  id: number
  author_sub: string
  body: string
  created_at: string
}

export interface TicketDetail extends TicketSummary {
  messages: TicketMessage[]
}
