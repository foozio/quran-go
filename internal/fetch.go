package data

import (
  "encoding/json"
  "fmt"
  "io"
  "net/http"
)

const rawBase = "https://raw.githubusercontent.com/semarketir/quranjson/master/source" // repo layout confirmed. :contentReference[oaicite:6]{index=6}

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

// Surah catalog = source/surah.json
func FetchSurahIndex() ([]map[string]any, error) {
  var out []map[string]any
  err := get("surah.json", &out) // path documented. :contentReference[oaicite:7]{index=7}
  return out, err
}

// Arabic text per surah = source/surah/surah_<n>.json
func FetchArabicSurah(n int) (map[string]any, error) {
  var out map[string]any
  err := get(fmt.Sprintf("surah/surah_%d.json", n), &out) // path documented. :contentReference[oaicite:8]{index=8}
  return out, err
}

// Tajweed = source/tajweed/surah_<n>.json
func FetchTajweed(n int) (map[string]any, error) {
  var out map[string]any
  err := get(fmt.Sprintf("tajweed/surah_%d.json", n), &out) // path documented. :contentReference[oaicite:9]{index=9}
  return out, err
}

// Translation per surah = source/translation/<lang>/<lang>_translation_<n>.json
func FetchTranslation(lang string, n int) (map[string]any, error) {
  var out map[string]any
  err := get(fmt.Sprintf("translation/%s/%s_translation_%d.json", lang, lang, n), &out) // path documented. :contentReference[oaicite:10]{index=10}
  return out, err
}

// Audio base per ayah = source/audio/<surah>/<ayah>.mp3 (index.json exists per surah)
// We'll just format URLs; no downloading needed. :contentReference[oaicite:11]{index=11}
func AudioURL(surah, ayah int) string {
  return fmt.Sprintf("%s/audio/%03d/%03d.mp3", rawBase, surah, ayah)
}
