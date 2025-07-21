
# Hướng Dẫn Sử Dụng WebSocket Theo Dõi Trạng Thái Tài Liệu

Hệ thống hỗ trợ kết nối **WebSocket** để theo dõi **trạng thái xử lý tài liệu theo thời gian thực** sau khi người dùng tải lên file (PDF, DOCX, TXT). Tính năng này giúp frontend hiển thị quá trình như:

- Đang trích xuất nội dung
- Đang làm sạch dữ liệu
- Đang tạo audio
- Hoàn tất xử lý

---

## Kiến trúc

```
Client (WebSocket) <---> Backend (Golang) <--> Trình xử lý tài liệu (services/*)
```

- Mỗi client sẽ kết nối tới đường dẫn WebSocket có dạng:
  ```
  /ws/doc/:documentID?token=JWT
  ```
- Khi backend xử lý tài liệu, hệ thống sẽ gửi thông báo trạng thái qua WebSocket tới các client đang theo dõi document đó.

---

## Yêu cầu để sử dụng

1. Backend đã được deploy (`localhost:8081` hoặc domain Railway)
2. Người dùng đã đăng nhập và có `JWT Token`
3. Đã có một tài liệu hợp lệ được upload để lấy `documentID`. (Vừa bấm gửi api xong thì có thể vào database để lấy id document để vào cmd chạy websocket test)

---

## Cách kết nối để test (dùng wscat)

### 1. Cài đặt công cụ

```bash
npm install -g wscat
```

### 2. Kết nối WebSocket

Cú pháp:

```bash
wscat -c "ws://localhost:8081/ws/doc/{documentID}?token={JWT}"
```

Ví dụ:

```bash
wscat -c "ws://localhost:8081/ws/doc/3ff86315-a9e3-4cd7-ac1a-78d99299f3c9?token=eyJhbGciOiJIUzI1NiIsInR5..."
```

> Nếu đang dùng Railway:
```bash
wscat -c "wss://podcastserver-production.up.railway.app/ws/doc/{documentID}?token={JWT}"
```
---

## Cách frontend kết nối WebSocket

```ts
const socket = new WebSocket(
  `wss://podcast-server-production.up.railway.app/ws/doc/${docID}?token=${token}`
);

socket.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log("Trạng thái xử lý:", data.status);
};

socket.onopen = () => console.log("Đã kết nối WebSocket");
socket.onclose = () => console.log("Đã ngắt kết nối WebSocket");
```

> Nếu deploy thì dùng `wss://` thay vì `ws://`

---

## Format tin nhắn nhận được

Mỗi khi trạng thái xử lý thay đổi, client sẽ nhận được:

```json
{
  "status": "Đang đã trích xuất"
}
```

# Danh sách trạng thái WebSocket khi xử lý tài liệu

Dưới đây là danh sách các trạng thái mà backend sẽ gửi qua WebSocket trong quá trình xử lý một tài liệu. Mỗi trạng thái phản ánh tiến trình thực hiện từng bước trong pipeline upload - xử lý - tạo audio.

## Trạng thái chi tiết

- **Đang tải lên tài liệu...**  
  Gửi ngay khi hệ thống bắt đầu upload file từ client.

- **Lỗi khi tải lên Supabase**  
  Gửi nếu quá trình upload file lên Supabase thất bại.

- **Không thể lưu tài liệu vào database**  
  Gửi khi có lỗi trong lúc lưu thông tin tài liệu vào cơ sở dữ liệu.

- **Đã tải lên**  
  Gửi khi file đã được upload thành công và lưu vào database.

- **Đang trích xuất nội dung...**  
  Gửi khi bắt đầu đọc nội dung văn bản từ file.

- **Lỗi khi trích xuất nội dung**  
  Gửi nếu có lỗi trong quá trình phân tích/trích xuất văn bản từ file.

- **Đang làm sạch nội dung...**  
  Gửi khi bắt đầu xử lý, chuẩn hóa và làm sạch nội dung văn bản.

- **Lỗi khi làm sạch nội dung**  
  Gửi nếu xảy ra lỗi trong bước xử lý nội dung.

- **Đã trích xuất**  
  Gửi khi hệ thống đã trích xuất và làm sạch nội dung xong.

- **Đang tạo audio...**  
  Gửi khi hệ thống đang dùng Google Text-to-Speech để tạo giọng đọc từ nội dung.

- **Lỗi khi tạo audio**  
  Gửi nếu gọi API TTS thất bại hoặc sinh audio lỗi.

- **Đang lưu audio...**  
  Gửi khi hệ thống bắt đầu upload file audio lên Supabase.

- **Lỗi upload audio**  
  Gửi nếu upload file âm thanh bị lỗi.

- **Hoàn thành xử lý tài liệu**  
  Gửi cuối cùng khi mọi công đoạn hoàn tất thành công: upload file, trích xuất nội dung, làm sạch, tạo audio và lưu audio.
---

## Lỗi thường gặp

- 401 Unauthorized: Token sai hoặc thiếu. Kiểm tra lại token.
- connection refused: Backend chưa chạy. Khởi động lại server.
- write EPROTO: Dùng ws:// trong khi server yêu cầu wss://. Hãy dùng wss:// nếu deploy trên Railway hoặc HTTPS.


---

