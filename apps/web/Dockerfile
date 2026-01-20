# ===========================================
# Pickle Go Web - Dockerfile
# ===========================================
# 多階段建置 Next.js 應用程式，優化生產環境映像大小

# Stage 1: 安裝依賴套件
FROM node:20-alpine AS deps
RUN apk add --no-cache libc6-compat
WORKDIR /app

# 複製 package 檔案
COPY package.json pnpm-lock.yaml* ./

# 安裝 pnpm 並安裝依賴
RUN corepack enable pnpm && pnpm install --frozen-lockfile

# Stage 2: 建置應用程式
FROM node:20-alpine AS builder
WORKDIR /app

# 複製依賴
COPY --from=deps /app/node_modules ./node_modules
COPY . .

# 建置時環境變數 (透過 build args 傳入)
ARG NEXT_PUBLIC_API_URL
ARG NEXT_PUBLIC_BASE_URL
ARG NEXT_PUBLIC_LINE_CHANNEL_ID
ARG NEXT_PUBLIC_LINE_REDIRECT_URI
ARG NEXT_PUBLIC_GOOGLE_MAPS_API_KEY
ARG NEXT_PUBLIC_SENTRY_DSN
ARG NEXT_PUBLIC_GA_MEASUREMENT_ID

ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_BASE_URL=$NEXT_PUBLIC_BASE_URL
ENV NEXT_PUBLIC_LINE_CHANNEL_ID=$NEXT_PUBLIC_LINE_CHANNEL_ID
ENV NEXT_PUBLIC_LINE_REDIRECT_URI=$NEXT_PUBLIC_LINE_REDIRECT_URI
ENV NEXT_PUBLIC_GOOGLE_MAPS_API_KEY=$NEXT_PUBLIC_GOOGLE_MAPS_API_KEY
ENV NEXT_PUBLIC_SENTRY_DSN=$NEXT_PUBLIC_SENTRY_DSN
ENV NEXT_PUBLIC_GA_MEASUREMENT_ID=$NEXT_PUBLIC_GA_MEASUREMENT_ID

# 啟用 Next.js 輸出獨立模式
ENV NEXT_TELEMETRY_DISABLED=1

# 建置
RUN corepack enable pnpm && pnpm build

# Stage 3: 執行環境
FROM node:20-alpine AS runner
WORKDIR /app

# 設定生產環境
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

# 建立非 root 使用者
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

# 複製必要檔案
COPY --from=builder /app/public ./public

# 設定正確的權限給 prerender cache
RUN mkdir .next
RUN chown nextjs:nodejs .next

# 複製獨立建置輸出
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

# 切換到非 root 使用者
USER nextjs

# 暴露端口
EXPOSE 3000

# 設定端口環境變數
ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

# 執行應用程式
CMD ["node", "server.js"]
