package config

import "os"

// JWTSecret dibaca dari environment variable JWT_SECRET.
// Jika belum diset, memakai nilai default (ganti sebelum production).
var JWTSecret = func() string {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s
	}
	return "school-management-secret-key-change-later"
}()