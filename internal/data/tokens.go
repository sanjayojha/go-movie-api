package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"time"

	"movieapi.sanjayojha.dev/internal/validator"
)

const (
	ScopeActivation = "activation"
)

type Token struct {
	PlainText string
	Hash      []byte
	UserID    int
	Expiry    time.Time
	Scope     string
}

type TokenModel struct {
	DB *sql.DB
}

func generateToken(userID int, ttl time.Duration, scope string) *Token {

	token := &Token{
		PlainText: rand.Text(),
		UserID:    userID,
		Expiry:    time.Now().Add(ttl),
		Scope:     scope,
	}

	hash := sha256.Sum256([]byte(token.PlainText))

	// sha256.Sum256() function returns an *array* of length 32, so to make it easier to work with we convert it to a slice using the [:] operator before storing it.
	token.Hash = hash[:]

	return token
}

func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

func (m TokenModel) New(userID int, ttl time.Duration, scope string) (*Token, error) {

	token := generateToken(userID, ttl, scope)
	err := m.Insert(token)
	return token, err
}

func (m TokenModel) Insert(token *Token) error {
	query := `
	INSERT INTO tokens (hash, user_id, expiry, scope)
	VALUES ($1, $2, $3, $4)
	`

	args := []any{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.QueryContext(ctx, query, args...)

	return err
}

// deletes all tokens for a specific user and scope.
func (m TokenModel) DeleteAllForUser(scope string, userID int) error {
	query := `
	DELETE FROM tokens
	WHERE scope = $1 AND user_id = $2
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, scope, userID)

	return err
}
