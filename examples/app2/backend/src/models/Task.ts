import { Schema, model, type InferSchemaType, Types } from 'mongoose'

export const TASK_COLUMNS = ['todo', 'doing', 'done'] as const
export type TaskColumn = (typeof TASK_COLUMNS)[number]

const taskSchema = new Schema(
  {
    boardId:  { type: Schema.Types.ObjectId, ref: 'Board', required: true, index: true },
    ownerSub: { type: String, required: true, index: true },
    title:    { type: String, required: true },
    column:   { type: String, enum: TASK_COLUMNS, default: 'todo', required: true },
    position: { type: Number, default: 0, required: true },
    done:     { type: Boolean, default: false, required: true },
  },
  { timestamps: true },
)

export type Task = InferSchemaType<typeof taskSchema>
export const TaskModel = model('Task', taskSchema)
export { Types as MongoTypes }
