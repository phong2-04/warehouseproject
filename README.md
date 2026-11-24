# HỆ THỐNG QUẢN LÝ KHO HÀNG (WAREHOUSE MANAGEMENT SYSTEM)

## 1. Giới thiệu

Warehouse Management System (WMS) là hệ thống quản lý kho hàng hỗ trợ quản lý sản phẩm, danh mục hàng hóa, số lượng tồn kho và các hoạt động nhập – xuất kho.

Dự án được xây dựng theo mô hình Full-stack, sử dụng Go cho Backend, React cho Frontend và PostgreSQL làm hệ quản trị cơ sở dữ liệu. Hệ thống tích hợp xác thực JWT, phân quyền người dùng và Docker để triển khai ứng dụng.

## 2. Chức năng chính

### 2.1. Xác thực và phân quyền
- Đăng nhập bằng tài khoản người dùng.
- Xác thực bằng JWT Access Token và Refresh Token.
- Hash mật khẩu bằng BCrypt.
- Phân quyền Admin và Staff.
- Kiểm tra quyền truy cập các chức năng.
- Đăng xuất và thu hồi Refresh Token.

### 2.2. Quản lý sản phẩm
- Xem danh sách sản phẩm.
- Thêm, sửa và xóa sản phẩm.
- Quản lý mã sản phẩm, tên, giá và danh mục.
- Theo dõi số lượng tồn kho.
- Thiết lập mức tồn kho tối thiểu.
- Tìm kiếm và lọc sản phẩm.

### 2.3. Quản lý danh mục
- Xem danh sách danh mục.
- Thêm, sửa và xóa danh mục.
- Kiểm tra dữ liệu đầu vào.
- Kiểm tra tên danh mục trùng lặp.

### 2.4. Quản lý nhập – xuất kho
- Thực hiện nhập kho.
- Thực hiện xuất kho.
- Kiểm tra số lượng tồn trước khi xuất.
- Lưu lịch sử giao dịch kho.
- Theo dõi người thực hiện và thời gian giao dịch.

### 2.5. Quản lý người dùng
- Xem danh sách người dùng.
- Thêm, cập nhật và xóa tài khoản.
- Phân quyền Admin và Staff.
- Kiểm soát các thao tác quản trị quan trọng.

### 2.6. Dashboard
- Hiển thị tổng quan tình hình kho hàng.
- Theo dõi số lượng sản phẩm.
- Theo dõi tình trạng tồn kho.
- Hiển thị thông tin thống kê phục vụ quản lý.

## 3. Công nghệ sử dụng

| Thành phần | Công nghệ |
|---|---|
| Backend | Go |
| HTTP API | Go net/http |
| Frontend | React |
| Build Frontend | Vite |
| Giao diện | Bootstrap, CSS |
| Database | PostgreSQL |
| Database Driver | pgx |
| Authentication | JWT |
| Password Hashing | BCrypt |
| Database Migration | golang-migrate |
| Containerization | Docker |
| Container Orchestration | Docker Compose |
| Web Server | Nginx |

## 4. Kiến trúc hệ thống

Hệ thống được tổ chức thành các thành phần:

- **Frontend:** React cung cấp giao diện và gửi HTTP Request đến Backend.
- **Backend:** Go xử lý nghiệp vụ, xác thực, phân quyền và truy vấn dữ liệu.
- **Database:** PostgreSQL lưu trữ thông tin người dùng, sản phẩm, danh mục và lịch sử giao dịch.
- **Docker Compose:** Quản lý và khởi chạy các thành phần của hệ thống.

### Luồng xử lý

```text
        USER
         |
         v
   REACT FRONTEND
         |
      HTTP API
         |
         v
     GO BACKEND
         |
   -----------------
   |       |       |
Handler  Service  Middleware
   |
Repository
   |
   v
 POSTGRESQL
```

## 5. Cấu trúc thư mục

```text
warehouseproject/
│
├── cmd/
│   ├── api/
│   └── hashpassword/
│
├── internal/
│   ├── database/
│   ├── handler/
│   ├── middleware/
│   ├── model/
│   ├── repository/
│   └── service/
│
├── migrations/
│
├── frontend/
│   ├── src/
│   │   ├── api/
│   │   ├── components/
│   │   └── App.jsx
│   ├── Dockerfile
│   └── nginx.conf
│
├── Dockerfile
├── docker-compose.yml
├── .dockerignore
├── .env.example
├── .gitignore
└── go.mod
```

## 6. Cơ sở dữ liệu

Hệ thống sử dụng PostgreSQL với các bảng chính:

| Bảng | Chức năng |
|---|---|
| users | Lưu thông tin tài khoản và phân quyền |
| categories | Lưu danh mục sản phẩm |
| products | Lưu thông tin sản phẩm và tồn kho |
| stock_transactions | Lưu lịch sử nhập – xuất kho |
| refresh_tokens | Quản lý Refresh Token |

