package secretary

import (
	"fmt"
	"testing"
)

func TestQuery(t *testing.T) {
	queries := []string{
		"SELECT id, name FROM users WHERE age > 30",
		"INSERT INTO products (name, price) VALUES ('Laptop', 1200.99)",
		"UPDATE employees SET salary = 50000 WHERE department = 'IT'",
		"DELETE FROM orders WHERE status = 'cancelled'",
	}

	for _, query := range queries {
		stmt := ParseSQL(query)
		fmt.Printf("Parsed Statement:\n%+v\n\n", stmt)
	}

	t.Fatal()
}
