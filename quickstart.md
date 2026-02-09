# Quick Start Guide

Get up and running with the WhatsApp Chatbot API in 5 minutes!

## Prerequisites

- Go 1.21+ installed
- MySQL running
- Git

## Step 1: Clone and Setup

```bash
# Clone repository
git clone <your-repo>
cd whatsapp-chatbot-go

# Install dependencies
go mod download

# Copy environment file
cp .env.example .env
```

## Step 2: Configure Database

Edit `.env`:

```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=whatsapp_chatbot
JWT_SECRET=change-this-to-something-secure
```

Create database:

```bash
mysql -u root -p -e "CREATE DATABASE whatsapp_chatbot CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

## Step 3: Run Application

```bash
# Run directly
go run cmd/server/main.go

# Or build and run
make build
./bin/server
```

You should see:
```
✅ Database connected successfully
✅ Database migrations completed
🚀 WhatsApp API Server running on port 3456
```

## Step 4: Test the API

```bash
# Health check
curl http://localhost:3456/api/health

# Should return:
# {"status":"ok","timestamp":"2024-..."}
```

## Step 5: Create Your First Chatbot

### 5.1 Generate JWT Token

For testing, you can use this simple token generator:

```go
// Save as generate_token.go
package main

import (
	"fmt"
	"time"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID    uint   `json:"user_id"`
	ChatbotID string `json:"chatbot_id"`
	jwt.RegisteredClaims
}

func main() {
	claims := Claims{
		UserID:    1,
		ChatbotID: "",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour * 30)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("change-this-secret-key"))
	
	fmt.Println("Your JWT Token:")
	fmt.Println(tokenString)
}
```

Run it:
```bash
go run generate_token.go
```

### 5.2 Initialize WhatsApp Session

```bash
export TOKEN="your-jwt-token-here"

curl -X POST http://localhost:3456/api/session/init \
  -H "Authorization: Bearer $TOKEN"
```

### 5.3 Get QR Code

```bash
curl http://localhost:3456/api/session/qr/1 > qr.json
cat qr.json
```

The response contains a base64 QR code. Decode it and scan with WhatsApp!

### 5.4 Check Session Status

```bash
curl http://localhost:3456/api/session/status/1
```

Wait until you see:
```json
{
  "status": "ready",
  "is_logged_in": true
}
```

### 5.5 Create Chatbot

```bash
curl -X POST http://localhost:3456/api/chatbot \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "919876543210@s.whatsapp.net",
    "welcomeMessage": "👋 Welcome! How can I help you?\n\n1️⃣ Services\n2️⃣ Pricing\n3️⃣ Contact",
    "isActive": true
  }'
```

### 5.6 Add FAQ Options

```bash
# Option 1: Services
curl -X POST http://localhost:3456/api/chatbot/option \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "optionKey": "1",
    "optionLabel": "Services",
    "answer": "We offer:\n✅ Web Development\n✅ Mobile Apps\n✅ Cloud Solutions\n✅ AI Integration",
    "order": 1
  }'

# Option 2: Pricing
curl -X POST http://localhost:3456/api/chatbot/option \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "optionKey": "2",
    "optionLabel": "Pricing",
    "answer": "Our pricing starts at:\n💰 Web: $1000+\n💰 Mobile: $2000+\n💰 Custom: Contact us!",
    "order": 2
  }'

# Option 3: Contact
curl -X POST http://localhost:3456/api/chatbot/option \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "optionKey": "3",
    "optionLabel": "Contact",
    "answer": "📞 Phone: +1-234-567-8900\n📧 Email: hello@company.com\n🌐 Web: www.company.com",
    "order": 3
  }'
```

## Step 6: Test Your Chatbot

Send a WhatsApp message to your number:
- Type: `hi` or `hello` → Get welcome message
- Type: `1` → Get services info
- Type: `2` → Get pricing info
- Type: `3` → Get contact info

## Step 7: Send Messages (Optional)

```bash
# Send single message
curl -X POST http://localhost:3456/api/message/send \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "1",
    "phone": "919876543210",
    "message": "Hello from Go API!"
  }'

# Send bulk messages
curl -X POST http://localhost:3456/api/message/send-many \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "1",
    "phones": ["919876543210", "919876543211"],
    "message": "Bulk message test!"
  }'
```

## Using Docker (Alternative)

If you prefer Docker:

```bash
# Start everything
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

## Troubleshooting

### Can't connect to MySQL?
```bash
# Check MySQL is running
sudo systemctl status mysql

# Or start it
sudo systemctl start mysql
```

### Port 3456 already in use?
```bash
# Find what's using it
lsof -i :3456

# Change port in .env
PORT=3457
```

### QR code not appearing?
- Wait 10-20 seconds after init
- Check sessions directory exists and is writable
- Try restarting the application

### Messages not sending?
- Verify session status is "ready"
- Check phone number format (include country code)
- Ensure rate limits not exceeded

## Next Steps

- 📖 Read full [README.md](README.md)
- 🔧 Check [MIGRATION.md](MIGRATION.md) if migrating from Node.js
- 🎯 Explore API endpoints
- 🚀 Deploy to production

## Production Checklist

Before going to production:

- [ ] Change JWT_SECRET to a strong secret
- [ ] Set ENV=production
- [ ] Enable HTTPS
- [ ] Set up monitoring
- [ ] Configure backups
- [ ] Set resource limits
- [ ] Enable rate limiting
- [ ] Review security settings

## Need Help?

- Check logs: Look for errors in console output
- Test endpoints: Use the `/api/health` endpoint
- Verify database: Check MySQL connection
- Review config: Ensure .env values are correct

Happy coding! 🎉