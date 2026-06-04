import type { Request, Response, NextFunction } from 'express'

export function requireUser(req: Request, res: Response, next: NextFunction): void {
  if (!req.session.userSub) {
    res.status(401).json({ error: 'Not authenticated' })
    return
  }
  next()
}
