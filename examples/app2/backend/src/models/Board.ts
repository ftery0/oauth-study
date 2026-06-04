import { Schema, model, type InferSchemaType } from 'mongoose'

const boardSchema = new Schema(
  {
    ownerSub: { type: String, required: true, index: true },
    title:    { type: String, required: true },
  },
  { timestamps: true },
)

export type Board = InferSchemaType<typeof boardSchema>
export const BoardModel = model('Board', boardSchema)
