package db

import (
  "context"
  "embed"
  "strings"

  "github.com/jmoiron/sqlx"
  _ "modernc.org/sqlite"
)

//go:embed migrate.sql
var migrations embed.FS

func Open(path string) (*sqlx.DB, error) {
  db, err := sqlx.Open("sqlite", path)
  if err != nil { return nil, err }
  if _, err = db.Exec(`PRAGMA journal_mode=WAL;`); err != nil { return nil, err }
  return db, nil
}

func Migrate(ctx context.Context, db *sqlx.DB) error {
  b, err := migrations.ReadFile("migrate.sql"); if err != nil { return err }
  _, err = db.ExecContext(ctx, string(b))
  return err
}

func SearchAyah(ctx context.Context, db *sqlx.DB, q string, limit int) ([]struct{
  Surah int `db:"surah"`
  Number int `db:"number"`
  Snip string `db:"snip"`
}, error) {
  q = strings.TrimSpace(q)
  if q == "" { q = "*" }
  rows := []struct{ Surah, Number int; Snip string }{}
  err := db.SelectContext(ctx, &rows, `
    SELECT ayah.surah, ayah.number,
      snippet(ayah_fts, 2, '<b>','</b>','…', 10) AS snip
    FROM ayah_fts JOIN ayah ON ayah_fts.rowid = ayah.rowid
    WHERE ayah_fts MATCH ?
    LIMIT ?`, q, limit)
  return rows, err
}
