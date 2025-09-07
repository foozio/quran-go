package main

import (
  "context"
  "html/template"
  "net/http"
  "strconv"

  "github.com/jmoiron/sqlx"
  _ "modernc.org/sqlite"
)

var tpl = template.Must(template.New("base").Parse(`
<!doctype html><html><head>
<meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>Quran Learn</title>
<script src="https://unpkg.com/htmx.org@1.9.12"></script>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css">
<style>
  .ayah{padding:.5rem;border-radius:.5rem}
  .ayah:hover{background:#f6f6f6}
  .ar{font-size:1.6rem;line-height:2.2rem;direction:rtl;text-align:right}
  .tj{opacity:.75}
</style>
</head><body class="container">
<header><hgroup><h1>Quran Learn</h1><p>Search • Read • Listen • Review</p></hgroup>
<input name="q" id="q" placeholder="Search Arabic or translation…" hx-get="/search" hx-target="#results" hx-trigger="keyup changed delay:400ms" />
</header>
<main class="grid">
<section>
  <h3>Surah</h3>
  <div id="surah-list">{{range .Surah}}<a href="/s/{{.Number}}">[{{.Number}}] {{.NameAr}}</a><br/>{{end}}</div>
</section>
<section>
  <div id="results"><em>Type to search…</em></div>
</section>
</main>
</body></html>
`))

func main(){
  ctx := context.Background()
  db, _ := sqlx.Open("sqlite", "quran.db")
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
    type srow struct{ Number int `db:"number"`; NameAr string `db:"name_ar"`}
    var list []srow
    _ = db.Select(&list, `SELECT number,name_ar FROM surah ORDER BY number`)
    _ = tpl.Execute(w, map[string]any{"Surah": list})
  })
  http.HandleFunc("/s/", func(w http.ResponseWriter, r *http.Request){
    n, _ := strconv.Atoi(r.URL.Path[len("/s/"):])
    type row struct{ Number int `db:"number"`; Arabic, Tajweed, Trans, Audio string `db:"arabic","tajweed","trans","audio_url"`}
    var rows []row
    _ = db.Select(&rows, `SELECT number,arabic,tajweed,trans,audio_url FROM ayah WHERE surah=?`, n)
    w.Header().Set("Content-Type","text/html; charset=utf-8")
    for _, a := range rows {
      w.Write([]byte(`<div class="ayah"><div class="ar">`+a.Arabic+`</div>`))
      if a.Tajweed != "" { w.Write([]byte(`<div class="tj">`+a.Tajweed+`</div>`)) }
      if a.Trans   != "" { w.Write([]byte(`<div>`+template.HTMLEscapeString(a.Trans)+`</div>`)) }
      if a.Audio   != "" { w.Write([]byte(`<audio controls preload="none" src="`+a.Audio+`"></audio>`)) }
      w.Write([]byte(`</div>`))
    }
  })
  http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request){
    q := r.URL.Query().Get("q")
    type hit struct{ Surah, Number int; Snip string }
    hits := []hit{}
    _ = db.Select(&hits, `
      SELECT ayah.surah AS surah, ayah.number as number,
             snippet(ayah_fts, 2, '<mark>','</mark>','…', 10) AS snip
      FROM ayah_fts JOIN ayah ON ayah_fts.rowid = ayah.rowid
      WHERE ayah_fts MATCH ?
      LIMIT 50`, q)
    w.Header().Set("Content-Type","text/html; charset=utf-8")
    if len(hits)==0 { w.Write([]byte("<em>No results.</em>")); return }
    for _, h := range hits {
      w.Write([]byte(
        `<div><a href="/s/`+strconv.Itoa(h.Surah)+`">`+
        `Surah `+strconv.Itoa(h.Surah)+`:`+strconv.Itoa(h.Number)+`</a> — `+h.Snip+`</div>`))
    }
  })
  _ = http.ListenAndServe(":8090", nil)
  _ = ctx
}
