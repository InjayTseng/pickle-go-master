# Cloud SQL (PostgreSQL) 生產環境設定指南

本文件說明如何在 GCP 上設定 Cloud SQL PostgreSQL 實例並與 Cloud Run 連接。

## 1. 建立 Cloud SQL 實例

```bash
# 設定專案和區域
export PROJECT_ID=your-project-id
export REGION=asia-east1
export INSTANCE_NAME=pickle-go-db

# 建立 Cloud SQL 實例 (PostgreSQL 15)
gcloud sql instances create $INSTANCE_NAME \
  --database-version=POSTGRES_15 \
  --tier=db-f1-micro \
  --region=$REGION \
  --storage-type=SSD \
  --storage-size=10GB \
  --storage-auto-increase \
  --backup-start-time=04:00 \
  --maintenance-window-day=SUN \
  --maintenance-window-hour=05 \
  --database-flags=max_connections=100 \
  --root-password=your-secure-root-password
```

## 2. 建立資料庫和使用者

```bash
# 建立資料庫
gcloud sql databases create picklego --instance=$INSTANCE_NAME

# 建立應用程式使用者
gcloud sql users create picklego-user \
  --instance=$INSTANCE_NAME \
  --password=your-secure-password
```

## 3. 啟用必要的擴展 (PostGIS)

連接到資料庫並執行:

```sql
-- 啟用 PostGIS 擴展
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

## 4. 執行資料庫遷移

```bash
# 使用 Cloud SQL Proxy 本地連接
cloud_sql_proxy -instances=$PROJECT_ID:$REGION:$INSTANCE_NAME=tcp:5432

# 在另一個終端執行遷移
DATABASE_URL="postgres://picklego-user:your-password@localhost:5432/picklego?sslmode=disable" \
  migrate -path apps/api/migrations -database $DATABASE_URL up
```

## 5. 設定 Secret Manager

將敏感資訊存入 Secret Manager:

```bash
# 資料庫連線字串
echo -n "postgres://picklego-user:your-password@/picklego?host=/cloudsql/$PROJECT_ID:$REGION:$INSTANCE_NAME" | \
  gcloud secrets create pickle-go-database-url --data-file=-

# JWT Secret
echo -n "your-strong-jwt-secret-key-at-least-32-chars" | \
  gcloud secrets create pickle-go-jwt-secret --data-file=-

# Line Channel ID
echo -n "your-line-channel-id" | \
  gcloud secrets create pickle-go-line-channel-id --data-file=-

# Line Channel Secret
echo -n "your-line-channel-secret" | \
  gcloud secrets create pickle-go-line-channel-secret --data-file=-

# Sentry DSN (API)
echo -n "https://xxx@sentry.io/xxx" | \
  gcloud secrets create pickle-go-sentry-dsn-api --data-file=-
```

## 6. 授權 Cloud Run 存取 Secrets

```bash
# 取得 Cloud Run 服務帳戶
export SERVICE_ACCOUNT=$(gcloud iam service-accounts list \
  --filter="displayName:Compute Engine default service account" \
  --format="value(email)")

# 授權存取 secrets
for SECRET in pickle-go-database-url pickle-go-jwt-secret pickle-go-line-channel-id pickle-go-line-channel-secret pickle-go-sentry-dsn-api; do
  gcloud secrets add-iam-policy-binding $SECRET \
    --member="serviceAccount:$SERVICE_ACCOUNT" \
    --role="roles/secretmanager.secretAccessor"
done
```

## 7. 連線字串格式

### Cloud Run 到 Cloud SQL (使用 Unix Socket)
```
postgres://USER:PASSWORD@/DATABASE?host=/cloudsql/PROJECT_ID:REGION:INSTANCE_NAME
```

### 本地開發 (使用 Cloud SQL Proxy)
```
postgres://USER:PASSWORD@localhost:5432/DATABASE?sslmode=disable
```

## 8. 備份與還原

### 手動備份
```bash
gcloud sql backups create --instance=$INSTANCE_NAME
```

### 還原備份
```bash
gcloud sql backups restore BACKUP_ID --restore-instance=$INSTANCE_NAME
```

## 9. 監控

建議在 GCP Console 設定以下警報:
- CPU 使用率 > 80%
- 記憶體使用率 > 80%
- 連線數 > 80
- 磁碟使用率 > 80%

## 10. 成本優化

開發/測試環境建議:
- 使用 `db-f1-micro` 或 `db-g1-small` 機型
- 設定自動停止閒置實例
- 考慮使用預留實例折扣

生產環境建議:
- 最少使用 `db-custom-2-4096` (2 vCPU, 4GB RAM)
- 啟用高可用性 (HA)
- 啟用自動備份
