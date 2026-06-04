export interface CurrentUser {
  sub: string
  client_id: string
  scope?: string
  display_name?: string
}

export interface Board {
  _id: string
  ownerSub: string
  title: string
  createdAt: string
  updatedAt: string
}

export type TaskColumn = 'todo' | 'doing' | 'done'

export interface Task {
  _id: string
  boardId: string
  ownerSub: string
  title: string
  column: TaskColumn
  position: number
  done: boolean
  createdAt: string
  updatedAt: string
}
