import mongoose from 'mongoose'
import { env } from './env'

export async function connectMongo(): Promise<void> {
  mongoose.set('strictQuery', true)
  await mongoose.connect(env.MONGO_URL)
  console.log(`[mongo] connected → ${env.MONGO_URL.replace(/\/\/.*@/, '//***@')}`)
}
