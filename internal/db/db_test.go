package db_test

import (
  "context"
  "testing"

  "github.com/jmoiron/sqlx"
  _ "modernc.org/sqlite"

  mydb "github.com/foozio/quran-go/internal/db"
)

func must(t *testing.T, err error) {
  t.Helper()
  if err != nil { t.Fatalf("unexpected error: %v", err) }
}

func setupDB(t *testing.T) *sqlx.DB {
  t.Helper()
  d, err := sqlx.Open("sqlite", ":memory:")
  must(t, err)
  must(t, mydb.Migrate(context.Background(), d))
  return d
}

func TestSearchAyah_Basic(t *testing.T) {
  d := setupDB(t)
  // minimal data
  d.MustExec(`INSERT INTO surah(number,name_ar,verses_count) VALUES(1,'الفاتحة',7)`) 
  d.MustExec(`INSERT INTO ayah(surah,number,juz,arabic,tajweed,trans,audio_url) VALUES(1,1,1,'الحمد لله','', 'Segala puji bagi Allah', '')`)

  hits, err := mydb.SearchAyah(context.Background(), d, "Allah", 10)
  must(t, err)
  if len(hits) == 0 { t.Fatalf("expected at least 1 hit") }
  if hits[0].Surah != 1 || hits[0].Number != 1 {
    t.Fatalf("unexpected first hit: %+v", hits[0])
  }
}

func TestSearchAyah_EmptyQueryWildcard(t *testing.T) {
  d := setupDB(t)
  d.MustExec(`INSERT INTO surah(number,name_ar,verses_count) VALUES(2,'البقرة',286)`) 
  d.MustExec(`INSERT INTO ayah(surah,number,juz,arabic,tajweed,trans,audio_url) VALUES(2,2,1,'ذَٰلِكَ ٱلْكِتَٰبُ','', 'It is a Book', '')`)

  hits, err := mydb.SearchAyah(context.Background(), d, "", 10)
  must(t, err)
  if len(hits) == 0 { t.Fatalf("expected hits for wildcard search") }
}

