package main

import (
  "context"
  "flag"
  "fmt"
  "html/template"
  "net/http"
  "os"
  "os/signal"
  "strconv"
  "strings"
  "syscall"
  "time"

  "github.com/gin-gonic/gin"
  "github.com/jmoiron/sqlx"

  qdb "github.com/foozio/quran-go/internal/db"
  "github.com/foozio/quran-go/internal/httpx"
)

func main(){
  ctx := context.Background()

  // Flags for container healthcheck
  selfcheck := flag.Bool("selfcheck", false, "run healthcheck and exit")
  flag.Parse()

  apiBind := os.Getenv("QURAN_API_BIND")
  if apiBind == "" { apiBind = os.Getenv("QURAN_BIND") }
  if apiBind == "" { apiBind = ":8080" }
  webBind := os.Getenv("QURAN_WEB_BIND")
  if webBind == "" { webBind = ":8090" }
  path := os.Getenv("QURAN_DB_PATH")
  if path == "" { path = "quran.db" }

  if *selfcheck {
    addr := apiBind
    if strings.HasPrefix(apiBind, ":") {
      addr = "127.0.0.1" + apiBind
    } else if strings.HasPrefix(apiBind, "0.0.0.0:") {
      addr = "127.0.0.1:" + strings.TrimPrefix(apiBind, "0.0.0.0:")
    } else if strings.HasPrefix(apiBind, "[::]:") {
      addr = "127.0.0.1:" + strings.TrimPrefix(apiBind, "[::]:")
    }
    url := "http://" + addr + "/healthz"
    hc := &http.Client{ Timeout: 2 * time.Second }
    resp, err := hc.Get(url)
    if err != nil || resp.StatusCode != http.StatusOK { os.Exit(1) }
    os.Exit(0)
  }

  d, err := qdb.Open(path); must(err)
  must(qdb.Migrate(ctx, d))

  api := buildAPI(d)
  web := buildWeb(d)

  apiSrv := &http.Server{ Addr: apiBind, Handler: api, ReadTimeout: 10*time.Second, WriteTimeout: 20*time.Second }
  webSrv := &http.Server{ Addr: webBind, Handler: web, ReadTimeout: 10*time.Second, WriteTimeout: 20*time.Second }

  go func(){ _ = apiSrv.ListenAndServe() }()
  go func(){ _ = webSrv.ListenAndServe() }()

  // Graceful shutdown
  sigc := make(chan os.Signal, 1)
  signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
  <-sigc
  ctx2, cancel := context.WithTimeout(context.Background(), 3*time.Second); defer cancel()
  _ = apiSrv.Shutdown(ctx2)
  _ = webSrv.Shutdown(ctx2)
}

func must(err error){ if err != nil { panic(err) } }

// API
func buildAPI(d *sqlx.DB) http.Handler {
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
    rows, err := qdb.SearchAyah(c.Request.Context(), d, q, 50)
    if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
    c.JSON(200, gin.H{"q": q, "hits": rows})
  })
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

