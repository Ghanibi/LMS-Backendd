package middleware

import (
	"fmt"
	"net/http"
	"school-management/config"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Request tidak menyertakan token"})
			c.Abort()
			return
		}

		// Format header harus "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Header otorisasi tidak valid"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validasi token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Pastikan metode signing adalah HS256
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Metode signing tidak terduga: %v", token.Header["alg"])
			}
			return []byte(config.JWTSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid: " + err.Error()})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Jika token valid, simpan informasi user ke context
			c.Set("user_id", claims["user_id"])
			c.Set("name", claims["name"])
			c.Set("email", claims["email"])
			c.Set("role", claims["role"])
			c.Next() // Lanjutkan ke handler berikutnya (controllers.GetMe)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid"})
			c.Abort()
			return
		}
	}
}

// RoleMiddleware adalah middleware untuk memeriksa peran (role) pengguna.
// Middleware ini harus dijalankan SETELAH AuthMiddleware.
func RoleMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ambil role dari context yang sudah di-set oleh AuthMiddleware
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Role pengguna tidak ditemukan di dalam token"})
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses role pengguna"})
			c.Abort()
			return
		}

		// Periksa apakah role pengguna ada di dalam daftar role yang diizinkan
		for _, requiredRole := range requiredRoles {
			if userRole == requiredRole {
				c.Next() // Role cocok, lanjutkan ke handler berikutnya
				return
			}
			// SUPER_ADMIN mewarisi akses ADMIN ke modul sekolah, tetapi tidak otomatis
			// mewarisi akses role lain.
			if userRole == "SUPER_ADMIN" && requiredRole == "ADMIN" {
				c.Next()
				return
			}
		}

		// Jika loop selesai dan tidak ada role yang cocok
		c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki hak akses untuk sumber daya ini"})
		c.Abort()
	}
}
