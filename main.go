package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	webview "github.com/webview/webview_go"
	"github.com/xuri/excelize/v2"
)

// FileInfo represents information about uploaded files
type FileInfo struct {
	Name string `json:"name"`
	Size string `json:"size"`
	Type string `json:"type"`
	Path string `json:"path"`
}

type Product struct {
	ID            int            `json:"id"`
	PartNo        string         `json:"partNo"`
	PartName      string         `json:"partName"`
	Description   string         `json:"description"`
	Cost          string         `json:"cost"`
	Qty           int            `json:"qty"`
	Material      string         `json:"material"`
	MaterialSize  string         `json:"materialSize"`
	MaterialCost  string         `json:"materialCost"`
	FinishingType string         `json:"finishingType"`
	FinishingCost string         `json:"finishingCost"`
	Photos        sql.NullString `json:"-"`
	Drawing2D     sql.NullString `json:"-"`
	Cad3D         sql.NullString `json:"-"`
	CncCode       sql.NullString `json:"-"`
	Invoice       sql.NullString `json:"-"`
	CreatedAt     string         `json:"createdAt"`
	UpdatedAt     string         `json:"updatedAt"`
	PhotoFiles    []FileInfo     `json:"photos,omitempty"`
}

type TemplateData struct {
	Products    []Product
	SearchQuery string
	SortBy      string
	SortOrder   string
}

type PaginatedResponse struct {
	Products    []Product `json:"products"`
	HasMore     bool      `json:"hasMore"`
	TotalCount  int       `json:"totalCount"`
	CurrentPage int       `json:"currentPage"`
}

var db *sql.DB
var uploadDir string

