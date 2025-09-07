package main

import (
  "context"
  "fmt"

  tea "github.com/charmbracelet/bubbletea"
  "github.com/charmbracelet/bubbles/textinput"
  "github.com/jmoiron/sqlx"
  _ "modernc.org/sqlite"
)

type model struct {
  db   *sqlx.DB
  ti   textinput.Model
  rows []struct{ Surah, Number int; Snip string }
  err  error
}

func (m model) Init() tea.Cmd { return textinput.Blink }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  switch msg := msg.(type) {
  case tea.KeyMsg:
    switch msg.String() {
    case "ctrl+c","esc": return m, tea.Quit
    case "enter":
      q := m.ti.Value()
      m.rows = nil
      m.err = m.db.Select(&m.rows, `
        SELECT ayah.surah, ayah.number, snippet(ayah_fts,2,'[',']','…',10) snip
        FROM ayah_fts JOIN ayah ON ayah_fts.rowid = ayah.rowid
        WHERE ayah_fts MATCH ? LIMIT 30`, q)
      return m, nil
    }
  }
  var cmd tea.Cmd
  m.ti, cmd = m.ti.Update(msg)
  return m, cmd
}
func (m model) View() string {
  s := "Quran TUI — search Arabic/translation:\n" + m.ti.View() + "\n\n"
  if m.err != nil { s += fmt.Sprintf("err: %v\n", m.err) }
  for _, r := range m.rows {
    s += fmt.Sprintf("(%d:%d) %s\n", r.Surah, r.Number, r.Snip)
  }
  return s
}
func main(){
  db, _ := sqlx.Open("sqlite","quran.db")
  in := textinput.New(); in.Placeholder="type query, press Enter"
  p := tea.NewProgram(model{db: db, ti: in})
  _, _ = p.RunWithContext(context.Background())
}
