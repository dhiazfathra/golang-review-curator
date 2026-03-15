package product

import "time"

type Product struct {
	ID         string    `db:"id"`
	Name       string    `db:"name"`
	Platform   string    `db:"platform"`
	ProductURL string    `db:"product_url"`
	ProductID  string    `db:"product_id"`
	Active     bool      `db:"active"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}
