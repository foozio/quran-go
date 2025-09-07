package quran

type SurahInfo struct {
    Number       int    `json:"number"`
    NameArabic   string `json:"name"`
    NameLatin    string `json:"name_latin,omitempty"`
    VersesCount  int    `json:"number_of_ayah"`
    Revelation   string `json:"place"`
}

type Ayah struct {
    Surah   int    `json:"surah"`
    Number  int    `json:"number"`
    Arabic  string `json:"arabic"`
    Tajweed string `json:"tajweed,omitempty"`
    Trans   string `json:"translation,omitempty"`
    Juz     int    `json:"juz"`
    Audio   string `json:"audio,omitempty"` // URL
}
