# 🚀 Hướng dẫn Deploy Backend

## 📋 Chuẩn bị trước khi deploy

### 1. ✅ Database Supabase
- Database URL: `postgresql://postgres:Duong19082k3@db.ofikztulwfiufxjymrsz.supabase.co:5432/postgres`
- Đã cập nhật trong `.env`

### 2. ✅ Environment Variables
Các biến cần thiết trong `.env`:
```
DB_HOST=db.ofikztulwfiufxjymrsz.supabase.co
DB_USER=postgres
DB_PASSWORD=Duong19082k3@
DB_NAME=postgres
DB_PORT=5432
DB_SSLMODE=require
PORT=8080
GIN_MODE=release
```

---

## 🐳 Cách 1: Deploy với Railway (Recommended)

### Bước 1: Chuẩn bị
```bash
# Kiểm tra Docker
docker --version
# Kiểm tra Railway CLI
npm install -g @railway/cli
# Hoặc qua: https://railway.app
```

### Bước 2: Login Railway
```bash
railway login
```

### Bước 3: Deploy
```bash
# Tại folder backend
railway up
```

### Bước 4: Quản lý Environment Variables
- Vào https://railway.app
- Chọn project
- Vào Settings → Variables
- Thêm các biến từ `.env`

---

## 🦅 Cách 2: Deploy với Vercel + Go

### Bước 1: Chuẩn bị
```bash
npm install -g vercel
```

### Bước 2: Deploy
```bash
vercel --prod
```

### Bước 3: Thêm Environment Variables
- Vào Vercel Dashboard
- Settings → Environment Variables
- Thêm biến theo `DB_*`

---

## 🏠 Cách 3: Deploy Local/Server riêng

### Bước 1: Build
```bash
go build -o api cmd/api/main.go
```

### Bước 2: Chạy
```bash
# Set biến môi trường
export DB_HOST=db.ofikztulwfiufxjymrsz.supabase.co
export DB_USER=postgres
export DB_PASSWORD=Duong19082k3@
export DB_NAME=postgres
export DB_PORT=5432
export DB_SSLMODE=require
export PORT=8080
export GIN_MODE=release

# Chạy
./api
```

---

## 🐳 Cách 4: Docker Local

### Build Docker Image
```bash
docker build -t movie-backend:latest .
```

### Chạy Container
```bash
docker run -d \
  -p 8080:8080 \
  -e DB_HOST=db.ofikztulwfiufxjymrsz.supabase.co \
  -e DB_USER=postgres \
  -e DB_PASSWORD=Duong19082k3@ \
  -e DB_NAME=postgres \
  -e DB_PORT=5432 \
  -e DB_SSLMODE=require \
  -e GIN_MODE=release \
  --name movie-backend \
  movie-backend:latest
```

### Kiểm tra
```bash
curl http://localhost:8080/api/health
```

---

## ✅ Kiểm tra sau khi deploy

### 1. Health Check
```bash
curl https://your-deployed-url/api/health
```

### 2. Database Connection
```bash
curl https://your-deployed-url/api/v1/movies?page=1&limit=10
```

### 3. Logs
- **Railway**: `railway logs`
- **Vercel**: Xem trong Vercel Dashboard
- **Docker**: `docker logs movie-backend`

---

## 🔧 Troubleshooting

### Lỗi: "Failed to connect database"
- ✅ Kiểm tra `DB_SSLMODE=require` (quan trọng cho Supabase)
- ✅ Kiểm tra IP được whitelist trong Supabase
- ✅ Kiểm tra username/password đúng

### Lỗi: "Connection refused"
- ✅ Kiểm tra database URL chính xác
- ✅ Kiểm tra port 5432 (Supabase mặc định)

### Lỗi: "Extension uuid-ossp"
- ✅ Warning này bình thường (extension có thể không cần)
- Nếu cần, enable trong Supabase: Extensions → uuid-ossp

---

## 📝 Notes
- Default PORT: `8080`
- Default MODE: `release` (không debug logs)
- SSL: **REQUIRED** cho Supabase
- Auto Migration: Chạy tự động khi khởi động
