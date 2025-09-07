package main

import (
  "context"
  "fmt"
  "os"
  "strings"

  tea "github.com/charmbracelet/bubbletea"
  "github.com/jmoiron/sqlx"
  "github.com/foozio/quran-go/internal/db"
)

type viewState int

const (
  stateList viewState = iota
  stateSurah
)

type surahRow struct{
  Number int    `db:"number"`
  NameAr string `db:"name_ar"`
  Verses int    `db:"verses_count"`
}
type ayahRow struct{
  Number int    `db:"number"`
  Arabic string `db:"arabic"`
  Tajweed string `db:"tajweed"`
  Trans  string `db:"trans"`
}

type model struct {
  db   *sqlx.DB
  st   viewState
  w, h int

  // list view
  list     []surahRow
  cursor   int
  listOff  int

  // status/error line
  status   string
  lastErr  string

  // surah view
  curSurah int
  ayat     []ayahRow
  ayOff    int
}

func initialModel(d *sqlx.DB) model {
  m := model{db: d, st: stateList}
  m.loadSurah()
  return m
}

func (m *model) loadSurah() {
  var rows []surahRow
  if err := m.db.Select(&rows, `SELECT number, name_ar, verses_count FROM surah ORDER BY number`); err != nil {
    m.lastErr = err.Error()
  } else {
    m.lastErr = ""
  }
  m.list = rows
  m.status = fmt.Sprintf("Loaded %d surah", len(rows))
}

func (m *model) loadAyah(n int) {
  m.curSurah = n
  var rows []ayahRow
  if err := m.db.Select(&rows, `SELECT number, arabic, tajweed, trans FROM ayah WHERE surah=? ORDER BY number`, n); err != nil {
    m.lastErr = err.Error()
  } else {
    m.lastErr = ""
  }
  m.ayat = rows
  m.status = fmt.Sprintf("Surah %d — %d ayah", n, len(rows))
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  switch msg := msg.(type) {
  case tea.KeyMsg:
    switch msg.String() {
    case "ctrl+c", "q":
      return m, tea.Quit
    }
    switch m.st {
    case stateList:
      switch msg.String() {
      case "up", "k":
        if m.cursor > 0 { m.cursor-- }
        if m.cursor < m.listOff { m.listOff = m.cursor }
      case "down", "j":
        if m.cursor < len(m.list)-1 { m.cursor++ }
        if m.h > 0 {
          maxVis := m.h - 4
          if m.cursor >= m.listOff+maxVis { m.listOff = m.cursor - maxVis + 1 }
        }
      case "enter":
        if len(m.list) > 0 {
          n := m.list[m.cursor].Number
          m.loadAyah(n)
          m.st = stateSurah
          m.ayOff = 0
        }
      }
    case stateSurah:
      switch msg.String() {
      case "b", "esc":
        m.st = stateList
      case "up", "k":
        if m.ayOff > 0 { m.ayOff-- }
      case "down", "j":
        m.ayOff++
      case "/":
        // simple focus: no input mode; leave for future
      }
    }
  case tea.WindowSizeMsg:
    m.w, m.h = msg.Width, msg.Height
  }
  return m, nil
}

func (m model) View() string {
  if m.st == stateList { return m.viewList() }
  return m.viewSurah()
}

func (m model) viewList() string {
  b := &strings.Builder{}
  fmt.Fprintln(b, "Quran TUI — Surah list (↑/↓, Enter, q)")
  fmt.Fprintln(b, strings.Repeat("-", max(10, m.w)))
  if m.lastErr != "" { fmt.Fprintln(b, "Err:", m.lastErr) }
  if m.status != "" { fmt.Fprintln(b, m.status) }
  if len(m.list) == 0 {
    fmt.Fprintln(b, "No surah found. Ensure quran.db exists or run 'make seed'.")
    return b.String()
  }
  start := m.listOff
  end := len(m.list)
  // Determine visible window size; fall back to full list if unknown/small
  maxVis := m.h - 4
  if maxVis <= 0 { maxVis = end }
  if vv := start + maxVis; vv < end { end = vv }
  for i := start; i < end; i++ {
    s := m.list[i]
    cur := "  "
    if i == m.cursor { cur = "> " }
    fmt.Fprintf(b, "%s[%3d] %-32s (%d)\n", cur, s.Number, s.NameAr, s.Verses)
  }
  return b.String()
}

func (m model) viewSurah() string {
  b := &strings.Builder{}
  fmt.Fprintf(b, "Surah %d — (b to back, q to quit)\n", m.curSurah)
  fmt.Fprintln(b, strings.Repeat("-", max(10, m.w)))
  // build lines once per view
  lines := make([]string, 0, len(m.ayat)*2)
  for _, a := range m.ayat {
    lines = append(lines, fmt.Sprintf("%d:%d  %s", m.curSurah, a.Number, a.Arabic))
    if strings.TrimSpace(a.Trans) != "" {
      lines = append(lines, "    "+a.Trans)
    }
  }
  start := clamp(m.ayOff, 0, max(0, len(lines)-1))
  end := len(lines)
  if m.h > 0 { if vv := start + (m.h - 4); vv < end { end = vv } }
  for i := start; i < end; i++ { fmt.Fprintln(b, lines[i]) }
  return b.String()
}

func clamp(v, lo, hi int) int { if v < lo { return lo }; if v > hi { return hi }; return v }
func max(a, b int) int { if a > b { return a }; return b }

func main() {
  ctx := context.Background()
  path := os.Getenv("QURAN_DB_PATH")
  if path == "" { path = "quran.db" }
  d, err := db.Open(path); must(err)
  must(db.Migrate(ctx, d))

  p := tea.NewProgram(initialModel(d))
  if _, err := p.Run(); err != nil { fmt.Println("error:", err) }
}

func must(err error){ if err != nil { panic(err) } }
