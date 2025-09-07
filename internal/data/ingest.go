package data

import (
  "context"
  "fmt"
  "strings"

  "github.com/jmoiron/sqlx"
)

func IngestAll(ctx context.Context, db *sqlx.DB, lang string) error {
  idx, err := FetchSurahIndex(); if err != nil { return err }
  tx := db.MustBegin()
  for _, s := range idx {
    n := int(s["number"].(float64))
    nameAr := strings.TrimSpace(s["name"].(string))
    nameLa, _ := s["name_latin"].(string)
    place, _ := s["place"].(string)
    cnt := int(s["number_of_ayah"].(float64))
    tx.MustExec(`INSERT OR REPLACE INTO surah(number,name_ar,name_latin,revelation,verses_count) VALUES(?,?,?,?,?)`,
      n, nameAr, nameLa, place, cnt)
  }
  if err := tx.Commit(); err != nil { return err }

  for surah := 1; surah <= 114; surah++ {
    ar, err := FetchArabicSurah(surah); if err != nil { return fmt.Errorf("surah %d: %w", surah, err) }
    taj, _ := FetchTajweed(surah)
    tr, _ := FetchTranslation(lang, surah)

    ayatAr := ar["verses"].([]any)
    var ayatTaj, ayatTr []any
    if v, ok := taj["verses"].([]any); ok { ayatTaj = v }
    if v, ok := tr["verses"].([]any); ok { ayatTr = v }

    tx := db.MustBegin()
    for i, v := range ayatAr {
      m := v.(map[string]any)
      ayNum := int(m["number"].(float64))
      arabic := strings.TrimSpace(m["text"].(string))
      var tj, trn string
      if i < len(ayatTaj) {
        if mm, ok := ayatTaj[i].(map[string]any); ok { tj, _ = mm["text"].(string) }
      }
      if i < len(ayatTr) {
        if mm, ok := ayatTr[i].(map[string]any); ok { trn, _ = mm["text"].(string) }
      }
      au := AudioURL(surah, ayNum)
      tx.MustExec(`INSERT OR REPLACE INTO ayah(surah,number,juz,arabic,tajweed,trans,audio_url)
        VALUES(?,?,?,?,?,?,?)`, surah, ayNum, 1, arabic, tj, trn, au)
    }
    if err := tx.Commit(); err != nil { return err }
  }
  return nil
}
