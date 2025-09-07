package data

import (
  "encoding/json"
  "fmt"
  "io"
  "net/http"
)

const rawBase = "https://raw.githubusercontent.com/semarketir/quranjson/master/source"

func get(path string, v any) error {
  url := fmt.Sprintf("%s/%s", rawBase, path)
  resp, err := http.Get(url)
  if err != nil { return err }
  defer resp.Body.Close()
  if resp.StatusCode != 200 {
    b, _ := io.ReadAll(resp.Body)
    return fmt.Errorf("GET %s: %s (%s)", url, resp.Status, string(b))
  }
  return json.NewDecoder(resp.Body).Decode(v)
}

func FetchSurahIndex() ([]map[string]any, error) {
  var out []map[string]any
  err := get("surah.json", &out)
  return out, err
}

func FetchArabicSurah(n int) (map[string]any, error) {
  var out map[string]any
  err := get(fmt.Sprintf("surah/surah_%d.json", n), &out)
  return out, err
}

func FetchTajweed(n int) (map[string]any, error) {
  var out map[string]any
  err := get(fmt.Sprintf("tajweed/surah_%d.json", n), &out)
  return out, err
}

func FetchTranslation(lang string, n int) (map[string]any, error) {
  var out map[string]any
  err := get(fmt.Sprintf("translation/%s/%s_translation_%d.json", lang, lang, n), &out)
  return out, err
}

func AudioURL(surah, ayah int) string {
  return fmt.Sprintf("%s/audio/%03d/%03d.mp3", rawBase, surah, ayah)
}
