package main

import (
  "context"
  "net/http"
  "time"

  "github.com/gin-gonic/gin"
  "github.com/jmoiron/sqlx"
  "github.com/foozio/quran-go/internal/db"
)

func main() {
  ctx := context.Background()
  d, err := db.Open("quran.db"); must(err)
  must(db.Migrate(ctx, d))
  // Initial ingest (id translations). In production, run `quran-indexer` separately.
  // _ = data.IngestAll(ctx, d, "id")

  r := gin.Default()
  r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"ok":true}) })
  r.GET("/surah", func(c *gin.Context) {
    var rows []struct{
      Number int `db:"number" json:"number"`
      NameAr string `db:"name_ar" json:"name_ar"`
      NameLatin *string `db:"name_latin" json:"name_latin,omitempty"`
      Revelation *string `db:"revelation" json:"revelation,omitempty"`
      Verses int `db:"verses_count" json:"verses"`
    }
    _ = d.Select(&rows, `SELECT number,name_ar,name_latin,revelation,verses_count FROM surah ORDER BY number`)
    c.JSON(200, rows)
  })
  r.GET("/surah/:n", func(c *gin.Context) {
    n := c.Param("n")
    var out []map[string]any
    _ = d.Select(&out, `SELECT number as ayah, arabic, tajweed, trans, audio_url FROM ayah WHERE surah=? ORDER BY number`, n)
    c.JSON(200, gin.H{"surah": n, "ayah": out})
  })
  r.GET("/search", func(c *gin.Context) {
    q := c.Query("q")
    rows, _ := db.SearchAyah(c, d, q, 50)
    c.JSON(200, gin.H{"q": q, "hits": rows})
  })

  s := &http.Server{ Addr: ":8080", Handler: r, ReadTimeout: 10*time.Second, WriteTimeout: 20*time.Second }
  must(s.ListenAndServe())
}
func must(err error){ if err != nil { panic(err) } }
