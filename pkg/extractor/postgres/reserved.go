package postgres

import (
	"context"
	"database/sql"
)

type words []string

const RESERVED_WORDS = 491

type ReservedGetter interface {
	ListReservedWord() []string
}

func InitReservedWords(ctx context.Context, db *sql.DB) ReservedGetter {
	reserved, err := listReservedWord(ctx, db)
	if err != nil {
		panic(err)
	}
	return reserved
}

func listReservedWord(ctx context.Context, db *sql.DB) (words, error) {
	result, err := db.QueryContext(
		ctx,
		`
		SELECT word
		FROM pg_get_keywords()
		`,
	)
	if err != nil {
		return nil, err
	}
	defer result.Close()
	words := make([]string, RESERVED_WORDS)
	for result.Next() {
		word := new(string)
		if err := result.Scan(word); err != nil {
			return nil, err
		}
		words = append(words, *word)
	}
	return words, nil
}

func (ws words) ListReservedWord() []string {
	return ws
}
