# 🔐 Authentication Guide - Login/Register Mekanizması

Bu döküman, uygulamanın kimlik doğrulama sisteminin nasıl çalıştığını junior geliştiriciler için açıklar.

## 📋 İçindekiler
1. [Genel Mimari](#genel-mimari)
2. [User Registration (Kullanıcı Kaydı)](#user-registration)
3. [User Login (Giriş)](#user-login)
4. [JWT Token Sistemi](#jwt-token-sistemi)
5. [Token Validasyonu](#token-validasyonu)
6. [Protected Endpoints](#protected-endpoints)
7. [Kod Örnekleri](#kod-örnekleri)
8. [Güvenlik Özellikleri](#güvenlik-özellikleri)
9. [Hata Yönetimi](#hata-yönetimi)
10. [Test Etme](#test-etme)

## 🏗️ Genel Mimari

Authentication sistemimiz şu katmanlardan oluşur:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Controller    │───▶│     Service     │───▶│   Repository    │
│ (HTTP Requests) │    │ (Business Logic)│    │   (Database)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
    ┌────▼────┐            ┌─────▼──────┐         ┌──────▼───────┐
    │   JWT   │            │  Password  │         │    Users     │
    │Middleware│           │  Hashing   │         │    Table     │
    └─────────┘            └────────────┘         └──────────────┘
```

## 👥 User Registration

### 1. Registration Flow

```mermaid
sequenceDiagram
    participant Client
    participant Controller
    participant Service
    participant Repository
    participant Database

    Client->>Controller: POST /api/v1/auth/register
    Controller->>Service: Register(username, email, password, firstName, lastName)
    Service->>Service: Validate input data
    Service->>Repository: Check if username/email exists
    Repository->>Database: SELECT from users
    Database->>Repository: Result
    Repository->>Service: User exists/not exists
    Service->>Service: Hash password (Argon2)
    Service->>Repository: AddUser(hashedUser)
    Repository->>Database: INSERT user
    Database->>Repository: Success/Error
    Repository->>Service: Result
    Service->>Controller: Success/Error
    Controller->>Client: 201 Created / Error Response
```

### 2. Registration Validation

Kayıt sırasında şu validasyonlar yapılır:

```go
// Username validasyonu
- Minimum 3 karakter
- Sadece alfanumerik karakterler ve boşluk
- Boş olamaz

// Email validasyonu
- Geçerli email formatı (regex ile kontrol)
- Veritabanında benzersiz olmalı

// Password validasyonu  
- Minimum 6 karakter
- Boş olamaz

// İsim validasyonları
- First name ve last name boş olamaz
- Alfanumerik karakterler ve boşluk
```

### 3. Password Hashing

Güvenlik için **Argon2** algoritması kullanılır:

```go
func hashPassword(password string) (string, error) {
    // 16 byte random salt
    salt := make([]byte, 16)
    rand.Read(salt)
    
    // Argon2 ID with recommended parameters
    hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
    
    // Format: $argon2id$v=19$m=65536,t=1,p=4$salt$hash
    return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
        argon2.Version, 64*1024, 1, 4,
        base64.RawStdEncoding.EncodeToString(salt),
        base64.RawStdEncoding.EncodeToString(hash)), nil
}
```

## 🔑 User Login

### 1. Login Flow

```mermaid
sequenceDiagram
    participant Client
    participant Controller
    participant Service
    participant Repository
    participant Middleware

    Client->>Controller: POST /api/v1/auth/login
    Controller->>Service: Login(usernameOrEmail, password)
    Service->>Repository: GetByUsername/GetByEmail
    Repository->>Service: User data
    Service->>Service: Verify password (Argon2)
    Service->>Controller: User data (if valid)
    Controller->>Middleware: GenerateToken(userId, username, email)
    Middleware->>Controller: JWT token
    Controller->>Client: 200 OK + JWT token + user data
```

### 2. Login Validation

```go
// Giriş validasyonu
1. Username/Email ve password boş olamaz
2. Email içeriyorsa "@" -> email ile ara
3. Email içermiyorsa -> username ile ara
4. Kullanıcı bulunamazsa -> "invalid credentials"
5. Password doğru değilse -> "invalid credentials"
6. Her şey OK ise -> JWT token oluştur
```

## 🎫 JWT Token Sistemi

### 1. Token Structure

JWT tokenımız şu bilgileri içerir:

```go
type Claims struct {
    UserId   int64  `json:"user_id"`    // Kullanıcı ID'si
    Username string `json:"username"`   // Kullanıcı adı
    Email    string `json:"email"`      // Email adresi
    jwt.RegisteredClaims                // Standard claims (exp, iat)
}
```

### 2. Token Generation

```go
func GenerateToken(userId int64, username, email string) (string, error) {
    expirationTime := time.Now().Add(24 * time.Hour) // 24 saat geçerli
    
    claims := &Claims{
        UserId:   userId,
        Username: username,
        Email:    email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}
```

### 3. Token Format

Client'a dönen response formatı:

```json
{
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "johndoe",
    "email": "john@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "created_at": "2024-01-01T10:00:00Z",
    "updated_at": "2024-01-01T10:00:00Z"
  }
}
```

## 🛡️ Token Validasyonu

### 1. JWT Middleware

Protected endpoint'lere erişim için JWT middleware kullanılır:

```go
func JWTMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // 1. Authorization header'ını al
            authHeader := c.Request().Header.Get("Authorization")
            
            // 2. Bearer format kontrolü
            tokenString := strings.TrimPrefix(authHeader, "Bearer ")
            
            // 3. Token'ı parse et ve doğrula
            claims := &Claims{}
            token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
                return jwtSecret, nil
            })
            
            // 4. Token geçerliyse kullanıcı bilgilerini context'e kaydet
            c.Set("user_id", claims.UserId)
            c.Set("username", claims.Username)
            c.Set("email", claims.Email)
            
            return next(c)
        }
    }
}
```

### 2. Validation Steps

Token validasyonu şu adımları takip eder:

1. **Header Kontrolü**: `Authorization: Bearer <token>` formatında mı?
2. **Token Parse**: JWT formatında parse edilebiliyor mu?
3. **Signature Kontrolü**: Token imzası geçerli mi?
4. **Expiry Kontrolü**: Token süresi dolmuş mu?
5. **Claims Kontrolü**: Gerekli claim'ler var mı?

### 3. User Ownership Tracking (Kullanıcı Sahipliği Takibi)

Sistemimizde **hangi kullanıcının hangi product'ı oluşturduğu** bilgisi JWT token üzerinden otomatik olarak takip edilir. Bu mekanizma güvenliği sağlar ve kullanıcıların sadece kendi verilerine erişmesini garanti eder.

#### 📊 User Ownership Flow

```mermaid
sequenceDiagram
    participant Client
    participant Middleware
    participant Controller
    participant Service
    participant Repository
    participant Database

    Client->>Middleware: POST /products + JWT Token
    Middleware->>Middleware: Parse JWT → Extract user_id
    Middleware->>Controller: Request + Context(user_id)
    Controller->>Controller: c.Get("user_id") → userId
    Controller->>Service: Add(productData, userId)
    Service->>Repository: AddProduct(product + userId)
    Repository->>Database: INSERT products (..., user_id)
    Database->>Client: Product created with owner tracking
```

#### 🔍 Detaylı Implementation

**1. JWT Token'dan UserID Extraction:**

```go
// middleware/auth.go - JWT Middleware
func JWTMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Token'ı parse et
            claims := &Claims{}
            token, err := jwt.ParseWithClaims(tokenString, claims, ...)
            
            // 🎯 Kullanıcı bilgilerini context'e kaydet
            c.Set("user_id", claims.UserId)     // ← Bu bilgi product oluştururken kullanılır
            c.Set("username", claims.Username)
            c.Set("email", claims.Email)
            
            return next(c)
        }
    }
}
```

**2. Controller'da UserID Okuma:**

```go
// controller/product_controller.go
func (productController *ProductController) AddProduct(c echo.Context) error {
    // 🔑 JWT middleware tarafından context'e kaydedilen user_id'yi oku
    userIdInterface := c.Get("user_id")
    userId, ok := userIdInterface.(int64)
    if !ok {
        // Token geçersiz veya user_id yok
        return c.JSON(http.StatusUnauthorized, response.ErrorResponse{
            ErrorDescription: "Invalid user authentication",
        })
    }

    // Request body'den product bilgilerini al
    var addProductRequest request.AddProductRequest
    c.Bind(&addProductRequest)
    
    // 🎯 Service'e hem product data'sını hem de userId'yi gönder
    err := productController.productService.Add(addProductRequest.ToModel(), userId)
    //                                                                        ↑
    //                                                           JWT'den otomatik alınan userID
    
    if err != nil {
        return c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse{
            ErrorDescription: err.Error(),
        })
    }
    return c.NoContent(http.StatusCreated)
}
```

**3. Service Layer'da Ownership Assignment:**

```go
// service/product_service.go
func (productService *ProductService) Add(productCreate model.ProductCreate, userId int64) error {
    // Input validasyonları
    validateError := validateProductCreate(productCreate)
    if validateError != nil {
        return validateError
    }
    
    // 🏷️ Domain Product nesnesi oluştururken userId'yi otomatik ata
    return productService.productRepository.AddProduct(domain.Product{
        Name:        productCreate.Name,
        Price:       productCreate.Price,
        Description: productCreate.Description,
        Discount:    productCreate.Discount,
        Store:       productCreate.Store,
        ImageUrls:   productCreate.ImageUrls,
        CategoryID:  productCreate.CategoryID,
        UserID:      userId,  // ← JWT'den gelen userID burada assign edilir
    })
}
```

**4. Database'e Ownership ile Kayıt:**

```go
// persistence/product_repository.go
func (productRepository *ProductRepository) AddProduct(product domain.Product) error {
    ctx := context.Background()

    // 💾 SQL query'de user_id'yi de dahil et
    insertProductSQL := `
        INSERT INTO products (name, price, description, discount, store, category_id, user_id)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id;
    `

    var productId int64
    err := productRepository.dbPool.QueryRow(ctx, insertProductSQL,
        product.Name, product.Price, product.Description, 
        product.Discount, product.Store, product.CategoryID, 
        product.UserID).Scan(&productId)  // ← user_id database'e kaydediliyor

    if err != nil {
        return fmt.Errorf("failed to insert product: %w", err)
    }
    
    // Resim URL'lerini de user'a bağlı olarak kaydet
    for i, url := range product.ImageUrls {
        // product_images tablosunda da user ownership tracking yapılabilir
    }

    return nil
}
```

#### 🛡️ Security Benefits (Güvenlik Faydaları)

**1. Client UserID Manipulation Koruması:**

```bash
# ❌ Bu çalışmaz - Client userID gönderemez/değiştiremez
curl -X POST /api/v1/products \
  -H "Authorization: Bearer <valid_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Malicious Product",
    "price": 100,
    "user_id": 999,          # ← Bu ignored edilir
    "fake_user": "admin"     # ← Bu da ignored edilir
  }'

