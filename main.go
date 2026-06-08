package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ItemRequest struct {
	Name string `json:"name"`
}

var (
	items  []Item
	nextID = 1
	mutex  sync.Mutex
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func itemsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		getItems(w, r)
	case http.MethodPost:
		addItem(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Method tidak diizinkan",
		})
	}
}

func itemByIDHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idText := strings.TrimPrefix(r.URL.Path, "/items/")
	id, err := strconv.Atoi(idText)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "ID tidak valid",
		})
		return
	}

	switch r.Method {
	case http.MethodPut:
		updateItem(w, r, id)
	case http.MethodDelete:
		deleteItem(w, r, id)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Method tidak diizinkan",
		})
	}
}

func getItems(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	json.NewEncoder(w).Encode(items)
}

func addItem(w http.ResponseWriter, r *http.Request) {
	var request ItemRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || strings.TrimSpace(request.Name) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Nama item tidak boleh kosong",
		})
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	newItem := Item{
		ID:   nextID,
		Name: request.Name,
	}

	items = append(items, newItem)
	nextID++

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Item berhasil ditambahkan",
		"data":    newItem,
	})
}

func updateItem(w http.ResponseWriter, r *http.Request, id int) {
	var request ItemRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || strings.TrimSpace(request.Name) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Nama item tidak boleh kosong",
		})
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	for i := range items {
		if items[i].ID == id {
			items[i].Name = request.Name

			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Item berhasil diperbarui",
				"data":    items[i],
			})
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Item tidak ditemukan",
	})
}

func deleteItem(w http.ResponseWriter, r *http.Request, id int) {
	mutex.Lock()
	defer mutex.Unlock()

	for i := range items {
		if items[i].ID == id {
			items = append(items[:i], items[i+1:]...)

			json.NewEncoder(w).Encode(map[string]string{
				"message": "Item berhasil dihapus",
			})
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Item tidak ditemukan",
	})
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/items", itemsHandler)
	http.HandleFunc("/items/", itemByIDHandler)

	log.Println("Server berjalan di port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}