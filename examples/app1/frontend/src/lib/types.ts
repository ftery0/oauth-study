export interface CurrentUser {
  sub: string
  preferred_username?: string
  name?: string
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

// 표시명 우선순위: name → preferred_username → sub.
export function pickDisplayName(u: CurrentUser): string {
  return u.name?.trim() || u.preferred_username?.trim() || u.sub
}
