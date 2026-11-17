package controllers

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"school-management/config"
	"school-management/models"
)

type teacherImportRow struct {
	Code, Name, Email, Password, NIP, Gender, Subject, Position, Phone, Address string
	PlusHours, RegularHours, TotalHours                                         int
}

func ImportTeachersCSV(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pilih file CSV untuk diimpor"})
		return
	}
	defer file.Close()
	reader := csv.NewReader(io.LimitReader(file, 5<<20))
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV tidak dapat dibaca: " + err.Error()})
		return
	}
	if len(records) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV harus berisi header dan minimal satu data guru"})
		return
	}
	if len(records) > 501 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maksimal 500 guru per impor"})
		return
	}

	columns := map[string]int{}
	for i, name := range records[0] {
		columns[strings.ToLower(strings.TrimSpace(strings.TrimPrefix(name, "\ufeff")))] = i
	}
	required := []string{"kode_guru", "nama", "email", "password", "bidang_tugas"}
	for _, col := range required {
		if _, ok := columns[col]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Kolom wajib '%s' tidak ditemukan", col)})
			return
		}
	}
	value := func(row []string, key string) string {
		i, ok := columns[key]
		if !ok || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}
	integer := func(row []string, key string) (int, error) {
		s := value(row, key)
		if s == "" {
			return 0, nil
		}
		n, e := strconv.Atoi(s)
		if e != nil || n < 0 {
			return 0, fmt.Errorf("kolom %s harus berupa angka 0 atau lebih", key)
		}
		return n, nil
	}
	rows := make([]teacherImportRow, 0, len(records)-1)
	codes, emails, nips := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for index, raw := range records[1:] {
		line := index + 2
		if len(raw) == 1 && strings.TrimSpace(raw[0]) == "" {
			continue
		}
		item := teacherImportRow{Code: value(raw, "kode_guru"), Name: value(raw, "nama"), Email: value(raw, "email"), Password: value(raw, "password"), NIP: value(raw, "nip"), Gender: value(raw, "jenis_kelamin"), Subject: value(raw, "bidang_tugas"), Position: value(raw, "jabatan"), Phone: value(raw, "no_hp"), Address: value(raw, "alamat")}
		if item.Code == "" || item.Name == "" || item.Email == "" || item.Password == "" || item.Subject == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Baris %d: kode, nama, email, password, dan bidang tugas wajib diisi", line)})
			return
		}
		if !strings.Contains(item.Email, "@") || len(item.Password) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Baris %d: email tidak valid atau password kurang dari 6 karakter", line)})
			return
		}
		codeKey, emailKey := strings.ToLower(item.Code), strings.ToLower(item.Email)
		if codes[codeKey] {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Baris %d: kode guru duplikat (%s)", line, item.Code)})
			return
		}
		codes[codeKey] = true
		if emails[emailKey] {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Baris %d: email duplikat (%s)", line, item.Email)})
			return
		}
		emails[emailKey] = true
		if item.NIP != "" {
			nipKey := strings.ToLower(item.NIP)
			if nips[nipKey] {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Baris %d: NIP duplikat (%s)", line, item.NIP)})
				return
			}
			nips[nipKey] = true
		}
		var parseErr error
		if item.PlusHours, parseErr = integer(raw, "jam_plus"); parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Baris %d: %v", line, parseErr)})
			return
		}
		if item.RegularHours, parseErr = integer(raw, "jam_reguler"); parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Baris %d: %v", line, parseErr)})
			return
		}
		if item.TotalHours, parseErr = integer(raw, "total_jam"); parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Baris %d: %v", line, parseErr)})
			return
		}
		rows = append(rows, item)
	}
	if len(rows) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak ada baris data guru yang dapat diimpor"})
		return
	}
	for _, item := range rows {
		if emailTakenByOther(item.Email, 0) {
			c.JSON(http.StatusConflict, gin.H{"error": "Email sudah digunakan: " + item.Email})
			return
		}
		var count int64
		config.DB.Model(&models.Teacher{}).Where("code = ?", item.Code).Count(&count)
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Kode guru sudah terdaftar: " + item.Code})
			return
		}
		if item.NIP != "" && nipTakenByOther(item.NIP, 0) {
			c.JSON(http.StatusConflict, gin.H{"error": "NIP sudah terdaftar: " + item.NIP})
			return
		}
	}
	tx := config.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tidak dapat memulai impor"})
		return
	}
	for _, item := range rows {
		hash, hashErr := bcrypt.GenerateFromPassword([]byte(item.Password), bcrypt.DefaultCost)
		if hashErr != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengenkripsi password akun guru"})
			return
		}
		user := models.User{Name: item.Name, Email: item.Email, Password: string(hash), Role: "TEACHER"}
		if err := tx.Create(&user).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusConflict, gin.H{"error": "Impor dibatalkan karena akun gagal dibuat: " + err.Error()})
			return
		}
		teacher := models.Teacher{UserID: user.ID, NIP: optionalString(item.NIP), Gender: item.Gender, Subject: item.Subject, Code: item.Code, Position: item.Position, Phone: item.Phone, Address: item.Address, PlusHours: item.PlusHours, RegularHours: item.RegularHours, TotalHours: item.TotalHours}
		if err := tx.Create(&teacher).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusConflict, gin.H{"error": "Impor dibatalkan karena data guru gagal dibuat: " + err.Error()})
			return
		}
	}
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impor gagal disimpan: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": fmt.Sprintf("%d akun guru berhasil diimpor", len(rows)), "imported": len(rows)})
}
