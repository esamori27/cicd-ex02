package main_test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	git "github.com/esamori27/cicd-ex02"
)

var a git.App

func TestMain(m *testing.M) {
	a.Initialize(
		"postgres",
		"cicd",
		"postgres")

	ensureTableExists()
	code := m.Run()
	clearTable()
	os.Exit(code)
}

func TestEmptyTable(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/products", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestGetNonExistentProduct(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/product/11", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Product not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Product not found'. Got '%s'", m["error"])
	}
}

func TestCreateProduct(t *testing.T) {
	clearTable()

	var jsonStr = []byte(`{"name":"test product", "price": 11.22}`)
	req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "test product" {
		t.Errorf("Expected product name to be 'test product'. Got '%v'", m["name"])
	}

	if m["price"] != 11.22 {
		t.Errorf("Expected product price to be '11.22'. Got '%v'", m["price"])
	}

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["id"] != 1.0 {
		t.Errorf("Expected product ID to be '1'. Got '%v'", m["id"])
	}
}

func TestGetProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestUpdateProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)
	var originalProduct map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalProduct)

	var jsonStr = []byte(`{"name":"test product - updated name", "price": 11.22}`)
	req, _ = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != originalProduct["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalProduct["id"], m["id"])
	}

	if m["name"] == originalProduct["name"] {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalProduct["name"], m["name"], m["name"])
	}

	if m["price"] == originalProduct["price"] {
		t.Errorf("Expected the price to change from '%v' to '%v'. Got '%v'", originalProduct["price"], m["price"], m["price"])
	}
}

func TestDeleteProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/product/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/product/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestGetProductStats(t *testing.T) {
	clearTable()

	_, err := a.DB.Exec("INSERT INTO products(name, price) VALUES($1, $2)", "Cheap Product", 20.0)
	if err != nil {
		t.Fatal(err)
	}
	_, err = a.DB.Exec("INSERT INTO products(name, price) VALUES($1, $2)", "Mid Product", 50.0)
	if err != nil {
		t.Fatal(err)
	}
	_, err = a.DB.Exec("INSERT INTO products(name, price) VALUES($1, $2)", "Expensive Product", 100.0)
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest("GET", "/products/stats", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var stats map[string]map[string]interface{}
	if err := json.Unmarshal(response.Body.Bytes(), &stats); err != nil {
		t.Fatal(err)
	}

	if name, ok := stats["cheapest"]["name"].(string); !ok || name != "Cheap Product" {
		t.Errorf("Expected cheapest product name to be 'Cheap Product', got '%v'", stats["cheapest"]["name"])
	}
	if name, ok := stats["most_expensive"]["name"].(string); !ok || name != "Expensive Product" {
		t.Errorf("Expected most expensive product name to be 'Expensive Product', got '%v'", stats["most_expensive"]["name"])
	}
}

func TestSearchProductsByName(t *testing.T) {
	clearTable()
	_, err := a.DB.Exec("INSERT INTO products(name, price) VALUES($1, $2)", "T-shirt", 25.0)
	if err != nil {
		t.Fatal(err)
	}
	_, err = a.DB.Exec("INSERT INTO products(name, price) VALUES($1, $2)", "Shirt", 30.0)
	if err != nil {
		t.Fatal(err)
	}
	_, err = a.DB.Exec("INSERT INTO products(name, price) VALUES($1, $2)", "Jeans", 40.0)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest("GET", "/products/search?name=shirt", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	var products []map[string]interface{}
	if err := json.Unmarshal(response.Body.Bytes(), &products); err != nil {
		t.Fatal(err)
	}
	if len(products) != 2 {
		t.Errorf("Expected 2 products, got %d", len(products))
	}
	for _, prod := range products {
		name, ok := prod["name"].(string)
		if !ok || !strings.Contains(strings.ToLower(name), "shirt") {
			t.Errorf("Expected product name to contain 'shirt', got %v", prod["name"])
		}
	}
}

func addProducts(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		a.DB.Exec("INSERT INTO products(name, price) VALUES($1, $2)", "Product "+strconv.Itoa(i), (i+1.0)*10)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	a.DB.Exec("DELETE FROM products")
	a.DB.Exec("ALTER SEQUENCE products_id_seq RESTART WITH 1")
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS products
(
    id SERIAL,
    name TEXT NOT NULL,
    price NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    CONSTRAINT products_pkey PRIMARY KEY (id)
)`
