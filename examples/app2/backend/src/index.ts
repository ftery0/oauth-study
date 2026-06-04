import express from 'express'
import session from 'express-session'
import { env } from './env'
import { connectMongo } from './db'
import { authRouter } from './routes/auth'
import { boardsRouter } from './routes/boards'
import { tasksRouter } from './routes/tasks'

async function main(): Promise<void> {
  await connectMongo()

  const app = express()
  app.use(express.json())
  app.use(express.urlencoded({ extended: true }))

  app.use(session({
    secret: env.SESSION_SECRET,
    resave: false,
    saveUninitialized: false,
    cookie: { httpOnly: true, maxAge: 24 * 60 * 60 * 1000 },
  }))

  app.use(authRouter)
  app.use('/api/boards', boardsRouter)
  app.use('/api/tasks', tasksRouter)

  app.use((err: unknown, _req: express.Request, res: express.Response, _next: express.NextFunction) => {
    console.error(err)
    res.status(500).json({ error: 'internal_server_error' })
  })

  app.listen(env.PORT, () => {
    console.log(`app2 backend  http://localhost:${env.PORT}`)
    console.log(`  client_id   ${env.CLIENT_ID}`)
    console.log(`  redirect    ${env.REDIRECT_URI}`)
    console.log(`  IdP         ${env.OAUTH_SERVER}`)
  })
}

main().catch(err => {
  console.error('startup failed:', err)
  process.exit(1)
})