# ✅ Bu çalışır - UserID JWT'den otomatik alınır
curl -X POST /api/v1/products \
  -H "Authorization: Bearer <valid_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Legitimate Product",
    "price": 100,
    "category_id": 1
  }'
# userID otomatik olarak JWT'den extract edilir ve assign edilir
```

**2. Ownership Validation:**

```go
// Her zaman token'dan gelen user_id kullanılır
func (controller *ProductController) UpdateProduct(c echo.Context) error {
    // Token'dan user_id al
    tokenUserId := c.Get("user_id").(int64)
    
    // Product sahibini kontrol et
    product, err := controller.service.GetById(productId)
    if err != nil {
        return err
    }
    
    // 🔒 Sadece product sahibi güncelleyebilir
    if product.UserID != tokenUserId {
        return c.JSON(http.StatusForbidden, map[string]string{
            "error": "You can only update your own products",
        })
    }
    
    // Update işlemine devam et...
}
```

#### 🎯 User-Specific Operations

**1. Kullanıcının Kendi Products'larını Getirme:**

```go
// GET /api/v1/products/my-products endpoint'i
func (productController *ProductController) GetMyProducts(c echo.Context) error {
    // JWT'den user_id otomatik al
    userIdInterface := c.Get("user_id")
    userId, ok := userIdInterface.(int64)
    if !ok {
        return c.JSON(http.StatusUnauthorized, response.ErrorResponse{
            ErrorDescription: "Invalid user authentication",
        })
    }

    // 🎯 Sadece bu kullanıcının products'larını getir
    products := productController.productService.GetAllProductsByUser(userId)
    return c.JSON(http.StatusOK, response.ToResponseList(products))
}
```

**2. Database Query with User Filtering:**

```go
// Repository layer - User-specific queries
func (productRepository *ProductRepository) GetAllProductsByUser(userId int64) []domain.Product {
    ctx := context.Background()

    // 🔍 WHERE user_id = $1 ile sadece o kullanıcının products'larını getir
    getProductByUserSql := `
        SELECT id, name, price, description, discount, store, category_id, user_id
        FROM products
        WHERE user_id = $1
        ORDER BY created_at DESC
    `

    productRows, err := productRepository.dbPool.Query(ctx, getProductByUserSql, userId)
    if err != nil {
        log.Errorf("❌ Error while querying products by user: %v", err)
        return []domain.Product{}
    }
    defer productRows.Close()

    var products []domain.Product
    for productRows.Next() {
        var p domain.Product
        err := productRows.Scan(&p.Id, &p.Name, &p.Price, &p.Description, 
                               &p.Discount, &p.Store, &p.CategoryID, &p.UserID)
        if err != nil {
            continue
        }
        products = append(products, p)
    }

    return products
}
```

#### 🧪 Testing User Ownership

**Manual Testing:**

```bash
# 1. İki farklı kullanıcı kaydet
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"user1","email":"user1@test.com","password":"pass123","first_name":"User","last_name":"One"}'

curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"user2","email":"user2@test.com","password":"pass123","first_name":"User","last_name":"Two"}'

# 2. Her kullanıcı için token al
TOKEN1=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username_or_email":"user1","password":"pass123"}' | jq -r '.token')

TOKEN2=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username_or_email":"user2","password":"pass123"}' | jq -r '.token')

# 3. User1 product oluştursun
curl -X POST http://localhost:8080/api/v1/products \
  -H "Authorization: Bearer $TOKEN1" \
  -H "Content-Type: application/json" \
  -d '{"name":"User1 Product","price":100,"store":"Store1","category_id":1}'

# 4. User2 product oluştursun  
curl -X POST http://localhost:8080/api/v1/products \
  -H "Authorization: Bearer $TOKEN2" \
  -H "Content-Type: application/json" \
  -d '{"name":"User2 Product","price":200,"store":"Store2","category_id":1}'

# 5. Her kullanıcı sadece kendi products'larını görsün
curl -X GET http://localhost:8080/api/v1/products/my-products \
  -H "Authorization: Bearer $TOKEN1"  # Sadece User1'in products'ları

curl -X GET http://localhost:8080/api/v1/products/my-products \
  -H "Authorization: Bearer $TOKEN2"  # Sadece User2'nin products'ları
```

#### 🔧 Advanced Ownership Features

**1. Ownership-based Permissions:**

```go
// Sadece kendi product'ını silebilme
func (productController *ProductController) DeleteProductById(c echo.Context) error {
    productId, _ := strconv.Atoi(c.Param("id"))
    currentUserId := c.Get("user_id").(int64)
    
    // Product sahibini kontrol et
    product, err := productController.productService.GetById(int64(productId))
    if err != nil {
        return c.JSON(http.StatusNotFound, response.ErrorResponse{
            ErrorDescription: "Product not found",
        })
    }
    
    // 🚫 Ownership kontrolü
    if product.UserID != currentUserId {
        return c.JSON(http.StatusForbidden, response.ErrorResponse{
            ErrorDescription: "You can only delete your own products",
        })
    }
    
    // Silme işlemine devam et
    err = productController.productService.DeleteById(int64(productId))
    return c.NoContent(http.StatusOK)
}
```

**2. Ownership Statistics:**

```go
// Kullanıcı istatistikleri
func (userService *UserService) GetUserStats(userId int64) UserStats {
    return UserStats{
        TotalProducts:    productRepository.CountByUser(userId),
        TotalRevenue:     productRepository.SumRevenueByUser(userId),
        AveragePrice:     productRepository.AvgPriceByUser(userId),
        MostExpensive:    productRepository.GetMostExpensiveByUser(userId),
        RecentProducts:   productRepository.GetRecentByUser(userId, 5),
    }
}
```

Bu sistem sayesinde:
- ✅ **Güvenlik**: Kullanıcılar sadece kendi verilerine erişebilir
- ✅ **Otomatik**: UserID manuel girilmez, JWT'den otomatik alınır  
- ✅ **Takip**: Hangi kullanıcının hangi veriyi oluşturduğu bilinir
- ✅ **Izolasyon**: Kullanıcı verileri birbirinden izole edilir
- ✅ **Audit**: Tüm işlemler kullanıcı bazında loglanabilir

## 🔒 Protected Endpoints

### 1. Route Registration

```go
func (userController *UserController) RegisterRoutes(e *echo.Echo) {
    // Public routes - kimlik doğrulama yok
    e.POST("/api/v1/auth/register", userController.Register)
    e.POST("/api/v1/auth/login", userController.Login)
    
    // Protected routes - JWT middleware ile korunmuş
    protected := e.Group("/api/v1/users", middleware.JWTMiddleware())
    protected.GET("/:id", userController.GetUserById)
    protected.PUT("/:id", userController.UpdateUser)
    protected.DELETE("/:id", userController.DeleteUser)
}
```

### 2. User Context Usage

Protected endpoint'lerde kullanıcı bilgilerine erişim:

```go
func (controller *Controller) ProtectedEndpoint(c echo.Context) error {
    // JWT middleware'den kullanıcı bilgilerini al
    userIdInterface := c.Get("user_id")
    userId, ok := userIdInterface.(int64)
    if !ok {
        return c.JSON(http.StatusUnauthorized, map[string]string{
            "error": "Invalid user authentication",
        })
    }
    
    // Artık userId'yi kullanabilirsin
    // Örnek: sadece kendi ürünlerini getir
    products := service.GetProductsByUser(userId)
    return c.JSON(http.StatusOK, products)
}
```

## 💻 Kod Örnekleri

### 1. Registration Example

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john@example.com",
    "password": "mypassword123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

**Response (201 Created):**
```json
{
  "message": "User registered successfully"
}
```

### 2. Login Example

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username_or_email": "johndoe",
    "password": "mypassword123"
  }'
```