// Web
var tpl = template.Must(template.New("base").Parse(`
<!doctype html><html lang="en"><head>
<meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>Quran Learn</title>
<script src="https://unpkg.com/htmx.org@1.9.12"></script>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css">
<style>
  :root {
    --bg: #0f1115; --card: #151821; --muted: #a8b3cf; --text: #e5e9f0; --accent: #7aa2f7;
    --radius: 14px; --gap: 16px;
  }
  html, body { background: var(--bg); color: var(--text); }
  header { margin: 2rem 0; }
  .app { display: grid; grid-template-columns: 260px 1fr 320px; gap: var(--gap); }
  .panel { background: var(--card); border-radius: var(--radius); padding: 16px; }
  .panel h3 { margin-top: 0; }
  .surah a { color: var(--muted); text-decoration: none; display: block; padding: 6px 8px; border-radius: 10px; }
  .surah a:hover { background: #1b2030; color: var(--text); }
  .ayah{padding:.5rem .75rem;border-radius:12px;margin-bottom:6px;background:#121621;border:1px solid #1c2233}
  .ayah:hover{border-color:#29324a}
  .ar{font-size:1.6rem;line-height:2.2rem;direction:rtl;text-align:right;margin-bottom:.25rem}
  .tj{opacity:.75;margin:.25rem 0}
  .row { display:flex; align-items:center; gap:10px; }
  .btn { cursor:pointer; border:1px solid #293046; background:#161b26; color:var(--text); border-radius:10px; padding:6px 10px;}
  .btn:hover { border-color:#3a4666; }
  textarea, input[type="text"] { background:#0f131c; color:var(--text); border:1px solid #222a3d; border-radius:10px; }
  mark { background: rgba(122,162,247,.15); color: var(--text); padding:0 .2em; border-radius:4px; }
  small.muted { color: var(--muted); }
</style>
</head><body class="container">
<header>
  <hgroup><h1 style="margin:0">Quran Learn</h1><p class="muted">Search • Read • Listen • Review</p></hgroup>
  <input name="q" id="q" placeholder="Search Arabic or translation…" hx-get="/search" hx-target="#results" hx-trigger="keyup changed delay:400ms" />
</header>

<div class="app">
  <!-- Left: Surah list -->
  <aside class="panel">
    <h3>Surah</h3>
    <div class="surah">{{range .Surah}}<a href="/s/{{.Number}}">[{{.Number}}] {{.NameAr}}</a>{{end}}</div>
  </aside>

  <!-- Middle: Content -->
  <main class="panel">
    <div id="results"><em class="muted">Type to search…</em></div>
    <div id="content" class="content"></div>
  </main>

  <!-- Right: Bookmarks & Notes -->
  <aside class="panel">
    <h3>Bookmarks</h3>
    <div id="bookmarks"></div>
    <hr>
    <h3>Notes</h3>
    <small class="muted">Notes are saved locally in your browser</small>
    <div class="row" style="margin-top:.5rem">
      <input id="note-ref" type="text" placeholder="e.g., 2:255 (Surah:Ayah)" />
      <button class="btn" onclick="saveNote()">Save</button>
    </div>
    <textarea id="note-text" rows="5" placeholder="Write your note..."></textarea>
    <div id="notes" style="margin-top:10px"></div>
  </aside>
</div>

<script>
  const LS_BOOK = 'quran_bookmarks';
  const LS_NOTES = 'quran_notes';

  function loadBookmarks() {
    const list = JSON.parse(localStorage.getItem(LS_BOOK) || "[]");
    const el = document.getElementById('bookmarks');
    el.innerHTML = list.length ? "" : "<em class='muted'>No bookmarks yet</em>";
    list.forEach(ref => {
      const a = document.createElement('a');
      a.href = "/s/" + ref.split(":")[0];
      a.textContent = ref;
      a.style.display = "block";
      a.style.color = "#a8b3cf";
      el.appendChild(a);
    });
  }
  function toggleBookmark(ref) {
    const list = new Set(JSON.parse(localStorage.getItem(LS_BOOK) || "[]"));
    if (list.has(ref)) list.delete(ref); else list.add(ref);
    localStorage.setItem(LS_BOOK, JSON.stringify(Array.from(list)));
    loadBookmarks();
  }
  function loadNotes() {
    const notes = JSON.parse(localStorage.getItem(LS_NOTES) || "{}");
    const el = document.getElementById('notes');
    el.innerHTML = "";
    Object.keys(notes).sort().forEach(ref => {
      const div = document.createElement('div');
      div.className = 'ayah';
      div.innerHTML = "<b>"+ref+"</b><br>"+notes[ref];
      el.appendChild(div);
    });
  }
  function saveNote() {
    const ref = document.getElementById('note-ref').value.trim();
    const text = document.getElementById('note-text').value.trim();
    if (!ref || !text) return;
    const notes = JSON.parse(localStorage.getItem(LS_NOTES) || "{}");
    notes[ref] = text;
    localStorage.setItem(LS_NOTES, JSON.stringify(notes));
    document.getElementById('note-text').value = "";
    loadNotes();
  }
  document.body.addEventListener('htmx:afterSwap', (e) => {
    if (e.detail.target.id === "content" || e.detail.target.id === "results") {
      document.querySelectorAll('[data-ref]').forEach(btn => {
        btn.addEventListener('click', () => toggleBookmark(btn.getAttribute('data-ref')));
      });
    }
  });

  loadBookmarks(); loadNotes();
</script>
</body></html>
`))

