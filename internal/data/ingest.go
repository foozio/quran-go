package data

import (
  "context"
  "fmt"
  "strconv"
  "strings"

  "github.com/jmoiron/sqlx"
)

func IngestAll(ctx context.Context, db *sqlx.DB, lang string) error {
  idx, err := FetchSurahIndex(); if err != nil { return err }
  tx := db.MustBegin()
  for _, s := range idx {
    // semarketir/quranjson structure
    // index: "001" (string), titleAr: arabic name, title: latin, place: Mecca/Medina, count: number of verses
    idxStr, _ := s["index"].(string)
    if idxStr == "" { continue }
    n64, _ := strconv.ParseInt(strings.TrimLeft(idxStr, "0"), 10, 32)
    if n64 == 0 { n64 = 0 } // keep as 0 if failed (should not)
    nameAr, _ := s["titleAr"].(string)
    if nameAr == "" { nameAr, _ = s["title"].(string) }
    nameAr = strings.TrimSpace(nameAr)
    nameLa, _ := s["title"].(string)
    place, _ := s["place"].(string)
    cnt := 0
    if v, ok := s["count"].(float64); ok { cnt = int(v) }
    tx.MustExec(`INSERT OR REPLACE INTO surah(number,name_ar,name_latin,revelation,verses_count) VALUES(?,?,?,?,?)`,
      int(n64), nameAr, nameLa, place, cnt)
  }
  if err := tx.Commit(); err != nil { return err }

  for surah := 1; surah <= 114; surah++ {
    ar, err := FetchArabicSurah(surah); if err != nil { return fmt.Errorf("surah %d: %w", surah, err) }
    // tajweed currently ignored (format differs)
    tr, _ := FetchTranslation(lang, surah)

    // Arabic verses live under object: verse: { verse_1: "text", ... }
    verseAr, _ := ar["verse"].(map[string]any)
    verseTr := map[string]any{}
    if v, ok := tr["verse"].(map[string]any); ok { verseTr = v }
    cnt := 0
    if v, ok := ar["count"].(float64); ok { cnt = int(v) }

    tx := db.MustBegin()
    for i := 1; i <= cnt; i++ {
      key := fmt.Sprintf("verse_%d", i)
      arabic, _ := verseAr[key].(string)
      trn, _ := verseTr[key].(string)
      au := AudioURL(surah, i)
      tx.MustExec(`INSERT OR REPLACE INTO ayah(surah,number,juz,arabic,tajweed,trans,audio_url)
        VALUES(?,?,?,?,?,?,?)`, surah, i, 1, strings.TrimSpace(arabic), "", trn, au)
    }
    if err := tx.Commit(); err != nil { return err }
  }
  return nil
}