**Response (200 OK):**
```json
{
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6ImpvaG5kb2UiLCJlbWFpbCI6ImpvaG5AZXhhbXBsZS5jb20iLCJleHAiOjE3MDY3ODQwMDB9.signature",
  "user": {
    "id": 1,
    "username": "johndoe",
    "email": "john@example.com",
    "first_name": "John",
    "last_name": "Doe"
  }
}
```

### 3. Protected Endpoint Usage

```bash
curl -X GET http://localhost:8080/api/v1/products/my-products \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### 4. Frontend Integration (JavaScript)

```javascript
// Login function
async function login(usernameOrEmail, password) {
    const response = await fetch('/api/v1/auth/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            username_or_email: usernameOrEmail,
            password: password
        })
    });
    
    const data = await response.json();
    
    if (response.ok) {
        // Token'ı localStorage'a kaydet
        localStorage.setItem('jwt_token', data.token);
        localStorage.setItem('user', JSON.stringify(data.user));
        return data;
    } else {
        throw new Error(data.error);
    }
}

// Protected API call
async function getMyProducts() {
    const token = localStorage.getItem('jwt_token');
    
    const response = await fetch('/api/v1/products/my-products', {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    
    if (response.ok) {
        return await response.json();
    } else if (response.status === 401) {
        // Token geçersiz, yeniden login yönlendir
        window.location.href = '/login';
    }
}
```

## 🔐 Güvenlik Özellikleri

### 1. Password Security
- **Argon2 Hashing**: Endüstri standardı password hashing
- **Random Salt**: Her password için benzersiz salt
- **Constant-time Comparison**: Timing attack'lara karşı koruma

### 2. JWT Security
- **HS256 Signature**: HMAC-SHA256 ile imzalama
- **Short Expiry**: 24 saatlik token geçerlilik süresi
- **Secret Key**: Ortam değişkeni ile gizli anahtar

### 3. Input Validation
- **SQL Injection**: Parametrized queries kullanımı
- **XSS Protection**: Input sanitization
- **Rate Limiting**: Brute force attack koruması (önerilir)

### 4. Best Practices
```go
// ✅ Güvenli - Environment variable kullan
jwtSecret := os.Getenv("JWT_SECRET")

// ❌ Güvensiz - Hard-coded secret
jwtSecret := "my-secret-key"

// ✅ Güvenli - Constant time comparison
subtle.ConstantTimeCompare(hash1, hash2)

// ❌ Güvensiz - Direct comparison
hash1 == hash2
```

## ❌ Hata Yönetimi

### 1. Authentication Errors

| HTTP Status | Error Code | Açıklama |
|------------|------------|-----------|
| 400 | Bad Request | Geçersiz request formatı |
| 401 | Unauthorized | Geçersiz kimlik bilgileri |
| 409 | Conflict | Kullanıcı adı/email zaten mevcut |
| 422 | Unprocessable Entity | Validation hataları |
| 500 | Internal Server Error | Sunucu hatası |

### 2. Error Response Format

```json
{
  "error": "Invalid credentials"
}
```

### 3. Common Error Messages

```go
// Registration errors
"username already exists"
"email already exists"
"username must be at least 3 characters long"
"invalid email format"
"password must be at least 6 characters long"

// Login errors
"invalid credentials"
"username/email and password are required"

// JWT errors
"Missing authorization header"
"Invalid authorization header format"
"Invalid or expired token"
"Invalid user authentication"
```

## 🧪 Test Etme

### 1. Manual Testing

```bash
# 1. Kullanıcı kaydı
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"test123","first_name":"Test","last_name":"User"}'

# 2. Giriş yap ve token al
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username_or_email":"testuser","password":"test123"}' \
  | jq -r '.token')

# 3. Protected endpoint'i test et
curl -X GET http://localhost:8080/api/v1/users/1 \
  -H "Authorization: Bearer $TOKEN"
```

### 2. Unit Testing

```go
func TestUserRegistration(t *testing.T) {
    // Test cases:
    // - Valid registration
    // - Duplicate username
    // - Duplicate email
    // - Invalid email format
    // - Short password
    // - Missing fields
}

func TestUserLogin(t *testing.T) {
    // Test cases:
    // - Valid credentials
    // - Invalid username
    // - Invalid password
    // - Missing fields
    // - Login with email
}

func TestJWTMiddleware(t *testing.T) {
    // Test cases:
    // - Valid token
    // - Expired token
    // - Invalid signature
    // - Missing authorization header
    // - Malformed token
}
```

## 🚀 Production Checklist

### Environment Variables
```bash
# .env file
JWT_SECRET=your-super-secret-jwt-key-min-32-chars
DATABASE_URL=postgresql://user:pass@localhost/db
```

### Security Headers
```go
// CORS, Security headers
e.Use(middleware.CORS())
e.Use(middleware.Secure())
```

### Rate Limiting
```go
// Rate limiting for auth endpoints
e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))
```

---

## 📚 Ek Kaynaklar

- [JWT.io](https://jwt.io/) - JWT token debugger
- [Argon2 Specification](https://github.com/P-H-C/phc-winner-argon2)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [Go JWT Library](https://github.com/golang-jwt/jwt)

---

**💡 İpucu**: Bu dökümanı referans olarak kullan ve authentication sistemini anlamak için kod örneklerini dene!