func buildWeb(db *sqlx.DB) http.Handler {
  mux := http.NewServeMux()
  mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request){
    w.Header().Set("Content-Type","application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"ok":true}`))
  })
  mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
    type srow struct{ Number int `db:"number"`; NameAr string `db:"name_ar"`}
    var list []srow
    if err := db.SelectContext(r.Context(), &list, `SELECT number,name_ar FROM surah ORDER BY number`); err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError); return
    }
    if err := tpl.Execute(w, map[string]any{"Surah": list}); err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
  })
  mux.HandleFunc("/s/", func(w http.ResponseWriter, r *http.Request){
    n, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/s/"))
    type row struct{
      Number int `db:"number"`
      Arabic string `db:"arabic"`
      Tajweed string `db:"tajweed"`
      Trans string `db:"trans"`
      Audio string `db:"audio_url"`
    }
    var rows []row
    if err := db.SelectContext(r.Context(), &rows, `SELECT number,arabic,tajweed,trans,audio_url FROM ayah WHERE surah=?`, n); err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError); return
    }
    w.Header().Set("Content-Type","text/html; charset=utf-8")
    _, _ = w.Write([]byte(`<div id="content">`))
    for _, a := range rows {
      ref := fmt.Sprintf("%d:%d", n, a.Number)
      _, _ = w.Write([]byte(`<div class="ayah"><div class="row"><button class="btn" data-ref="`+ref+`">★</button><small class="muted">`+ref+`</small></div>`))
      _, _ = w.Write([]byte(`<div class="ar">`+a.Arabic+`</div>`))
      if a.Tajweed != "" { _, _ = w.Write([]byte(`<div class="tj">`+a.Tajweed+`</div>`)) }
      if a.Trans   != "" { _, _ = w.Write([]byte(`<div>`+template.HTMLEscapeString(a.Trans)+`</div>`)) }
      if a.Audio   != "" { _, _ = w.Write([]byte(`<audio controls preload="none" src="`+a.Audio+`"></audio>`)) }
      _, _ = w.Write([]byte(`</div>`))
    }
    _, _ = w.Write([]byte(`</div>`))
  })
  mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request){
    q := r.URL.Query().Get("q")
    type hit struct{ Surah, Number int; Snip string }
    hits := []hit{}
    if err := db.SelectContext(r.Context(), &hits, `
      SELECT ayah.surah AS surah, ayah.number as number,
             snippet(ayah_fts, 2, '<mark>','</mark>','…', 10) AS snip
      FROM ayah_fts JOIN ayah ON ayah_fts.rowid = ayah.rowid
      WHERE ayah_fts MATCH ?
      LIMIT 50`, q); err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError); return
    }
    w.Header().Set("Content-Type","text/html; charset=utf-8")
    if len(hits)==0 { _, _ = w.Write([]byte("<em class='muted'>No results.</em>")); return }
    for _, h := range hits {
      _, _ = w.Write([]byte(
        `<div><a href="/s/`+strconv.Itoa(h.Surah)+`">`+
        `Surah `+strconv.Itoa(h.Surah)+`:`+strconv.Itoa(h.Number)+`</a> — `+h.Snip+`</div>`))
    }
  })
  return mux
}
