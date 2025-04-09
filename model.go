package main

import (
	"database/sql"
)

type product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func (p *product) getProduct(db *sql.DB) error {
	return db.QueryRow("SELECT name, price FROM products WHERE id=$1",
		p.ID).Scan(&p.Name, &p.Price)
}

func (p *product) updateProduct(db *sql.DB) error {
	_, err :=
		db.Exec("UPDATE products SET name=$1, price=$2 WHERE id=$3",
			p.Name, p.Price, p.ID)

	return err
}

func (p *product) deleteProduct(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM products WHERE id=$1", p.ID)

	return err
}

func (p *product) createProduct(db *sql.DB) error {
	err := db.QueryRow(
		"INSERT INTO products(name, price) VALUES($1, $2) RETURNING id",
		p.Name, p.Price).Scan(&p.ID)

	if err != nil {
		return err
	}

	return nil
}

func getProducts(db *sql.DB, start, count int) ([]product, error) {
	rows, err := db.Query(
		"SELECT id, name,  price FROM products LIMIT $1 OFFSET $2",
		count, start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	products := []product{}

	for rows.Next() {
		var p product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

type productStats struct {
	Cheapest      product `json:"cheapest"`
	MostExpensive product `json:"most_expensive"`
}

func getProductStats(db *sql.DB) (productStats, error) {
	var stats productStats

	err := db.QueryRow(
		"SELECT id, name, price FROM products ORDER BY price ASC LIMIT 1").
		Scan(&stats.Cheapest.ID, &stats.Cheapest.Name, &stats.Cheapest.Price)
	if err != nil && err != sql.ErrNoRows {
		return stats, err
	}

	err = db.QueryRow(
		"SELECT id, name, price FROM products ORDER BY price DESC LIMIT 1").
		Scan(&stats.MostExpensive.ID, &stats.MostExpensive.Name, &stats.MostExpensive.Price)
	if err != nil && err != sql.ErrNoRows {
		return stats, err
	}

	return stats, nil
}

func searchProductsByName(db *sql.DB, name string) ([]product, error) {
	rows, err := db.Query("SELECT id, name, price FROM products WHERE name ILIKE '%' || $1 || '%'", name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	products := []product{}
	for rows.Next() {
		var p product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}
