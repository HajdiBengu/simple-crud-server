package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

// Item in the database
type Item struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// In-memory database
type Database struct {
	items map[string]Item
	mu    sync.RWMutex
}

// New Database
func NewDatabase() *Database {
	return &Database{
		items: make(map[string]Item),
	}
}

// Adds a new item to the database
func (db *Database) Create(name string, price float64) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.items[name]; exists {
		return fmt.Errorf("item already exists")
	}

	db.items[name] = Item{Name: name, Price: price}
	return nil
}

// Reads an item
func (db *Database) Read(name string) (Item, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	item, exists := db.items[name]
	if !exists {
		return Item{}, fmt.Errorf("item not found")
	}

	return item, nil
}

// Updates the price of an existing item
func (db *Database) Update(name string, price float64) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.items[name]; !exists {
		return fmt.Errorf("item not found")
	}

	db.items[name] = Item{Name: name, Price: price}
	return nil
}

// Deletes an item
func (db *Database) Delete(name string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.items[name]; !exists {
		return fmt.Errorf("item not found")
	}

	delete(db.items, name)
	return nil
}

// String representation of the database
func (db *Database) VisualizeDatabase() string {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if len(db.items) == 0 {
		return "Database is empty."
	}

	visualization := "Database Contents:\n"
	visualization += "------------------\n"
	for name, item := range db.items {
		visualization += fmt.Sprintf("Item: %s, Price: $%.2f\n", name, item.Price)
	}
	return visualization
}

// Handles the creation of a new item
func CreateHandler(db *Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item := r.URL.Query().Get("item")
		priceStr := r.URL.Query().Get("price")

		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			http.Error(w, "invalid price", http.StatusBadRequest)
			return
		}

		if err := db.Create(item, price); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Item created: %s, Price: $%.2f\n\n%s", item, price, db.VisualizeDatabase())
	}
}

// Handles reading an item
func ReadHandler(db *Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item := r.URL.Query().Get("item")

		itemData, err := db.Read(item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Item found: %s, Price: $%.2f\n\n%s", itemData.Name, itemData.Price, db.VisualizeDatabase())
	}
}

// Handles updating an item's price
func UpdateHandler(db *Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item := r.URL.Query().Get("item")
		priceStr := r.URL.Query().Get("price")

		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			http.Error(w, "invalid price", http.StatusBadRequest)
			return
		}

		if err := db.Update(item, price); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Item updated: %s, New Price: $%.2f\n\n%s", item, price, db.VisualizeDatabase())
	}
}

// Handles deleting an item
func DeleteHandler(db *Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item := r.URL.Query().Get("item")

		if err := db.Delete(item); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Item deleted: %s\n\n%s", item, db.VisualizeDatabase())
	}
}

func main() {
	db := NewDatabase()

	http.HandleFunc("/create", CreateHandler(db))
	http.HandleFunc("/read", ReadHandler(db))
	http.HandleFunc("/update", UpdateHandler(db))
	http.HandleFunc("/delete", DeleteHandler(db))

	fmt.Println("Server started at :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
