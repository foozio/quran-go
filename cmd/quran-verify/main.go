package main

import (
  "context"
  "fmt"
  "log"
  "os"

  "github.com/jmoiron/sqlx"
  _ "modernc.org/sqlite"
)

func main(){
  ctx := context.Background()
  path := os.Getenv("QURAN_DB_PATH")
  if path == "" { path = "quran.db" }
  db, err := sqlx.Open("sqlite", path)
  if err != nil { log.Fatalf("open db: %v", err) }
  defer db.Close()

  var surahTotal, ayahTotal, tail int
  must(db.GetContext(ctx, &surahTotal, `SELECT COUNT(*) FROM surah`))
  must(db.GetContext(ctx, &ayahTotal, `SELECT COUNT(*) FROM ayah`))
  must(db.GetContext(ctx, &tail, `SELECT COUNT(*) FROM ayah WHERE surah BETWEEN 94 AND 114`))

  type row struct{
    Surah int `db:"Surah"`
    Verses int `db:"Verses"`
    Ayah int `db:"Ayah"`
  }
  rows := []row{}
  must(db.SelectContext(ctx, &rows, `
    SELECT s.number AS Surah, s.verses_count AS Verses,
      (SELECT COUNT(*) FROM ayah a WHERE a.surah = s.number) AS Ayah
    FROM surah s ORDER BY s.number`))

  bad := 0
  for _, r := range rows { if r.Verses != r.Ayah { fmt.Printf("mismatch: surah %d verses=%d ayah=%d\n", r.Surah, r.Verses, r.Ayah); bad++ } }

  if surahTotal != 114 || bad > 0 || tail == 0 {
    log.Fatalf("verify failed: surah=%d mismatches=%d tail_ayah=%d", surahTotal, bad, tail)
  }
  fmt.Println("verify: OK (114 surah; counts consistent)")
}

func must(err error){ if err != nil { log.Fatal(err) } }

