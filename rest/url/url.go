package url

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encore.dev/storage/sqldb"
)

type URL struct {
	ID  string
	URL string
}

type ShortenParams struct {
	URL string
}

// encore: api public method=POST path=/url
func Shorten(ctx context.Context, p *ShortenParams) (*URL, error) {

	id, err := generateID()
	if err != nil {
		return nil, err
	} else if err := insert(ctx, id, p.URL); err != nil {
		return nil, err
	}
	return &URL{ID: id, URL: p.URL}, nil

}

// encore: api public method=GET path=/url/:id
func Get(ctx context.Context, id string) (*URL, error) {
	u := &URL{ID: id}
	err := db.QueryRow(ctx, `
        SELECT original_url FROM url
        WHERE id = $1
    `, id).Scan(&u.URL)
	return u, err
}

// generateID generates a random short ID.
func generateID() (string, error) {
	var data [6]byte // 6 bytes of entropy
	if _, err := rand.Read(data[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data[:]), nil
}

func insert(ctx context.Context, id, url string) error {
	_, err := db.Exec(ctx, `
        INSERT INTO url (id, original_url)
        VALUES ($1, $2)
    `, id, url)
	return err
}

// Define a database named 'url', using the database
// migrations in the "./migrations" folder.

var db = sqldb.NewDatabase("url", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})
