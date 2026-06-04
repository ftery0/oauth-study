# app2 TaskBoard backend — Node + Express + Mongoose
FROM node:20-alpine AS builder
WORKDIR /src
COPY package*.json tsconfig.json ./
RUN npm ci
COPY src ./src
RUN npm run build

FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --omit=dev
COPY --from=builder /src/dist ./dist
EXPOSE 8012
ENV PORT=8012 NODE_ENV=production
CMD ["node","dist/index.js"]
