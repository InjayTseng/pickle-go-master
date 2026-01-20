# Pickle Go 部署文件

## 概述

本文件說明如何將 Pickle Go 部署到 Google Cloud Platform (GCP)。應用程式使用以下 GCP 服務：

- **Cloud Run**: 運行後端 API 和前端應用程式
- **Cloud SQL (PostgreSQL)**: 資料庫服務
- **Secret Manager**: 管理敏感資訊
- **Cloud Build**: CI/CD 自動化部署
- **Container Registry / Artifact Registry**: 儲存 Docker 映像檔

---

## 目錄

1. [前置準備](#1-前置準備)
2. [資料庫設定](#2-資料庫設定)
3. [Secret Manager 設定](#3-secret-manager-設定)
4. [後端 API 部署](#4-後端-api-部署)
5. [前端部署](#5-前端部署)
6. [CI/CD 設定](#6-cicd-設定)
7. [環境變數配置](#7-環境變數配置)
8. [網域設定](#8-網域設定)
9. [監控與告警](#9-監控與告警)
10. [疑難排解](#10-疑難排解)

---

## 1. 前置準備

### 1.1 安裝必要工具

```bash
# Google Cloud SDK
curl https://sdk.cloud.google.com | bash
exec -l $SHELL

# Docker
# 參考: https://docs.docker.com/get-docker/

# golang-migrate (資料庫遷移工具)
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### 1.2 設定 GCP 專案

```bash
# 登入 GCP
gcloud auth login

# 建立新專案（或使用現有專案）
export PROJECT_ID=pickle-go-prod
gcloud projects create $PROJECT_ID --name="Pickle Go Production"

# 設定預設專案
gcloud config set project $PROJECT_ID

# 啟用必要的 API
gcloud services enable \
  run.googleapis.com \
  sql-component.googleapis.com \
  sqladmin.googleapis.com \
  secretmanager.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com
```

### 1.3 設定環境變數

```bash
export PROJECT_ID=pickle-go-prod
export REGION=asia-east1          # 台灣區域
export INSTANCE_NAME=pickle-go-db
export SERVICE_NAME_API=pickle-go-api
export SERVICE_NAME_WEB=pickle-go-web
```

---

## 2. 資料庫設定

### 2.1 建立 Cloud SQL 實例

```bash
# 建立 PostgreSQL 實例
gcloud sql instances create $INSTANCE_NAME \
  --database-version=POSTGRES_15 \
  --tier=db-custom-2-4096 \
  --region=$REGION \
  --storage-type=SSD \
  --storage-size=10GB \
  --storage-auto-increase \
  --storage-auto-increase-limit=100 \
  --backup-start-time=04:00 \
  --maintenance-window-day=SUN \
  --maintenance-window-hour=05 \
  --database-flags=max_connections=100,shared_buffers=256MB \
  --root-password=$(openssl rand -base64 32)

# 記錄 root 密碼
echo "Root password: [顯示的密碼]" > .secrets/db-root-password.txt
```

### 2.2 建立資料庫和使用者

```bash
# 建立應用程式資料庫
gcloud sql databases create picklego --instance=$INSTANCE_NAME

# 建立應用程式使用者
export DB_PASSWORD=$(openssl rand -base64 32)
gcloud sql users create picklego-user \
  --instance=$INSTANCE_NAME \
  --password=$DB_PASSWORD

# 儲存密碼
echo "DB User: picklego-user" > .secrets/db-credentials.txt
echo "DB Password: $DB_PASSWORD" >> .secrets/db-credentials.txt
```

### 2.3 啟用 PostGIS 擴展

使用 Cloud SQL Proxy 連接到資料庫：

```bash
# 啟動 Cloud SQL Proxy
cloud_sql_proxy -instances=$PROJECT_ID:$REGION:$INSTANCE_NAME=tcp:5432 &

# 連接到資料庫
psql "host=127.0.0.1 port=5432 user=postgres dbname=picklego"
```

在 psql 中執行：

```sql
-- 啟用擴展
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- 驗證擴展
SELECT name, default_version, installed_version
FROM pg_available_extensions
WHERE name IN ('postgis', 'uuid-ossp', 'pg_stat_statements');

-- 授權給應用程式使用者
GRANT ALL PRIVILEGES ON DATABASE picklego TO "picklego-user";
\c picklego
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO "picklego-user";
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO "picklego-user";
```

### 2.4 執行資料庫遷移

```bash
# 設定連線字串
export DATABASE_URL="postgres://picklego-user:$DB_PASSWORD@localhost:5432/picklego?sslmode=disable"

# 執行 migration
cd apps/api
migrate -path migrations -database "$DATABASE_URL" up

# 驗證 migration 狀態
migrate -path migrations -database "$DATABASE_URL" version
```

---

## 3. Secret Manager 設定

### 3.1 建立 Secrets

```bash
# 資料庫連線字串
export DB_CONNECTION_NAME="$PROJECT_ID:$REGION:$INSTANCE_NAME"
echo -n "postgres://picklego-user:$DB_PASSWORD@/picklego?host=/cloudsql/$DB_CONNECTION_NAME" | \
  gcloud secrets create pickle-go-database-url --data-file=-

# JWT Secret (至少 32 字元)
echo -n $(openssl rand -base64 48) | \
  gcloud secrets create pickle-go-jwt-secret --data-file=-

# Line Login 設定
echo -n "YOUR_LINE_CHANNEL_ID" | \
  gcloud secrets create pickle-go-line-channel-id --data-file=-

echo -n "YOUR_LINE_CHANNEL_SECRET" | \
  gcloud secrets create pickle-go-line-channel-secret --data-file=-

# Line Redirect URI
echo -n "https://picklego.tw/auth/callback" | \
  gcloud secrets create pickle-go-line-redirect-uri --data-file=-

# CORS Allowed Origins
echo -n "https://picklego.tw,https://www.picklego.tw" | \
  gcloud secrets create pickle-go-cors-origins --data-file=-

# Sentry DSN (API)
echo -n "YOUR_SENTRY_DSN" | \
  gcloud secrets create pickle-go-sentry-dsn-api --data-file=-

# Sentry DSN (Web)
echo -n "YOUR_SENTRY_DSN" | \
  gcloud secrets create pickle-go-sentry-dsn-web --data-file=-

# Google Maps API Key
echo -n "YOUR_GOOGLE_MAPS_API_KEY" | \
  gcloud secrets create pickle-go-google-maps-key --data-file=-

# Google Analytics Measurement ID
echo -n "G-XXXXXXXXXX" | \
  gcloud secrets create pickle-go-ga-measurement-id --data-file=-
```

### 3.2 授權 Cloud Run 存取 Secrets

```bash
# 取得 Cloud Run 預設服務帳戶
export SERVICE_ACCOUNT=$(gcloud iam service-accounts list \
  --filter="displayName:Compute Engine default service account" \
  --format="value(email)")

# 授權存取所有 secrets
for SECRET in $(gcloud secrets list --format="value(name)" --filter="name~pickle-go"); do
  gcloud secrets add-iam-policy-binding $SECRET \
    --member="serviceAccount:$SERVICE_ACCOUNT" \
    --role="roles/secretmanager.secretAccessor"
done
```

---

## 4. 後端 API 部署

### 4.1 建立 Dockerfile

確認 `apps/api/Dockerfile` 內容：

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app

# 安裝必要工具
RUN apk add --no-cache git

# 複製 go mod files
COPY go.mod go.sum ./
RUN go mod download

# 複製原始碼
COPY . .

# 編譯
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Run stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/main .

# 設定 port
ENV PORT=8080
EXPOSE 8080

CMD ["./main"]
```

### 4.2 部署到 Cloud Run

```bash
cd apps/api

# 設定 Cloud Build
gcloud builds submit --tag gcr.io/$PROJECT_ID/$SERVICE_NAME_API

# 部署到 Cloud Run
gcloud run deploy $SERVICE_NAME_API \
  --image gcr.io/$PROJECT_ID/$SERVICE_NAME_API \
  --platform managed \
  --region $REGION \
  --allow-unauthenticated \
  --memory 512Mi \
  --cpu 1 \
  --min-instances 0 \
  --max-instances 10 \
  --timeout 60 \
  --concurrency 80 \
  --add-cloudsql-instances $DB_CONNECTION_NAME \
  --set-env-vars="PORT=8080,ENVIRONMENT=production" \
  --set-secrets="DATABASE_URL=pickle-go-database-url:latest,\
JWT_SECRET=pickle-go-jwt-secret:latest,\
LINE_CHANNEL_ID=pickle-go-line-channel-id:latest,\
LINE_CHANNEL_SECRET=pickle-go-line-channel-secret:latest,\
LINE_REDIRECT_URI=pickle-go-line-redirect-uri:latest,\
CORS_ALLOWED_ORIGINS=pickle-go-cors-origins:latest,\
SENTRY_DSN=pickle-go-sentry-dsn-api:latest"

# 取得服務 URL
export API_URL=$(gcloud run services describe $SERVICE_NAME_API \
  --platform managed \
  --region $REGION \
  --format 'value(status.url)')

echo "API URL: $API_URL"
```

### 4.3 驗證部署

```bash
# 健康檢查
curl $API_URL/health

# 預期回應:
# {"status":"ok","service":"pickle-go-api","version":"0.1.0"}

# 測試 API
curl $API_URL/api/v1/events | jq
```

---

## 5. 前端部署

### 5.1 建立 Dockerfile

確認 `apps/web/Dockerfile` 內容：

```dockerfile
# Dependencies stage
FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

# Builder stage
FROM node:20-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

# Build 參數
ARG NEXT_PUBLIC_API_URL
ARG NEXT_PUBLIC_BASE_URL
ARG NEXT_PUBLIC_GOOGLE_MAPS_API_KEY
ARG NEXT_PUBLIC_LINE_CHANNEL_ID
ARG NEXT_PUBLIC_LINE_REDIRECT_URI
ARG NEXT_PUBLIC_SENTRY_DSN
ARG NEXT_PUBLIC_GA_MEASUREMENT_ID

ENV NEXT_TELEMETRY_DISABLED 1

RUN npm run build

# Runner stage
FROM node:20-alpine AS runner
WORKDIR /app

ENV NODE_ENV production
ENV NEXT_TELEMETRY_DISABLED 1

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000
ENV PORT 3000

CMD ["node", "server.js"]
```

### 5.2 更新 next.config.js

```javascript
/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  images: {
    domains: ['profile.line-scdn.net'],
  },
};

module.exports = nextConfig;
```

### 5.3 部署到 Cloud Run

```bash
cd apps/web

# 從 Secret Manager 取得環境變數
export GOOGLE_MAPS_KEY=$(gcloud secrets versions access latest --secret=pickle-go-google-maps-key)
export LINE_CHANNEL_ID=$(gcloud secrets versions access latest --secret=pickle-go-line-channel-id)
export SENTRY_DSN_WEB=$(gcloud secrets versions access latest --secret=pickle-go-sentry-dsn-web)
export GA_ID=$(gcloud secrets versions access latest --secret=pickle-go-ga-measurement-id)

# Build with Cloud Build
gcloud builds submit --tag gcr.io/$PROJECT_ID/$SERVICE_NAME_WEB \
  --build-arg NEXT_PUBLIC_API_URL=https://api.picklego.tw/api/v1 \
  --build-arg NEXT_PUBLIC_BASE_URL=https://picklego.tw \
  --build-arg NEXT_PUBLIC_GOOGLE_MAPS_API_KEY=$GOOGLE_MAPS_KEY \
  --build-arg NEXT_PUBLIC_LINE_CHANNEL_ID=$LINE_CHANNEL_ID \
  --build-arg NEXT_PUBLIC_LINE_REDIRECT_URI=https://picklego.tw/auth/callback \
  --build-arg NEXT_PUBLIC_SENTRY_DSN=$SENTRY_DSN_WEB \
  --build-arg NEXT_PUBLIC_GA_MEASUREMENT_ID=$GA_ID

# 部署到 Cloud Run
gcloud run deploy $SERVICE_NAME_WEB \
  --image gcr.io/$PROJECT_ID/$SERVICE_NAME_WEB \
  --platform managed \
  --region $REGION \
  --allow-unauthenticated \
  --memory 512Mi \
  --cpu 1 \
  --min-instances 0 \
  --max-instances 10 \
  --timeout 60 \
  --concurrency 80 \
  --port 3000

# 取得服務 URL
export WEB_URL=$(gcloud run services describe $SERVICE_NAME_WEB \
  --platform managed \
  --region $REGION \
  --format 'value(status.url)')

echo "Web URL: $WEB_URL"
```

---

## 6. CI/CD 設定

### 6.1 建立 Cloud Build 配置

建立 `cloudbuild.yaml` 於專案根目錄：

```yaml
steps:
  # Build API
  - name: 'gcr.io/cloud-builders/docker'
    id: 'build-api'
    args:
      - 'build'
      - '-t'
      - 'gcr.io/$PROJECT_ID/pickle-go-api'
      - './apps/api'

  # Build Web
  - name: 'gcr.io/cloud-builders/docker'
    id: 'build-web'
    args:
      - 'build'
      - '-t'
      - 'gcr.io/$PROJECT_ID/pickle-go-web'
      - '--build-arg'
      - 'NEXT_PUBLIC_API_URL=https://api.picklego.tw/api/v1'
      - '--build-arg'
      - 'NEXT_PUBLIC_BASE_URL=https://picklego.tw'
      - './apps/web'

  # Push images
  - name: 'gcr.io/cloud-builders/docker'
    id: 'push-api'
    args: ['push', 'gcr.io/$PROJECT_ID/pickle-go-api']
    waitFor: ['build-api']

  - name: 'gcr.io/cloud-builders/docker'
    id: 'push-web'
    args: ['push', 'gcr.io/$PROJECT_ID/pickle-go-web']
    waitFor: ['build-web']

  # Deploy API
  - name: 'gcr.io/google.com/cloudsdktool/cloud-sdk'
    id: 'deploy-api'
    entrypoint: gcloud
    args:
      - 'run'
      - 'deploy'
      - 'pickle-go-api'
      - '--image'
      - 'gcr.io/$PROJECT_ID/pickle-go-api'
      - '--region'
      - 'asia-east1'
      - '--platform'
      - 'managed'
    waitFor: ['push-api']

  # Deploy Web
  - name: 'gcr.io/google.com/cloudsdktool/cloud-sdk'
    id: 'deploy-web'
    entrypoint: gcloud
    args:
      - 'run'
      - 'deploy'
      - 'pickle-go-web'
      - '--image'
      - 'gcr.io/$PROJECT_ID/pickle-go-web'
      - '--region'
      - 'asia-east1'
      - '--platform'
      - 'managed'
    waitFor: ['push-web']

images:
  - 'gcr.io/$PROJECT_ID/pickle-go-api'
  - 'gcr.io/$PROJECT_ID/pickle-go-web'

timeout: 1200s
```

### 6.2 設定 GitHub 連結

```bash
# 連結 GitHub Repository
gcloud builds triggers create github \
  --repo-name=pickle-go \
  --repo-owner=YOUR_GITHUB_USERNAME \
  --branch-pattern="^main$" \
  --build-config=cloudbuild.yaml
```

### 6.3 手動觸發部署

```bash
gcloud builds submit --config=cloudbuild.yaml
```

---

## 7. 環境變數配置

### 7.1 後端環境變數

| 變數名稱 | 來源 | 說明 |
|---------|------|------|
| `PORT` | 環境變數 | 服務埠號（預設 8080） |
| `ENVIRONMENT` | 環境變數 | 環境名稱（production） |
| `DATABASE_URL` | Secret Manager | 資料庫連線字串 |
| `JWT_SECRET` | Secret Manager | JWT 簽章密鑰 |
| `JWT_EXPIRY` | 環境變數 | Token 有效期（預設 168h） |
| `LINE_CHANNEL_ID` | Secret Manager | Line Channel ID |
| `LINE_CHANNEL_SECRET` | Secret Manager | Line Channel Secret |
| `LINE_REDIRECT_URI` | Secret Manager | Line OAuth 回呼 URI |
| `CORS_ALLOWED_ORIGINS` | Secret Manager | CORS 允許的來源 |
| `SENTRY_DSN` | Secret Manager | Sentry 錯誤追蹤 DSN |
| `SENTRY_ENVIRONMENT` | 環境變數 | Sentry 環境名稱 |
| `SENTRY_RELEASE` | 環境變數 | Sentry Release 版本 |

### 7.2 前端環境變數

| 變數名稱 | 來源 | 說明 |
|---------|------|------|
| `NEXT_PUBLIC_API_URL` | Build Args | API 端點 URL |
| `NEXT_PUBLIC_BASE_URL` | Build Args | 網站基礎 URL |
| `NEXT_PUBLIC_GOOGLE_MAPS_API_KEY` | Build Args | Google Maps API Key |
| `NEXT_PUBLIC_LINE_CHANNEL_ID` | Build Args | Line Channel ID |
| `NEXT_PUBLIC_LINE_REDIRECT_URI` | Build Args | Line OAuth 回呼 URI |
| `NEXT_PUBLIC_SENTRY_DSN` | Build Args | Sentry DSN |
| `NEXT_PUBLIC_GA_MEASUREMENT_ID` | Build Args | Google Analytics ID |

### 7.3 更新環境變數

```bash
# 更新 Cloud Run 環境變數
gcloud run services update $SERVICE_NAME_API \
  --region $REGION \
  --set-env-vars="NEW_VAR=value"

# 更新 Secret
echo -n "new-secret-value" | gcloud secrets versions add SECRET_NAME --data-file=-
```

---

## 8. 網域設定

### 8.1 設定自訂網域

```bash
# API 網域
gcloud run domain-mappings create \
  --service $SERVICE_NAME_API \
  --domain api.picklego.tw \
  --region $REGION

# Web 網域
gcloud run domain-mappings create \
  --service $SERVICE_NAME_WEB \
  --domain picklego.tw \
  --region $REGION

gcloud run domain-mappings create \
  --service $SERVICE_NAME_WEB \
  --domain www.picklego.tw \
  --region $REGION
```

### 8.2 設定 DNS 記錄

在你的 DNS 提供商（如 Cloudflare、GoDaddy）新增以下記錄：

```
# API
Type: CNAME
Name: api
Value: ghs.googlehosted.com

# Web
Type: A
Name: @
Value: [Cloud Run 提供的 IP]

Type: A
Name: www
Value: [Cloud Run 提供的 IP]
```

### 8.3 SSL 憑證

Cloud Run 會自動配置 SSL 憑證（Let's Encrypt），通常需要 15-20 分鐘生效。

驗證 SSL：

```bash
# 檢查憑證狀態
gcloud run domain-mappings describe \
  --domain api.picklego.tw \
  --region $REGION
```

---

## 9. 監控與告警

### 9.1 設定 Cloud Monitoring

```bash
# 建立通知頻道（Email）
gcloud alpha monitoring channels create \
  --display-name="Team Email" \
  --type=email \
  --channel-labels=email_address=team@picklego.tw
```

### 9.2 建立告警政策

```bash
# API 錯誤率告警
gcloud alpha monitoring policies create \
  --notification-channels=CHANNEL_ID \
  --display-name="API Error Rate High" \
  --condition-display-name="Error rate > 5%" \
  --condition-threshold-value=0.05 \
  --condition-threshold-duration=300s
```

### 9.3 重要監控指標

在 GCP Console 的 Cloud Monitoring 中追蹤：

#### Cloud Run 指標
- Request Count (請求數)
- Request Latency (延遲時間)
- Error Rate (錯誤率)
- CPU Utilization (CPU 使用率)
- Memory Utilization (記憶體使用率)
- Instance Count (實例數量)

#### Cloud SQL 指標
- CPU Utilization (CPU 使用率)
- Memory Utilization (記憶體使用率)
- Connections (連線數)
- Disk Usage (磁碟使用率)
- Query Latency (查詢延遲)

### 9.4 日誌查詢

```bash
# 查看 API 日誌
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=$SERVICE_NAME_API" \
  --limit 50 \
  --format json

# 查看錯誤日誌
gcloud logging read "resource.type=cloud_run_revision AND severity>=ERROR" \
  --limit 50
```

---

## 10. 疑難排解

### 10.1 服務無法啟動

```bash
# 查看詳細日誌
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=$SERVICE_NAME_API" \
  --limit 100 \
  --format="table(timestamp,textPayload)"

# 檢查服務狀態
gcloud run services describe $SERVICE_NAME_API \
  --region $REGION \
  --format="value(status.conditions)"
```

### 10.2 資料庫連線失敗

```bash
# 檢查 Cloud SQL 實例狀態
gcloud sql instances describe $INSTANCE_NAME

# 測試連線
cloud_sql_proxy -instances=$PROJECT_ID:$REGION:$INSTANCE_NAME=tcp:5432 &
psql "host=127.0.0.1 port=5432 user=picklego-user dbname=picklego"

# 檢查連線字串格式
# 正確格式: postgres://USER:PASS@/DB?host=/cloudsql/PROJECT:REGION:INSTANCE
```

### 10.3 環境變數問題

```bash
# 列出所有環境變數
gcloud run services describe $SERVICE_NAME_API \
  --region $REGION \
  --format="value(spec.template.spec.containers[0].env)"

# 列出所有 secrets
gcloud run services describe $SERVICE_NAME_API \
  --region $REGION \
  --format="value(spec.template.spec.containers[0].env)" | grep secret
```

### 10.4 記憶體不足

```bash
# 增加記憶體配置
gcloud run services update $SERVICE_NAME_API \
  --region $REGION \
  --memory 1Gi

# 查看記憶體使用情況
gcloud monitoring time-series list \
  --filter='metric.type="run.googleapis.com/container/memory/utilizations"' \
  --format=json
```

### 10.5 冷啟動時間過長

```bash
# 設定最小實例數（避免冷啟動）
gcloud run services update $SERVICE_NAME_API \
  --region $REGION \
  --min-instances 1

# 注意：最小實例會產生持續費用
```

---

## 11. 成本優化

### 11.1 Cloud Run 成本優化

```bash
# 調整資源配置
gcloud run services update $SERVICE_NAME_API \
  --region $REGION \
  --memory 256Mi \
  --cpu 1 \
  --max-instances 5

# 設定請求超時時間
gcloud run services update $SERVICE_NAME_API \
  --region $REGION \
  --timeout 30
```

### 11.2 Cloud SQL 成本優化

```bash
# 開發環境：使用較小機型
gcloud sql instances patch $INSTANCE_NAME \
  --tier db-f1-micro

# 生產環境：啟用自動備份但限制保留期
gcloud sql instances patch $INSTANCE_NAME \
  --backup-start-time 04:00 \
  --retained-backups-count 7

# 設定維護視窗（避免高峰時段）
gcloud sql instances patch $INSTANCE_NAME \
  --maintenance-window-day=SUN \
  --maintenance-window-hour=5
```

### 11.3 預估成本

| 服務 | 配置 | 月費用（USD） |
|-----|------|-------------|
| Cloud Run API | 512Mi, 1 CPU, 100K req/month | ~$5 |
| Cloud Run Web | 512Mi, 1 CPU, 500K req/month | ~$20 |
| Cloud SQL | db-custom-2-4096, 10GB | ~$90 |
| Cloud Storage | 10GB | ~$0.20 |
| Secret Manager | 10 secrets | ~$0.60 |
| **總計** | | **~$116/月** |

> 注意：實際費用取決於使用量。請參考 [GCP 定價計算器](https://cloud.google.com/products/calculator)。

---

## 12. 備份與災難恢復

### 12.1 資料庫備份

```bash
# 立即建立備份
gcloud sql backups create \
  --instance $INSTANCE_NAME \
  --description "Manual backup $(date +%Y%m%d)"

# 列出所有備份
gcloud sql backups list --instance $INSTANCE_NAME

# 從備份還原
gcloud sql backups restore BACKUP_ID \
  --backup-instance $INSTANCE_NAME \
  --backup-id BACKUP_ID
```

### 12.2 容器映像備份

```bash
# 列出所有映像
gcloud container images list

# 複製映像到 Artifact Registry（建議）
gcloud artifacts repositories create pickle-go-repo \
  --repository-format=docker \
  --location=$REGION

# 標記並推送
docker tag gcr.io/$PROJECT_ID/pickle-go-api \
  $REGION-docker.pkg.dev/$PROJECT_ID/pickle-go-repo/pickle-go-api:latest
docker push $REGION-docker.pkg.dev/$PROJECT_ID/pickle-go-repo/pickle-go-api:latest
```

### 12.3 災難恢復計劃

1. **RTO (Recovery Time Objective)**: 30 分鐘
2. **RPO (Recovery Point Objective)**: 1 小時（自動備份頻率）

**恢復步驟**：

```bash
# 1. 從最新備份還原資料庫
gcloud sql backups restore LATEST_BACKUP_ID \
  --backup-instance $INSTANCE_NAME

# 2. 從已知良好的映像重新部署
gcloud run deploy $SERVICE_NAME_API \
  --image gcr.io/$PROJECT_ID/pickle-go-api:LAST_KNOWN_GOOD_TAG \
  --region $REGION

# 3. 驗證服務
curl https://api.picklego.tw/health

# 4. 監控錯誤日誌
gcloud logging tail "resource.type=cloud_run_revision" --format=json
```

---

## 13. 安全性檢查清單

- [ ] 所有敏感資訊都儲存在 Secret Manager
- [ ] 資料庫使用私有 IP（僅 Cloud Run 可存取）
- [ ] 啟用 Cloud SQL 自動備份
- [ ] 設定 Cloud Armor（如需防護 DDoS）
- [ ] 啟用 VPC Service Controls（企業級安全）
- [ ] 設定 IAM 角色和權限（最小權限原則）
- [ ] 啟用審計日誌
- [ ] 定期更新依賴套件
- [ ] 啟用 HTTPS（Cloud Run 預設）
- [ ] 設定 CORS 白名單

---

## 14. 效能優化

### 14.1 CDN 設定（Cloud CDN）

```bash
# 建立 Load Balancer（需要靜態 IP）
gcloud compute addresses create pickle-go-ip \
  --global

# 建立後端服務
gcloud compute backend-services create pickle-go-backend \
  --global \
  --enable-cdn

# 啟用 CDN 快取
gcloud compute backend-services update pickle-go-backend \
  --global \
  --cache-mode=CACHE_ALL_STATIC
```

### 14.2 快取策略

在應用程式層實作：

```go
// API Response 快取 Header
c.Header("Cache-Control", "public, max-age=300") // 5 分鐘

// 靜態資源快取
c.Header("Cache-Control", "public, max-age=31536000, immutable") // 1 年
```

### 14.3 資料庫連線池

```go
// apps/api/internal/database/db.go
dbConfig := database.DefaultConfig(cfg.DatabaseURL)
dbConfig.MaxOpenConns = 25
dbConfig.MaxIdleConns = 5
dbConfig.ConnMaxLifetime = 5 * time.Minute
```

---

## 15. 版本管理與回滾

### 15.1 版本標記

```bash
# 建立版本標記
export VERSION=v1.0.0
docker tag gcr.io/$PROJECT_ID/pickle-go-api gcr.io/$PROJECT_ID/pickle-go-api:$VERSION
docker push gcr.io/$PROJECT_ID/pickle-go-api:$VERSION

# 使用特定版本部署
gcloud run deploy $SERVICE_NAME_API \
  --image gcr.io/$PROJECT_ID/pickle-go-api:$VERSION \
  --region $REGION
```

### 15.2 快速回滾

```bash
# 列出歷史修訂版本
gcloud run revisions list \
  --service $SERVICE_NAME_API \
  --region $REGION

# 切換到特定修訂版本
gcloud run services update-traffic $SERVICE_NAME_API \
  --region $REGION \
  --to-revisions REVISION_NAME=100
```

### 15.3 金絲雀部署

```bash
# 新版本部署但不導流
gcloud run deploy $SERVICE_NAME_API \
  --image gcr.io/$PROJECT_ID/pickle-go-api:v2.0.0 \
  --region $REGION \
  --no-traffic

# 導入 10% 流量測試
gcloud run services update-traffic $SERVICE_NAME_API \
  --region $REGION \
  --to-revisions LATEST=10,REVISION_V1=90

# 確認無誤後全部切換
gcloud run services update-traffic $SERVICE_NAME_API \
  --region $REGION \
  --to-latest
```

---

## 16. 資源連結

- [Cloud Run 文件](https://cloud.google.com/run/docs)
- [Cloud SQL 文件](https://cloud.google.com/sql/docs)
- [Secret Manager 文件](https://cloud.google.com/secret-manager/docs)
- [Cloud Build 文件](https://cloud.google.com/build/docs)
- [GCP 最佳實踐](https://cloud.google.com/architecture/best-practices)

---

## 17. 支援與聯絡

如有部署相關問題：

- **技術文件**: 本資料夾內其他文件
- **GitHub Issues**: https://github.com/anthropics/pickle-go/issues
- **Email**: devops@picklego.tw

---

**版本**: 1.0.0
**最後更新**: 2026-01-21
**適用環境**: GCP Cloud Run + Cloud SQL
