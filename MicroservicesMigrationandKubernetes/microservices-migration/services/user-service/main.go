package main

import (
    "log"
    "net/http"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
    "gorm.io/gorm"
    "gorm.io/driver/postgres"
)

// User represents the user model
type User struct {
    gorm.Model
    Username string `json:"username"`
    Email    string `json:"email"`
}

var (
    db          *gorm.DB
    redisClient *redis.Client
)

func main() {
    // Initialize database connection
    initDB()

    // Initialize Redis client
    initRedis()

    // Set up Gin router
    r := gin.Default()

    // Define routes
    r.GET("/users/:id", getUserHandler)
    r.POST("/users", createUserHandler)

    // Start the server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    log.Printf("Starting server on port %s", port)
    if err := r.Run(":" + port); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}

func initDB() {
    var err error
    dsn := os.Getenv("DATABASE_URL")
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    // Auto-migrate the User model
    db.AutoMigrate(&User{})
}

func initRedis() {
    redisAddr := os.Getenv("REDIS_ADDR")
    if redisAddr == "" {
        redisAddr = "localhost:6379"
    }
    redisClient = redis.NewClient(&redis.Options{
        Addr: redisAddr,
    })
}

func getUserHandler(c *gin.Context) {
    id := c.Param("id")
    var user User
    if err := db.First(&user, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }
    c.JSON(http.StatusOK, user)
}

func createUserHandler(c *gin.Context) {
    var newUser User
    if err := c.BindJSON(&newUser); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    if err := db.Create(&newUser).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
        return
    }
    c.JSON(http.StatusCreated, newUser)
}