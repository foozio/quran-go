package main

import (
  "context"
  "flag"
  "html/template"
  "net/http"
  "os"
  "strconv"
  "strings"
  "time"

  "github.com/jmoiron/sqlx"
  _ "modernc.org/sqlite"
)

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

func main(){
  ctx := context.Background()
  selfcheck := flag.Bool("selfcheck", false, "run healthcheck and exit")
  flag.Parse()

  bind := os.Getenv("QURAN_BIND")
  if bind == "" { bind = ":8090" }
  path := os.Getenv("QURAN_DB_PATH")
  if path == "" { path = "quran.db" }
  db, err := sqlx.Open("sqlite", path)
  if err != nil { panic(err) }
  defer db.Close()
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
  http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request){
    w.Header().Set("Content-Type","application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"ok":true}`))
  })
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
    type srow struct{ Number int `db:"number"`; NameAr string `db:"name_ar"`}
    var list []srow
    if err := db.SelectContext(r.Context(), &list, `SELECT number,name_ar FROM surah ORDER BY number`); err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError); return
    }
    if err := tpl.Execute(w, map[string]any{"Surah": list}); err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
  })
  http.HandleFunc("/s/", func(w http.ResponseWriter, r *http.Request){
    n, _ := strconv.Atoi(r.URL.Path[len("/s/"):])
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
    w.Write([]byte(`<div id="content">`))
    for _, a := range rows {
      ref := strconv.Itoa(n)+":"+strconv.Itoa(a.Number)
      w.Write([]byte(`<div class="ayah"><div class="row"><button class="btn" data-ref="`+ref+`">★</button><small class="muted">`+ref+`</small></div>`))
      w.Write([]byte(`<div class="ar">`+a.Arabic+`</div>`))
      if a.Tajweed != "" { w.Write([]byte(`<div class="tj">`+a.Tajweed+`</div>`)) }
      if a.Trans   != "" { w.Write([]byte(`<div>`+template.HTMLEscapeString(a.Trans)+`</div>`)) }
      if a.Audio   != "" { w.Write([]byte(`<audio controls preload="none" src="`+a.Audio+`"></audio>`)) }
      w.Write([]byte(`</div>`))
    }
    w.Write([]byte(`</div>`))
  })
  http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request){
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
    if len(hits)==0 { w.Write([]byte("<em class='muted'>No results.</em>")); return }
    for _, h := range hits {
      w.Write([]byte(
        `<div><a href="/s/`+strconv.Itoa(h.Surah)+`">`+
        `Surah `+strconv.Itoa(h.Surah)+`:`+strconv.Itoa(h.Number)+`</a> — `+h.Snip+`</div>`))
    }
  })
  _ = http.ListenAndServe(bind, nil)
  _ = ctx
}
