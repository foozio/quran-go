package main

import (
  "context"
  "net/http"
  "os"
  "time"

  "github.com/gin-gonic/gin"
  "github.com/foozio/quran-go/internal/db"
  "github.com/foozio/quran-go/internal/httpx"
)

func main() {
  ctx := context.Background()

  bind := os.Getenv("QURAN_BIND")
  if bind == "" { bind = ":8080" }
  path := os.Getenv("QURAN_DB_PATH")
  if path == "" { path = "quran.db" }

  d, err := db.Open(path); must(err)
  must(db.Migrate(ctx, d))

  r := gin.New()
  r.Use(gin.Recovery())
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

  h := httpx.CORS(r)
  h = httpx.RateLimit(h)
  s := &http.Server{ Addr: bind, Handler: h, ReadTimeout: 10*time.Second, WriteTimeout: 20*time.Second }
  must(s.ListenAndServe())
}

func must(err error){ if err != nil { panic(err) } }
