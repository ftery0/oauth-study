import { Schema, model, type InferSchemaType } from 'mongoose'

const userSchema = new Schema(
  {
    sub:         { type: String, required: true, unique: true, index: true },
    displayName: { type: String, required: true },
    email:       { type: String, default: null },
  },
  { timestamps: { createdAt: 'createdAt', updatedAt: false } },
)

export type User = InferSchemaType<typeof userSchema>
export const UserModel = model('User', userSchema)
