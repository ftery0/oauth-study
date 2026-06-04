import { Router } from 'express'
import { Types } from 'mongoose'
import { BoardModel } from '../models/Board'
import { TaskModel } from '../models/Task'
import { requireUser } from '../middleware/auth'

export const boardsRouter = Router()
boardsRouter.use(requireUser)

boardsRouter.get('/', async (req, res) => {
  const items = await BoardModel.find({ ownerSub: req.session.userSub! })
    .sort({ createdAt: -1 })
    .lean()
  res.json(items)
})

boardsRouter.post('/', async (req, res) => {
  const title = (req.body?.title ?? '').toString().trim()
  if (!title) return res.status(400).json({ error: 'title is required' })

  const created = await BoardModel.create({ ownerSub: req.session.userSub!, title })
  res.status(201).json(created)
})

boardsRouter.patch('/:id', async (req, res) => {
  const title = (req.body?.title ?? '').toString().trim()
  if (!title) return res.status(400).json({ error: 'title is required' })

  if (!Types.ObjectId.isValid(req.params.id)) return res.sendStatus(404)

  const updated = await BoardModel.findOneAndUpdate(
    { _id: req.params.id, ownerSub: req.session.userSub! },
    { $set: { title } },
    { new: true },
  ).lean()
  if (!updated) return res.sendStatus(404)
  res.json(updated)
})

boardsRouter.delete('/:id', async (req, res) => {
  if (!Types.ObjectId.isValid(req.params.id)) return res.sendStatus(404)
  const deleted = await BoardModel.findOneAndDelete({
    _id: req.params.id,
    ownerSub: req.session.userSub!,
  })
  if (!deleted) return res.sendStatus(404)
  await TaskModel.deleteMany({ boardId: deleted._id, ownerSub: req.session.userSub! })
  res.sendStatus(204)
})
