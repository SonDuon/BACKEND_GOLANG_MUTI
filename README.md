
## Các tính năng

### 1. Trang chủ (Home)
- Hiển thị danh sách phim với phân trang
- Loading spinner khi đang tải dữ liệu
- Xử lý lỗi kết nối
- Responsive design với Bootstrap 5

### 2. Chi tiết phim (Movie Detail)
- Hiển thị thông tin chi tiết: poster, tiêu đề, mô tả, thể loại, quốc gia
- Nút "Xem ngay" để chuyển sang chế độ xem phim

### 3. Xem phim (Watch)
- Video player với HTML5 video tag
- Chọn tập phim (nếu có nhiều tập)
- Tự động phát video

### 4. Admin Panel
- Kiểm tra tình trạng các providers (ophim1, self-hosted, etc.)
- Hiển thị priority của từng provider
- Nút refresh để kiểm tra lại

## API Endpoints được sử dụng

| Method | Endpoint | Mô tả |
|--------|----------|-------|
| GET | `/api/v1/movies?page=1&limit=24` | Lấy danh sách phim |
| GET | `/api/v1/movies/:slug` | Lấy chi tiết phim |
| GET | `/api/v1/movies/:slug/watch?episode=full` | Lấy link xem phim |
| GET | `/api/v1/admin/providers/health` | Kiểm tra health providers |

## Cách chạy

### Bước 1: Khởi động Backend
```bash
cd /workspace
go run cmd/api/main.go
```

Backend sẽ chạy tại `http://localhost:8080`

### Bước 2: Chạy Frontend

**Option 1: Dùng Python Simple HTTP Server**
```bash
cd /workspace/frontend
python3 -m http.server 3000
```

Sau đó mở trình duyệt truy cập: `http://localhost:3000`

**Option 2: Dùng Live Server (VS Code Extension)**
- Cài đặt extension "Live Server" trong VS Code
- Click chuột phải vào `index.html` và chọn "Open with Live Server"

**Option 3: Dùng Node.js http-server**
```bash
npm install -g http-server
cd /workspace/frontend
http-server -p 3000
```

## Lưu ý quan trọng

### CORS
Backend đã được cấu hình CORS để cho phép frontend gọi API từ domain khác. Nếu gặp lỗi CORS, kiểm tra middleware trong `cmd/api/main.go`:

```go
r.Use(cors.New(cors.Config{
    AllowOriginFunc:  func(origin string) bool { return true },
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
    ExposeHeaders:    []string{"Content-Length", "Authorization"},
    AllowCredentials: true,
    MaxAge:           12 * time.Hour,
}))
```

### API Base URL
Trong `app.js`, thay đổi `API_BASE_URL` nếu backend chạy trên port khác:

```javascript
var API_BASE_URL = 'http://localhost:8080/api/v1';
```

## Công nghệ sử dụng

- **AngularJS 1.8.2** - JavaScript framework
- **Bootstrap 5.3** - CSS framework
- **Font Awesome 6.4** - Icons
- **HTML5 Video** - Video player

## Mở rộng

Để thêm tính năng mới:

1. **Search**: Backend cần hỗ trợ query parameter `?search=keyword`
2. **Filter theo thể loại**: Thêm endpoint `/movies?category=action`
3. **User authentication**: Thêm login/register pages
4. **Favorite movies**: Lưu phim yêu thích vào localStorage

## Troubleshooting

### Lỗi: "Cannot load movies"
- Kiểm tra backend đã chạy chưa
- Kiểm tra CORS configuration
- Mở Console (F12) để xem lỗi chi tiết

### Lỗi: "Video cannot be played"
- Kiểm tra định dạng video (MP4, M3U8)
- Kiểm tra link có accessible không
- Thử trên browser khác

---
© 2024 Movie Streaming - Built with AngularJS & Golang
