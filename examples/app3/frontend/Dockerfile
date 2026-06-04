# app3 HelpDesk frontend — Next.js (standalone)
FROM node:20-alpine AS builder
WORKDIR /src
ARG BACKEND_URL=http://localhost:8013
ENV BACKEND_URL=$BACKEND_URL
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:20-alpine
WORKDIR /app
ENV NODE_ENV=production PORT=5183 HOSTNAME=0.0.0.0
COPY --from=builder /src/.next/standalone ./
COPY --from=builder /src/.next/static ./.next/static
EXPOSE 5183
CMD ["node","server.js"]
