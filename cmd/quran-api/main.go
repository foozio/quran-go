package main

import (
  "context"
  "flag"
  "net/http"
  "os"
  "strconv"
  "strings"
  "time"

  "github.com/gin-gonic/gin"
  "github.com/jmoiron/sqlx"
  "github.com/foozio/quran-go/internal/db"
  "github.com/foozio/quran-go/internal/httpx"
)

func main() {
  ctx := context.Background()
  selfcheck := flag.Bool("selfcheck", false, "run healthcheck and exit")
  flag.Parse()

  bind := os.Getenv("QURAN_BIND")
  if bind == "" { bind = ":8080" }
  path := os.Getenv("QURAN_DB_PATH")
  if path == "" { path = "quran.db" }

  if *selfcheck {
    addr := bind
    if strings.HasPrefix(bind, ":") {
      addr = "127.0.0.1" + bind
    } else if strings.HasPrefix(bind, "0.0.0.0:") {
      addr = "127.0.0.1:" + strings.TrimPrefix(bind, "0.0.0.0:")
    } else if strings.HasPrefix(bind, "[::]:") {
      addr = "127.0.0.1:" + strings.TrimPrefix(bind, "[::]:")
    }
    url := "http://" + addr + "/healthz"
    hc := &http.Client{ Timeout: 2 * time.Second }
    resp, err := hc.Get(url)
    if err != nil || resp.StatusCode != http.StatusOK { os.Exit(1) }
    os.Exit(0)
  }

  d, err := db.Open(path); must(err)
  must(db.Migrate(ctx, d))

  h := newRouter(d)
  s := &http.Server{ Addr: bind, Handler: h, ReadTimeout: 10*time.Second, WriteTimeout: 20*time.Second }
  must(s.ListenAndServe())
}

func must(err error){ if err != nil { panic(err) } }

// newRouter builds the HTTP router so tests can exercise handlers.
func newRouter(d *sqlx.DB) http.Handler {
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
    if err := d.SelectContext(c.Request.Context(), &rows, `SELECT number,name_ar,name_latin,revelation,verses_count FROM surah ORDER BY number`); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return
    }
    c.JSON(200, rows)
  })
  r.GET("/surah/:n", func(c *gin.Context) {
    nStr := c.Param("n")
    n, err := strconv.Atoi(nStr)
    if err != nil || n < 1 || n > 114 {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid surah number"}); return
    }
    type row struct{
      Ayah int `db:"ayah" json:"ayah"`
      Arabic string `db:"arabic" json:"arabic"`
      Tajweed string `db:"tajweed" json:"tajweed"`
      Trans string `db:"trans" json:"trans"`
      AudioURL string `db:"audio_url" json:"audio_url"`
    }
    var out []row
    if err := d.SelectContext(c.Request.Context(), &out, `SELECT number as ayah, arabic, tajweed, trans, audio_url FROM ayah WHERE surah=? ORDER BY number`, n); err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return
    }
    c.JSON(200, gin.H{"surah": n, "ayah": out})
  })
  r.GET("/search", func(c *gin.Context) {
    q := strings.TrimSpace(c.Query("q"))
    if len(q) > 100 { c.JSON(http.StatusBadRequest, gin.H{"error": "query too long"}); return }
    rows, err := db.SearchAyah(c.Request.Context(), d, q, 50)
    if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
    c.JSON(200, gin.H{"q": q, "hits": rows})
  })

  // Stats endpoint: verifies content consistency at runtime
  r.GET("/stats", func(c *gin.Context) {
    type mismatch struct{
      Surah int `db:"Surah" json:"surah"`
      Verses int `db:"Verses" json:"verses"`
      Ayah int `db:"Ayah" json:"ayah"`
    }
    var surahTotal, ayahTotal, tail int
    if err := d.GetContext(c, &surahTotal, `SELECT COUNT(*) FROM surah`); err != nil {
      c.JSON(500, gin.H{"error": err.Error()}); return
    }
    if err := d.GetContext(c, &ayahTotal, `SELECT COUNT(*) FROM ayah`); err != nil {
      c.JSON(500, gin.H{"error": err.Error()}); return
    }
    if err := d.GetContext(c, &tail, `SELECT COUNT(*) FROM ayah WHERE surah BETWEEN 94 AND 114`); err != nil {
      c.JSON(500, gin.H{"error": err.Error()}); return
    }
    rows := []mismatch{}
    if err := d.SelectContext(c, &rows, `
      SELECT s.number AS Surah, s.verses_count AS Verses,
        (SELECT COUNT(*) FROM ayah a WHERE a.surah = s.number) AS Ayah
      FROM surah s ORDER BY s.number`); err != nil {
      c.JSON(500, gin.H{"error": err.Error()}); return
    }
    bad := []mismatch{}
    for _, r := range rows { if r.Verses != r.Ayah { bad = append(bad, r) } }
    c.JSON(200, gin.H{
      "ok": surahTotal == 114 && len(bad) == 0,
      "surah_total": surahTotal,
      "ayah_total": ayahTotal,
      "tail_94_114_ayah": tail,
      "mismatches": bad,
    })
  })

  h := httpx.CORS(r)
  h = httpx.RateLimit(h)
  return h
}