// Template functions
var funcMap = template.FuncMap{
	"parseJSON": func(s string) []FileInfo {
		var result []FileInfo
		if s == "" {
			return result
		}
		json.Unmarshal([]byte(s), &result)
		return result
	},
	"parseNullJSON": func(ns sql.NullString) []FileInfo {
		var result []FileInfo
		if !ns.Valid || ns.String == "" {
			return result
		}
		json.Unmarshal([]byte(ns.String), &result)
		return result
	},
	"formatDate": func(date string) string {
		t, err := time.Parse("2006-01-02T15:04:05Z", date)
		if err != nil {
			return date
		}
		return t.Format("2006-01-02 15:04")
	},
	"subtract": func(a, b int) int {
		return a - b
	},
	"hasFiles": hasFiles,
	"hasNullFiles": func(ns sql.NullString) bool {
		if !ns.Valid || ns.String == "" || ns.String == "null" {
			return false
		}
		var arr []map[string]interface{}
		if err := json.Unmarshal([]byte(ns.String), &arr); err != nil {
			return false
		}
		return len(arr) > 0
	},
	"nullString": func(ns sql.NullString) string {
		if ns.Valid {
			return ns.String
		}
		return ""
	},
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./products.db")
	if err != nil {
		log.Fatal(err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		partNo TEXT UNIQUE,
		partName TEXT,
		description TEXT,
		cost TEXT,
		qty INTEGER DEFAULT 0,
		material TEXT,
		material_size TEXT,
		material_cost TEXT,
		finishing_type TEXT,
		finishing_cost TEXT,
		photos TEXT,
		drawing_2d TEXT,
		cad_3d TEXT,
		cnc_code TEXT,
		invoice TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	triggerSQL := `
	CREATE TRIGGER IF NOT EXISTS update_products_timestamp 
	AFTER UPDATE ON products
	BEGIN
		UPDATE products SET updated_at = DATETIME('now') 
		WHERE id = NEW.id;
	END;
	`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(triggerSQL)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Database initialized successfully")
}

func init() {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	uploadDir = filepath.Join(currentDir, "uploads")
	err = os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		log.Fatal("Error creating upload directory:", err)
	}

	fmt.Println("Upload directory set to:", uploadDir)
}

func main() {
	initDB()
	defer db.Close()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir))))

	// Routes
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/save", saveHandler)
	http.HandleFunc("/modify/", modifyHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/remove-file", removeFileHandler)
	http.HandleFunc("/delete/", deleteHandler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/detail/", detailHandler)
	http.HandleFunc("/export", exportHandler)
	http.HandleFunc("/open-folder", openFolderHandler)
	http.HandleFunc("/api/products", apiProductsHandler)
	http.HandleFunc("/open-file", openFileHandler)

	go func() {
		log.Println("Server starting on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal(err)
		}
	}()

	time.Sleep(500 * time.Millisecond)

	debug := true
	w := webview.New(debug)
	defer w.Destroy()
	w.SetTitle("Product Manager")
	w.SetSize(1280, 800, webview.HintNone)
	w.Navigate("http://localhost:8080")
	w.Run()
}

func apiProductsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	sortBy := r.URL.Query().Get("sort")
	sortOrder := r.URL.Query().Get("order")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 5000
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := (page - 1) * limit

	if sortBy == "" {
		sortBy = "updated_at"
		sortOrder = "DESC"
	}

	validSortColumns := map[string]bool{
		"id":             true,
		"partNo":         true,
		"partName":       true,
		"description":    true,
		"cost":           true,
		"qty":            true,
		"material":       true,
		"material_size":  true,
		"material_cost":  true,
		"finishing_type": true,
		"finishing_cost": true,
		"created_at":     true,
		"updated_at":     true,
	}
	validSortOrders := map[string]bool{"ASC": true, "DESC": true}

	if !validSortColumns[sortBy] {
		sortBy = "updated_at"
	}
	if !validSortOrders[sortOrder] {
		sortOrder = "DESC"
	}

	baseQuery := `
		SELECT id, partNo, partName, description, cost, qty, material,
			   material_size, material_cost, finishing_type, finishing_cost,
			   photos, drawing_2d, cad_3d, cnc_code, invoice, 
			   created_at, updated_at 
		FROM products 
	`
	countQuery := "SELECT COUNT(*) FROM products "
	whereClause := ""
	args := []interface{}{}

	if query != "" {
		whereClause = "WHERE partNo LIKE ? OR partName LIKE ? OR description LIKE ? OR material LIKE ?"
		args = []interface{}{"%" + query + "%", "%" + query + "%", "%" + query + "%", "%" + query + "%"}
	}

	var totalCount int
	countQuery += whereClause
	err := db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		http.Error(w, "Error counting products: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fullQuery := baseQuery + whereClause + " ORDER BY " + sortBy + " " + sortOrder + " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.Query(fullQuery, args...)
	if err != nil {
		http.Error(w, "Error querying products: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		var photosJSON sql.NullString

		err := rows.Scan(
			&p.ID,
			&p.PartNo,
			&p.PartName,
			&p.Description,
			&p.Cost,
			&p.Qty,
			&p.Material,
			&p.MaterialSize,
			&p.MaterialCost,
			&p.FinishingType,
			&p.FinishingCost,
			&photosJSON,
			&p.Drawing2D,
			&p.Cad3D,
			&p.CncCode,
			&p.Invoice,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			http.Error(w, "Error scanning product: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if photosJSON.Valid && photosJSON.String != "" {
			var photoFiles []FileInfo
			if err := json.Unmarshal([]byte(photosJSON.String), &photoFiles); err == nil {
				p.PhotoFiles = photoFiles
			}
		}

		products = append(products, p)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, "Error iterating products: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := PaginatedResponse{
		Products:    products,
		HasMore:     (page * limit) < totalCount,
		TotalCount:  totalCount,
		CurrentPage: page,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		apiProductsHandler(w, r)
		return
	}

	query := r.URL.Query().Get("q")
	sortBy := r.URL.Query().Get("sort")
	sortOrder := r.URL.Query().Get("order")

	if sortBy == "" {
		sortBy = "updated_at"
		sortOrder = "DESC"
	}

	validSortColumns := map[string]bool{
		"partNo":         true,
		"partName":       true,
		"description":    true,
		"cost":           true,
		"qty":            true,
		"material":       true,
		"material_size":  true,
		"material_cost":  true,
		"finishing_type": true,
		"finishing_cost": true,
		"updated_at":     true,
	}
	validSortOrders := map[string]bool{"ASC": true, "DESC": true}

	if !validSortColumns[sortBy] {
		sortBy = "updated_at"
	}
	if !validSortOrders[sortOrder] {
		sortOrder = "DESC"
	}

	limit := 500
	querySQL := `
		SELECT id, partNo, partName, description, cost, qty, material,
			   material_size, material_cost, finishing_type, finishing_cost,
			   photos, drawing_2d, cad_3d, cnc_code, invoice,
			   created_at, updated_at
		FROM products
	`

	args := []interface{}{}
	if query != "" {
		querySQL += " WHERE partNo LIKE ? OR partName LIKE ? OR description LIKE ? OR material LIKE ?"
		args = append(args, "%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%")
	}

	querySQL += " ORDER BY " + sortBy + " " + sortOrder + " LIMIT ?"
	args = append(args, limit)

	rows, err := db.Query(querySQL, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err := rows.Scan(
			&p.ID,
			&p.PartNo,
			&p.PartName,
			&p.Description,
			&p.Cost,
			&p.Qty,
			&p.Material,
			&p.MaterialSize,
			&p.MaterialCost,
			&p.FinishingType,
			&p.FinishingCost,
			&p.Photos,
			&p.Drawing2D,
			&p.Cad3D,
			&p.CncCode,
			&p.Invoice,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		products = append(products, p)
	}

	data := TemplateData{
		Products:    products,
		SearchQuery: query,
		SortBy:      sortBy,
		SortOrder:   sortOrder,
	}

	tmpl := template.Must(template.New("index.html").Funcs(funcMap).ParseFiles("templates/index.html"))
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		apiProductsHandler(w, r)
		return
	}

	query := r.URL.Query().Get("q")
	sortBy := r.URL.Query().Get("sort")
	sortOrder := r.URL.Query().Get("order")

	if sortBy == "" {
		sortBy = "updated_at"
		sortOrder = "DESC"
	}

	validSortColumns := map[string]bool{
		"id":             true,
		"partNo":         true,
		"partName":       true,
		"description":    true,
		"cost":           true,
		"qty":            true,
		"material":       true,
		"material_size":  true,
		"material_cost":  true,
		"finishing_type": true,
		"finishing_cost": true,
		"created_at":     true,
		"updated_at":     true,
	}
	validSortOrders := map[string]bool{"ASC": true, "DESC": true}

	if !validSortColumns[sortBy] {
		sortBy = "updated_at"
	}
	if !validSortOrders[sortOrder] {
		sortOrder = "DESC"
	}

	limit := 5000
	querySQL := `
		SELECT id, partNo, partName, description, cost, qty, material,
			   material_size, material_cost, finishing_type, finishing_cost,
			   photos, drawing_2d, cad_3d, cnc_code, invoice, 
			   created_at, updated_at 
		FROM products 
	`
	args := []interface{}{}

	if query != "" {
		querySQL += "WHERE partNo LIKE ? OR partName LIKE ? OR description LIKE ? OR material LIKE ? "
		args = append(args, "%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%")
	}

	querySQL += "ORDER BY " + sortBy + " " + sortOrder + " LIMIT ?"
	args = append(args, limit)

	rows, err := db.Query(querySQL, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err := rows.Scan(
			&p.ID,
			&p.PartNo,
			&p.PartName,
			&p.Description,
			&p.Cost,
			&p.Qty,
			&p.Material,
			&p.MaterialSize,
			&p.MaterialCost,
			&p.FinishingType,
			&p.FinishingCost,
			&p.Photos,
			&p.Drawing2D,
			&p.Cad3D,
			&p.CncCode,
			&p.Invoice,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		products = append(products, p)
	}

	tmpl := template.Must(template.New("index.html").Funcs(funcMap).ParseFiles("templates/index.html"))
	data := TemplateData{
		Products:    products,
		SearchQuery: query,
		SortBy:      sortBy,
		SortOrder:   sortOrder,
	}
	tmpl.Execute(w, data)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("add.html").Funcs(funcMap).ParseFiles("templates/add.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func detailHandler(w http.ResponseWriter, r *http.Request) {
	partNo := strings.TrimPrefix(r.URL.Path, "/detail/")

	row := db.QueryRow(`
		SELECT id, partNo, partName, description, cost, qty, material,
			   material_size, material_cost, finishing_type, finishing_cost,
			   photos, drawing_2d, cad_3d, cnc_code, invoice,
			   created_at, updated_at
		FROM products WHERE partNo = ?`, partNo)

	var p Product
	err := row.Scan(
		&p.ID,
		&p.PartNo,
		&p.PartName,
		&p.Description,
		&p.Cost,
		&p.Qty,
		&p.Material,
		&p.MaterialSize,
		&p.MaterialCost,
		&p.FinishingType,
		&p.FinishingCost,
		&p.Photos,
		&p.Drawing2D,
		&p.Cad3D,
		&p.CncCode,
		&p.Invoice,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.New("detail.html").Funcs(funcMap).ParseFiles("templates/detail.html"))
	err = tmpl.Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	partNo := r.FormValue("partNo")
	partName := r.FormValue("partName")
	description := r.FormValue("description")
	cost := r.FormValue("cost")
	qty, _ := strconv.Atoi(r.FormValue("qty"))
	material := r.FormValue("material")
	materialSize := r.FormValue("materialSize")
	materialCost := r.FormValue("materialCost")
	finishingType := r.FormValue("finishingType")
	finishingCost := r.FormValue("finishingCost")

	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM products WHERE partNo = ?)", partNo).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if exists {
		data := struct {
			Error         string
			PartNo        string
			PartName      string
			Description   string
			Cost          string
			Qty           int
			Material      string
			MaterialSize  string
			MaterialCost  string
			FinishingType string
			FinishingCost string
		}{
			Error:         "PartNo number already exists",
			PartNo:        partNo,
			PartName:      partName,
			Description:   description,
			Cost:          cost,
			Qty:           qty,
			Material:      material,
			MaterialSize:  materialSize,
			MaterialCost:  materialCost,
			FinishingType: finishingType,
			FinishingCost: finishingCost,
		}

		tmpl := template.Must(template.ParseFiles("templates/add.html"))
		w.WriteHeader(http.StatusBadRequest)
		tmpl.Execute(w, data)
		return
	}

	partNoDir := filepath.Join(uploadDir, sanitizeFilename(partNo))
	err = os.MkdirAll(partNoDir, os.ModePerm)
	if err != nil {
		http.Error(w, "Error creating partNo directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	subDirs := []string{"photos", "drawings", "cad", "cnc", "invoice"}
	for _, dir := range subDirs {
		dirPath := filepath.Join(partNoDir, dir)
		err = os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			http.Error(w, "Error creating "+dir+" directory: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	photoInfo := handleFileUpload(r, "photos", "photos")
	drawingInfo := handleFileUpload(r, "drawings", "drawings")
	cadInfo := handleFileUpload(r, "cad", "cad")
	cncInfo := handleFileUpload(r, "cnc", "cnc")
	invoiceInfo := handleFileUpload(r, "invoice", "invoice")

	photosJSON, _ := json.Marshal(photoInfo)
	drawingsJSON, _ := json.Marshal(drawingInfo)
	cadJSON, _ := json.Marshal(cadInfo)
	cncJSON, _ := json.Marshal(cncInfo)
	invoiceJSON, _ := json.Marshal(invoiceInfo)

	stmt, err := db.Prepare(`
		INSERT INTO products(
			partNo, partName, description, cost, qty, material,
			material_size, material_cost, finishing_type, finishing_cost,
			photos, drawing_2d, cad_3d, cnc_code, invoice
		) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		partNo, partName, description, cost, qty, material,
		materialSize, materialCost, finishingType, finishingCost,
		string(photosJSON), string(drawingsJSON),
		string(cadJSON), string(cncJSON), string(invoiceJSON),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func modifyHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/modify/")
	row := db.QueryRow(`
		SELECT id, partNo, partName, description, cost, qty, material,
			   material_size, material_cost, finishing_type, finishing_cost,
			   photos, drawing_2d, cad_3d, cnc_code, invoice,
			   created_at, updated_at
		FROM products WHERE id = ?`, id)

	var p Product
	err := row.Scan(
		&p.ID, &p.PartNo, &p.PartName, &p.Description, &p.Cost, &p.Qty,
		&p.Material, &p.MaterialSize, &p.MaterialCost, &p.FinishingType, &p.FinishingCost,
		&p.Photos, &p.Drawing2D, &p.Cad3D, &p.CncCode, &p.Invoice,
		&p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.New("modify.html").Funcs(funcMap).ParseFiles("templates/modify.html"))
	err = tmpl.Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	id := r.FormValue("id")
	newPartNo := r.FormValue("partNo")
	partName := r.FormValue("partName")
	description := r.FormValue("description")
	cost := r.FormValue("cost")
	qty, _ := strconv.Atoi(r.FormValue("qty"))
	material := r.FormValue("material")
	materialSize := r.FormValue("materialSize")
	materialCost := r.FormValue("materialCost")
	finishingType := r.FormValue("finishingType")
	finishingCost := r.FormValue("finishingCost")

	var existingPhotos, existingDrawings, existingCad, existingCnc, existingInvoice string
	err := db.QueryRow(`
		SELECT photos, drawing_2d, cad_3d, cnc_code, invoice 
		FROM products WHERE id = ?`, id).Scan(
		&existingPhotos, &existingDrawings, &existingCad, &existingCnc, &existingInvoice)
	if err != nil {
		http.Error(w, "Error getting existing files: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var oldPartNo string
	err = db.QueryRow("SELECT partNo FROM products WHERE id = ?", id).Scan(&oldPartNo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	photoInfo := handleFileUploadWithPartNo(r, "photos", "photos", newPartNo)
	drawingInfo := handleFileUploadWithPartNo(r, "drawings", "drawings", newPartNo)
	cadInfo := handleFileUploadWithPartNo(r, "cad", "cad", newPartNo)
	cncInfo := handleFileUploadWithPartNo(r, "cnc", "cnc", newPartNo)
	invoiceInfo := handleFileUploadWithPartNo(r, "invoice", "invoice", newPartNo)

	photos := mergeFileData(existingPhotos, photoInfo)
	drawings := mergeFileData(existingDrawings, drawingInfo)
	cad := mergeFileData(existingCad, cadInfo)
	cnc := mergeFileData(existingCnc, cncInfo)
	invoice := mergeFileData(existingInvoice, invoiceInfo)

	if oldPartNo != newPartNo {
		oldPath := filepath.Join(uploadDir, sanitizeFilename(oldPartNo))
		newPath := filepath.Join(uploadDir, sanitizeFilename(newPartNo))

		if _, err := os.Stat(oldPath); err == nil {
			if err := os.MkdirAll(filepath.Dir(newPath), os.ModePerm); err != nil {
				http.Error(w, "Error creating new directory: "+err.Error(), http.StatusInternalServerError)
				return
			}

			if err := os.Rename(oldPath, newPath); err != nil {
				http.Error(w, "Error moving directory: "+err.Error(), http.StatusInternalServerError)
				return
			}

			updateFilePaths := func(fileJSON, oldPartNo, newPartNo string) string {
				if fileJSON == "" {
					return ""
				}
				var files []FileInfo
				if err := json.Unmarshal([]byte(fileJSON), &files); err != nil {
					return fileJSON
				}
				for i := range files {
					pathParts := strings.Split(files[i].Path, "/")
					if len(pathParts) > 0 {
						pathParts[0] = sanitizeFilename(newPartNo)
						files[i].Path = strings.Join(pathParts, "/")
					}
				}
				newJSON, _ := json.Marshal(files)
				return string(newJSON)
			}

			photos = updateFilePaths(photos, oldPartNo, newPartNo)
			drawings = updateFilePaths(drawings, oldPartNo, newPartNo)
			cad = updateFilePaths(cad, oldPartNo, newPartNo)
			cnc = updateFilePaths(cnc, oldPartNo, newPartNo)
			invoice = updateFilePaths(invoice, oldPartNo, newPartNo)
		} else {
			newPartNoDir := filepath.Join(uploadDir, sanitizeFilename(newPartNo))
			if err := os.MkdirAll(newPartNoDir, os.ModePerm); err != nil {
				http.Error(w, "Error creating new partNo directory: "+err.Error(), http.StatusInternalServerError)
				return
			}

			subDirs := []string{"photos", "drawings", "cad", "cnc", "invoice"}
			for _, dir := range subDirs {
				dirPath := filepath.Join(newPartNoDir, dir)
				if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
					http.Error(w, "Error creating "+dir+" directory: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}
	}

	stmt, err := db.Prepare(`
		UPDATE products 
		SET partNo=?, partName=?, description=?, cost=?, qty=?, material=?,
			material_size=?, material_cost=?, finishing_type=?, finishing_cost=?,
			photos=?, drawing_2d=?, cad_3d=?, cnc_code=?, invoice=?,
			updated_at=CURRENT_TIMESTAMP
		WHERE id=?`)
	if err != nil {
		http.Error(w, "Error preparing statement: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		newPartNo, partName, description, cost, qty, material,
		materialSize, materialCost, finishingType, finishingCost,
		photos, drawings, cad, cnc, invoice,
		id)
	if err != nil {
		http.Error(w, "Error updating product: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ... (keep all the remaining helper functions the same as before)
// [The rest of your helper functions remain unchanged - handleFileUpload, removeFileHandler, deleteHandler, exportHandler, etc.]

// Helper functions (keep your existing implementations)
func handleFileUpload(r *http.Request, fieldName, subDir string) []FileInfo {
	var fileInfo []FileInfo

	if r.MultipartForm == nil {
		return fileInfo
	}

	partNo := sanitizeFilename(r.FormValue("partNo"))
	if partNo == "" {
		return fileInfo
	}

	files := r.MultipartForm.File[fieldName]
	if len(files) == 0 {
		return fileInfo
	}

	partNoPath := filepath.Join(uploadDir, partNo, subDir)
	if err := os.MkdirAll(partNoPath, os.ModePerm); err != nil {
		log.Printf("Error creating directory %s: %v", partNoPath, err)
		return fileInfo
	}

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("Error opening file: %v", err)
			continue
		}
		defer file.Close()

		filename := filepath.Base(fileHeader.Filename)
		// // Keep original filename with spaces - only get base name
		// filename = strings.ReplaceAll(filename, " ", "_")
		filePath := filepath.Join(partNo, subDir, filename)
		fullPath := filepath.Join(uploadDir, filePath)

		dst, err := os.Create(fullPath)
		if err != nil {
			log.Printf("Error creating file: %v", err)
			continue
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			log.Printf("Error copying file content: %v", err)
			continue
		}

		info := FileInfo{
			Name: fileHeader.Filename,
			Size: formatFileSize(fileHeader.Size),
			Type: fileHeader.Header.Get("Content-Type"),
			Path: filepath.ToSlash(filePath),
		}
		fileInfo = append(fileInfo, info)

		log.Printf("Successfully uploaded file: %s to %s", fileHeader.Filename, fullPath)
	}

	return fileInfo
}

func handleFileUploadWithPartNo(r *http.Request, fieldName, subDir, partNo string) []FileInfo {
	var fileInfo []FileInfo

	if r.MultipartForm == nil {
		return fileInfo
	}

	files := r.MultipartForm.File[fieldName]
	if len(files) == 0 {
		return fileInfo
	}

	partNoPath := filepath.Join(uploadDir, sanitizeFilename(partNo), subDir)
	if err := os.MkdirAll(partNoPath, os.ModePerm); err != nil {
		log.Printf("Error creating directory %s: %v", partNoPath, err)
		return fileInfo
	}

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("Error opening file: %v", err)
			continue
		}
		defer file.Close()

		filename := filepath.Base(fileHeader.Filename)
		// Keep original filename with spaces - only get base name
		// filename = strings.ReplaceAll(filename, " ", "_")

		fullPath := filepath.Join(partNoPath, filename)

		if _, err := os.Stat(fullPath); err == nil {
			ext := filepath.Ext(filename)
			name := strings.TrimSuffix(filename, ext)

			counter := 1
			for {
				newName := fmt.Sprintf("%s(%d)%s", name, counter, ext)
				newPath := filepath.Join(partNoPath, newName)
				if _, err := os.Stat(newPath); os.IsNotExist(err) {
					fullPath = newPath
					filename = newName
					break
				}
				counter++
			}
		}

		dst, err := os.Create(fullPath)
		if err != nil {
			log.Printf("Error creating file: %v", err)
			continue
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			log.Printf("Error copying file content: %v", err)
			continue
		}

		relativePath := filepath.Join(sanitizeFilename(partNo), subDir, filename)

		info := FileInfo{
			Name: filename,
			Size: formatFileSize(fileHeader.Size),
			Type: fileHeader.Header.Get("Content-Type"),
			Path: filepath.ToSlash(relativePath),
		}
		fileInfo = append(fileInfo, info)

		log.Printf("Successfully uploaded file: %s to %s", filename, fullPath)
	}

	return fileInfo
}

func mergeFileData(existingJSON string, newFiles []FileInfo) string {
	var existing []FileInfo
	if existingJSON != "" {
		json.Unmarshal([]byte(existingJSON), &existing)
	}

	existing = append(existing, newFiles...)

	merged, err := json.Marshal(existing)
	if err != nil {
		log.Printf("Error marshaling file data: %v", err)
		return existingJSON
	}
	return string(merged)
}

func sanitizeFilename(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")

	var safe []rune
	for _, r := range name {
		if (r >= 'a' && r <= 'z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_' {
			safe = append(safe, r)
		}
	}

	if len(safe) == 0 {
		return "file"
	}

	return string(safe)
}

func formatFileSize(bytes int64) string {
	if bytes == 0 {
		return "0 Bytes"
	}

	const unit = 1024
	sizes := []string{"Bytes", "KB", "MB", "GB"}

	i := 0
	size := float64(bytes)
	for size >= unit && i < len(sizes)-1 {
		size /= unit
		i++
	}

	return fmt.Sprintf("%.2f %s", size, sizes[i])
}

func hasFiles(s string) bool {
	if s == "" || s == "null" {
		return false
	}
	var arr []map[string]interface{}
	if err := json.Unmarshal([]byte(s), &arr); err != nil {
		return false
	}
	return len(arr) > 0
}

// func sanitizeFileNameLegacy(name string) string {
// 	name = filepath.Base(name)
// 	name = strings.ReplaceAll(name, " ", "_")
// 	return name
// }

func sanitizeFileName(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, " ", "_")
	return name
}

func saveUploadedFile(r *http.Request, formField, uploadDir string) (string, error) {
	file, fileHeader, err := r.FormFile(formField)
	if err != nil {
		return "", err
	}
	defer file.Close()

	filename := sanitizeFileName(fileHeader.Filename)
	fullPath := filepath.Join(uploadDir, filename)

	if _, err := os.Stat(fullPath); err == nil {
		ext := filepath.Ext(filename)
		name := strings.TrimSuffix(filename, ext)

		counter := 1
		for {
			newName := fmt.Sprintf("%s(%d)%s", name, counter, ext)
			newPath := filepath.Join(uploadDir, newName)
			if _, err := os.Stat(newPath); os.IsNotExist(err) {
				fullPath = newPath
				filename = newName
				break
			}
			counter++
		}
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return "", err
	}

	return filename, nil
}

// func removeFileHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	var request struct {
// 		Filename  string `json:"filename"`
// 		Type      string `json:"type"`
// 		ProductID string `json:"productId"`
// 	}

// 	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	var partNo string
// 	err := db.QueryRow("SELECT partNo FROM products WHERE id = ?", request.ProductID).Scan(&partNo)
// 	if err != nil {
// 		http.Error(w, "Error getting product: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	var fileJSON string
// 	var updateField string

// 	switch request.Type {
// 	case "photos":
// 		updateField = "photos"
// 	case "drawings":
// 		updateField = "drawing_2d"
// 	case "cad":
// 		updateField = "cad_3d"
// 	case "cnc":
// 		updateField = "cnc_code"
// 	case "invoice":
// 		updateField = "invoice"
// 	default:
// 		http.Error(w, "Invalid file type", http.StatusBadRequest)
// 		return
// 	}

// 	err = db.QueryRow("SELECT "+updateField+" FROM products WHERE id = ?", request.ProductID).Scan(&fileJSON)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	var files []FileInfo
// 	if err := json.Unmarshal([]byte(fileJSON), &files); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	var newFiles []FileInfo
// 	var fileToRemove *FileInfo
// 	for _, file := range files {
// 		if file.Name == request.Filename {
// 			fileToRemove = &file
// 		} else {
// 			newFiles = append(newFiles, file)
// 		}
// 	}

// 	if fileToRemove == nil {
// 		http.Error(w, "File not found", http.StatusNotFound)
// 		return
// 	}

// 	fullPath := filepath.Join(uploadDir, fileToRemove.Path)
// 	if err := os.Remove(fullPath); err != nil {
// 		log.Printf("Warning: Could not remove file %s: %v", fullPath, err)
// 	}

// 	newFileJSON, err := json.Marshal(newFiles)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	_, err = db.Exec("UPDATE products SET "+updateField+" = ? WHERE id = ?", string(newFileJSON), request.ProductID)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte("File removed successfully"))
// }

func removeFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Filename  string `json:"filename"`
		Type      string `json:"type"`
		ProductID string `json:"productId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Remove file request - ProductID: %s, Type: %s, Filename: %s",
		request.ProductID, request.Type, request.Filename)

	// Get partNo for the product
	var partNo string
	err := db.QueryRow("SELECT partNo FROM products WHERE id = ?", request.ProductID).Scan(&partNo)
	if err != nil {
		log.Printf("Error getting product: %v", err)
		http.Error(w, "Error getting product: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Determine which field to update
	var updateField string
	switch request.Type {
	case "photos":
		updateField = "photos"
	case "drawings":
		updateField = "drawing_2d"
	case "cad":
		updateField = "cad_3d"
	case "cnc":
		updateField = "cnc_code"
	case "invoice":
		updateField = "invoice"
	default:
		log.Printf("Invalid file type: %s", request.Type)
		http.Error(w, "Invalid file type", http.StatusBadRequest)
		return
	}

	// Get current file list
	var fileJSON string
	err = db.QueryRow("SELECT "+updateField+" FROM products WHERE id = ?", request.ProductID).Scan(&fileJSON)
	if err != nil {
		log.Printf("Error getting file list: %v", err)
		http.Error(w, "Error getting file list: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Current file JSON: %s", fileJSON)

	// Parse current files
	var files []FileInfo
	if fileJSON != "" && fileJSON != "null" {
		if err := json.Unmarshal([]byte(fileJSON), &files); err != nil {
			log.Printf("Error unmarshaling file list: %v", err)
			http.Error(w, "Error parsing file list: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	log.Printf("Parsed %d files", len(files))

	// Find and remove the file from the list
	var newFiles []FileInfo
	var fileToRemove *FileInfo
	for i, file := range files {
		log.Printf("Checking file %d: %s (looking for %s)", i, file.Name, request.Filename)
		if file.Name == request.Filename {
			fileToRemove = &files[i]
			log.Printf("Found file to remove: %s at path: %s", file.Name, file.Path)
		} else {
			newFiles = append(newFiles, file)
		}
	}

	if fileToRemove == nil {
		log.Printf("File not found in database: %s", request.Filename)
		http.Error(w, "File not found in database", http.StatusNotFound)
		return
	}

	// Delete physical file
	fullPath := filepath.Join(uploadDir, fileToRemove.Path)
	log.Printf("Attempting to delete file at: %s", fullPath)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			log.Printf("File does not exist: %s", fullPath)
		} else {
			log.Printf("Warning: Could not remove file %s: %v", fullPath, err)
		}
	} else {
		log.Printf("Successfully deleted physical file: %s", fullPath)
	}

	// Update database with new file list
	newFileJSON, err := json.Marshal(newFiles)
	if err != nil {
		log.Printf("Error marshaling new file list: %v", err)
		http.Error(w, "Error updating file list: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("New file JSON: %s", string(newFileJSON))

	query := fmt.Sprintf("UPDATE products SET %s = ? WHERE id = ?", updateField)
	result, err := db.Exec(query, string(newFileJSON), request.ProductID)
	if err != nil {
		log.Printf("Error updating database: %v", err)
		http.Error(w, "Error updating database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("Database updated successfully, rows affected: %d", rowsAffected)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "File removed successfully",
	})
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/delete/")

	row := db.QueryRow("SELECT partNo FROM products WHERE id = ?", id)
	var partNo string
	err := row.Scan(&partNo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	partNoDir := filepath.Join(uploadDir, sanitizeFilename(partNo))
	if _, err := os.Stat(partNoDir); err == nil {
		if err := os.RemoveAll(partNoDir); err != nil {
			log.Printf("Error removing partNo directory %s: %v", partNoDir, err)
		} else {
			log.Printf("Successfully deleted partNo directory: %s", partNoDir)
		}
	}

	_, err = db.Exec("DELETE FROM products WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func exportHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Export request received")

	f := excelize.NewFile()

	sheetName := "Products"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		log.Printf("Error creating sheet: %v", err)
		http.Error(w, "Failed to create Excel sheet", http.StatusInternalServerError)
		return
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	headers := []string{"PartNo", "PartName", "Description", "Cost", "Quantity", "Material",
		"Material Size", "Material Cost", "Finishing Type", "Finishing Cost", "Created At", "Updated At"}
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 12},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#C6EFCE"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "bottom", Color: "#000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	if err != nil {
		log.Printf("Error creating header style: %v", err)
	}

	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	rows, err := db.Query(`
		SELECT partNo, partName, description, cost, qty, material,
			   material_size, material_cost, finishing_type, finishing_cost,
			   created_at, updated_at 
		FROM products ORDER BY partNo`)
	if err != nil {
		log.Printf("Error querying products: %v", err)
		http.Error(w, "Failed to query products", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	rowNum := 2
	productCount := 0
	for rows.Next() {
		var qty int
		var partNo, partName, description, cost, material string
		var materialSize, materialCost, finishingType, finishingCost string
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&partNo, &partName, &description, &cost, &qty, &material,
			&materialSize, &materialCost, &finishingType, &finishingCost, &createdAt, &updatedAt); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), partNo)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), partName)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), description)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), cost)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), qty)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), material)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), materialSize)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowNum), materialCost)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", rowNum), finishingType)
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", rowNum), finishingCost)
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", rowNum), createdAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue(sheetName, fmt.Sprintf("L%d", rowNum), updatedAt.Format("2006-01-02 15:04:05"))

		rowNum++
		productCount++
	}

	log.Printf("Added %d products to Excel export", productCount)

	for i := range headers {
		col := string('A' + i)
		width := 15.0
		if i == 1 { // PartName
			width = 30.0
		} else if i == 2 { // Description
			width = 40.0
		} else if i == 5 { // Material
			width = 20.0
		} else if i == 6 { // Material Size
			width = 15.0
		} else if i == 8 { // Finishing Type
			width = 20.0
		}
		f.SetColWidth(sheetName, col, col, width)
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=products.xlsx")

	if err := f.Write(w); err != nil {
		log.Printf("Error writing Excel file: %v", err)
		http.Error(w, "Failed to generate Excel file", http.StatusInternalServerError)
		return
	}

	log.Println("Excel export completed successfully")
}

func openFolderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		PartNo string `json:"partNo"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.PartNo == "" {
		http.Error(w, "PartNo is required", http.StatusBadRequest)
		return
	}

	folderPath := filepath.Join(uploadDir, sanitizeFilename(request.PartNo))

	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		http.Error(w, "Folder does not exist", http.StatusNotFound)
		return
	}

	absPath, err := filepath.Abs(folderPath)
	if err != nil {
		http.Error(w, "Error getting absolute path", http.StatusInternalServerError)
		return
	}

	cmd := exec.Command("explorer", absPath)
	if err := cmd.Start(); err != nil {
		http.Error(w, "Failed to open folder: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Folder opened successfully"))
}

func openFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		FilePath string `json:"filePath"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.FilePath == "" {
		http.Error(w, "FilePath is required", http.StatusBadRequest)
		return
	}

	// Construct full file path
	fullPath := filepath.Join(uploadDir, request.FilePath)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.Error(w, "File does not exist", http.StatusNotFound)
		return
	}

	// Get absolute path
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		http.Error(w, "Error getting absolute path", http.StatusInternalServerError)
		return
	}

	// Open file with default Windows program
	// For Windows, use 'cmd /c start "" "filepath"'
	cmd := exec.Command("cmd", "/c", "start", "", absPath)
	if err := cmd.Start(); err != nil {
		log.Printf("Error opening file: %v", err)
		http.Error(w, "Failed to open file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "File opened successfully",
	})
}
