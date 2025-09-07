package main

import (
  "bufio"
  "fmt"
  "os"
  "strings"

  "github.com/jmoiron/sqlx"
  _ "modernc.org/sqlite"
)

func main(){
  db, _ := sqlx.Open("sqlite","quran.db")
  in := bufio.NewScanner(os.Stdin)
  fmt.Println("quran-cli :: type 's <query>' to search, 'r <surah:ayah>' to review, 'q' to quit")
  for {
    fmt.Print("> ")
    if !in.Scan() { return }
    cmd := strings.TrimSpace(in.Text())
    switch {
    case cmd=="q":
      return
    case strings.HasPrefix(cmd, "s "):
      q := strings.TrimSpace(strings.TrimPrefix(cmd,"s "))
      var hits []struct{ Surah, Number int; Snip string }
      _ = db.Select(&hits, `SELECT ayah.surah, ayah.number, snippet(ayah_fts,2,'[',']','…',10) snip
                             FROM ayah_fts JOIN ayah ON ayah_fts.rowid=ayah.rowid
                             WHERE ayah_fts MATCH ? LIMIT 20`, q)
      for _, h := range hits {
        fmt.Printf("(%d:%d) %s\n", h.Surah, h.Number, h.Snip)
      }
    case strings.HasPrefix(cmd, "r "):
      ref := strings.TrimSpace(strings.TrimPrefix(cmd,"r "))
      var s, a int
      fmt.Sscanf(ref, "%d:%d", &s, &a)
      var row struct{ Arabic, Trans string }
      _ = db.Get(&row, `SELECT arabic,trans FROM ayah WHERE surah=? AND number=?`, s, a)
      fmt.Println(row.Arabic)
      if row.Trans != "" { fmt.Println(row.Trans) }
      fmt.Println("[enter quality 0–5]")
      in.Scan()
      // TODO: feed into SM-2 algorithm
    default:
      fmt.Println("unknown. try: s <query> | r <surah:ayah> | q")
    }
  }
}
