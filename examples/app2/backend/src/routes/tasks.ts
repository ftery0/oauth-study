import { Router } from 'express'
import { Types } from 'mongoose'
import { BoardModel } from '../models/Board'
import { TaskModel, TASK_COLUMNS, type TaskColumn } from '../models/Task'
import { requireUser } from '../middleware/auth'

export const tasksRouter = Router()
tasksRouter.use(requireUser)

function isColumn(v: unknown): v is TaskColumn {
  return typeof v === 'string' && (TASK_COLUMNS as readonly string[]).includes(v)
}

tasksRouter.get('/', async (req, res) => {
  const boardId = (req.query.boardId ?? '').toString()
  if (!Types.ObjectId.isValid(boardId)) return res.status(400).json({ error: 'invalid boardId' })

  const board = await BoardModel.findOne({ _id: boardId, ownerSub: req.session.userSub! }).lean()
  if (!board) return res.sendStatus(404)

  const items = await TaskModel.find({
    boardId: board._id,
    ownerSub: req.session.userSub!,
  })
    .sort({ column: 1, position: 1, createdAt: 1 })
    .lean()
  res.json(items)
})

tasksRouter.post('/', async (req, res) => {
  const boardId = (req.body?.boardId ?? '').toString()
  const title   = (req.body?.title ?? '').toString().trim()
  const column  = isColumn(req.body?.column) ? req.body.column : 'todo'

  if (!Types.ObjectId.isValid(boardId)) return res.status(400).json({ error: 'invalid boardId' })
  if (!title) return res.status(400).json({ error: 'title is required' })

  const board = await BoardModel.findOne({ _id: boardId, ownerSub: req.session.userSub! }).lean()
  if (!board) return res.sendStatus(404)

  const last = await TaskModel.findOne({ boardId: board._id, column })
    .sort({ position: -1 })
    .lean()
  const position = last ? last.position + 1 : 0

  const created = await TaskModel.create({
    boardId: board._id,
    ownerSub: req.session.userSub!,
    title,
    column,
    position,
    done: column === 'done',
  })
  res.status(201).json(created)
})

tasksRouter.patch('/:id', async (req, res) => {
  if (!Types.ObjectId.isValid(req.params.id)) return res.sendStatus(404)

  const update: Record<string, unknown> = {}
  if (typeof req.body?.title === 'string' && req.body.title.trim()) update.title = req.body.title.trim()
  if (isColumn(req.body?.column)) {
    update.column = req.body.column
    update.done = req.body.column === 'done'
  }
  if (typeof req.body?.position === 'number') update.position = req.body.position
  if (typeof req.body?.done === 'boolean') update.done = req.body.done

  if (Object.keys(update).length === 0) {
    return res.status(400).json({ error: 'no updatable fields' })
  }

  const updated = await TaskModel.findOneAndUpdate(
    { _id: req.params.id, ownerSub: req.session.userSub! },
    { $set: update },
    { new: true },
  ).lean()
  if (!updated) return res.sendStatus(404)
  res.json(updated)
})

tasksRouter.delete('/:id', async (req, res) => {
  if (!Types.ObjectId.isValid(req.params.id)) return res.sendStatus(404)
  const deleted = await TaskModel.findOneAndDelete({
    _id: req.params.id,
    ownerSub: req.session.userSub!,
  })
  if (!deleted) return res.sendStatus(404)
  res.sendStatus(204)
})
