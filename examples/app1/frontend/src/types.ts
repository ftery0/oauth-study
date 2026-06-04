export interface CurrentUser {
  sub: string
  client_id: string
  scope?: string
  display_name?: string
}

export interface Notebook {
  id: number
  title: string
  createdAt: string
  updatedAt: string
}

export interface Note {
  id: number
  notebookId: number
  title: string
  bodyMd: string
  createdAt: string
  updatedAt: string
}