### Database Migration

Dự án sử dụng Migration để quản lý thay đổi cấu trúc cơ sở dữ liệu.

Các phiên bản Migration hiện tại:

- `000001`: Khởi tạo cấu trúc cơ sở dữ liệu.
- `000002`: Bổ sung bảng quản lý Refresh Token.
- `000003`: Bổ sung các ràng buộc cho dữ liệu nhập – xuất kho.

## 7. Hướng dẫn cài đặt và chạy dự án

### 7.1. Yêu cầu môi trường

- Git
- Docker Desktop
- Docker Compose

### 7.2. Clone dự án

```bash
git clone https://github.com/phong2-04/warehouseproject.git
cd warehouseproject
```

### 7.3. Cấu hình môi trường

Tạo file `.env` từ file mẫu:

```powershell
Copy-Item .env.example .env
```

Cấu hình các biến môi trường trong file `.env`:

```env
DB_USER=postgres
DB_PASSWORD=your_database_password
DB_NAME=warehouse_db
JWT_SECRET=your_random_secret_at_least_32_bytes
```

Lưu ý:
- Sử dụng mật khẩu database riêng.
- JWT_SECRET phải có độ dài tối thiểu 32 byte.
- Không đưa file `.env` lên GitHub.

### 7.4. Khởi chạy hệ thống

```bash
docker compose up -d --build
```

Docker Compose sẽ khởi chạy:
1. PostgreSQL Database.
2. Database Migration.
3. Go Backend API.
4. React Frontend sử dụng Nginx.

### 7.5. Truy cập ứng dụng

| Thành phần | Địa chỉ |
|---|---|
| Frontend | http://localhost:5173 |
| Backend API | http://localhost:8080 |
| PostgreSQL | localhost:5433 |

### 7.6. Kiểm tra trạng thái container

```bash
docker compose ps
```

Xem log Backend:

```bash
docker compose logs -f backend
```

Xem log Database:

```bash
docker compose logs -f db
```

### 7.7. Kiểm tra tình trạng hoạt động (Health Check)

Backend cung cấp endpoint `/health` để kiểm tra trạng thái API và kết nối PostgreSQL.

Kiểm tra trực tiếp:

```bash
curl http://localhost:8080/health
```

Kết quả khi hệ thống hoạt động bình thường:

```json
{
  "database": "connected",
  "status": "healthy"
}
```

Docker Compose tự động kiểm tra tình trạng Backend định kỳ.

Kiểm tra trạng thái:

```bash
docker compose ps
```

Backend có trạng thái `healthy` khi API và kết nối PostgreSQL hoạt động bình thường.

### 7.8. Dừng hệ thống

```bash
docker compose down
```

Lệnh trên dừng các container nhưng giữ lại dữ liệu PostgreSQL trong Docker volume.

Nếu muốn xóa cả volume dữ liệu:

```bash
docker compose down -v
```

**Cảnh báo:** Lệnh này xóa dữ liệu database được lưu trong Docker volume.

## 8. Các API chính

| Nhóm chức năng | Endpoint | Phương thức |
|---|---|---|
| Authentication | `/api/login` | POST |
| Authentication | `/api/refresh-token` | POST |
| Authentication | `/api/logout` | POST |
| User | `/api/profile` | GET |
| Category | `/api/categories` | GET |
| Product | `/api/products` | GET |
| Stock | `/api/stock/import` | POST |
| Stock | `/api/stock/export` | POST |

Các API yêu cầu xác thực sử dụng Access Token trong HTTP Authorization Header:

```http
Authorization: Bearer <access_token>
```

## 9. Kiểm thử

Chạy kiểm thử Backend:

```bash
go test ./...
```

Kiểm tra độ bao phủ kiểm thử:

```bash
go test ./... -cover
```

Kiểm tra build:

```bash
go build ./...
```

## 10. Định hướng phát triển

Một số hướng có thể mở rộng trong tương lai:

- Hoàn thiện báo cáo thống kê nhập – xuất – tồn.
- Bổ sung xuất báo cáo Excel.
- Bổ sung quản lý nhà cung cấp.
- Bổ sung quản lý đơn hàng nhập và xuất.
- Tối ưu hiệu năng truy vấn khi dữ liệu tăng.
- Mở rộng kiểm thử tự động cho Backend và Frontend.

## 11. Tác giả

**Tên dự án:** Warehouse Management System (WMS)

**Mục đích:** Dự án thực hành và phát triển kỹ năng lập trình Full-stack, xây dựng Backend API, quản lý cơ sở dữ liệu và triển khai ứng dụng bằng Docker.

**GitHub:** https://github.com/phong2-04/warehouseproject