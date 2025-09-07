package main

import (
  "context"
  "flag"
  "fmt"
  "os"
  "strings"

  "github.com/jmoiron/sqlx"
  "github.com/foozio/quran-go/internal/db"
)

func main() {
  ctx := context.Background()
  path := os.Getenv("QURAN_DB_PATH")
  if path == "" { path = "quran.db" }

  d, err := db.Open(path); must(err)
  must(db.Migrate(ctx, d))

  if len(os.Args) < 2 {
    usage()
    return
  }
  cmd := os.Args[1]
  switch cmd {
  case "list":
    listSurah(d)
  case "surah":
    flags := flag.NewFlagSet("surah", flag.ExitOnError)
    n := flags.Int("n", 1, "surah number (1-114)")
    _ = flags.Parse(os.Args[2:])
    getSurah(d, *n)
  case "search":
    q := strings.Join(os.Args[2:], " ")
    if strings.TrimSpace(q) == "" { fmt.Println("Usage: quran-cli search <query>"); return }
    search(ctx, d, q)
  case "help", "-h", "--help":
    usage()
  default:
    fmt.Println("Unknown command:", cmd)
    usage()
  }
}

func usage() {
  fmt.Println("quran-cli — simple Quran CLI")
  fmt.Println("Commands:")
  fmt.Println("  list                 List all surah")
  fmt.Println("  surah -n <N>         Show ayah for surah N")
  fmt.Println("  search <query>       Search Arabic/translation")
}

func listSurah(d *sqlx.DB) {
  type row struct{ Number int `db:"number"`; NameAr string `db:"name_ar"`; Verses int `db:"verses_count"` }
  var rows []row
  _ = d.Select(&rows, `SELECT number,name_ar,verses_count FROM surah ORDER BY number`)
  for _, s := range rows {
    fmt.Printf("[%3d] %-32s (%d)\n", s.Number, s.NameAr, s.Verses)
  }
}

func getSurah(d *sqlx.DB, n int) {
  fmt.Printf("Surah %d\n", n)
  type row struct{
    Number int    `db:"number"`
    Arabic string `db:"arabic"`
    Tajweed string `db:"tajweed"`
    Trans  string `db:"trans"`
    Audio  string `db:"audio_url"`
  }
  var rows []row
  _ = d.Select(&rows, `SELECT number,arabic,tajweed,trans,audio_url FROM ayah WHERE surah=? ORDER BY number`, n)
  for _, a := range rows {
    fmt.Printf("%d:%d\n%s\n", n, a.Number, a.Arabic)
    if strings.TrimSpace(a.Trans) != "" { fmt.Printf("  %s\n", a.Trans) }
  }
}

func search(ctx context.Context, d *sqlx.DB, q string) {
  hits, _ := db.SearchAyah(ctx, d, q, 50)
  for _, h := range hits {
    fmt.Printf("%d:%d  %s\n", h.Surah, h.Number, stripHTML(h.Snip))
  }
}

func stripHTML(s string) string {
  s = strings.ReplaceAll(s, "<b>", "")
  s = strings.ReplaceAll(s, "</b>", "")
  s = strings.ReplaceAll(s, "<mark>", "")
  s = strings.ReplaceAll(s, "</mark>", "")
  return s
}

func must(err error){ if err != nil { panic(err) } }
