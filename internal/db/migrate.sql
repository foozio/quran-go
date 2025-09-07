PRAGMA foreign_keys=ON;

CREATE TABLE IF NOT EXISTS surah (
  number INTEGER PRIMARY KEY,
  name_ar TEXT NOT NULL,
  name_latin TEXT,
  revelation TEXT,
  verses_count INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS ayah (
  surah INTEGER NOT NULL,
  number INTEGER NOT NULL,
  juz INTEGER NOT NULL,
  arabic TEXT NOT NULL,
  tajweed TEXT,
  trans TEXT,
  audio_url TEXT,
  PRIMARY KEY (surah, number),
  FOREIGN KEY (surah) REFERENCES surah(number) ON DELETE CASCADE
);

-- FTS for search across Arabic & translation
CREATE VIRTUAL TABLE IF NOT EXISTS ayah_fts
USING fts5(surah, number, arabic, trans, content='ayah', content_rowid='rowid');

-- triggers to keep FTS in sync
CREATE TRIGGER IF NOT EXISTS ayah_ai AFTER INSERT ON ayah BEGIN
  INSERT INTO ayah_fts(rowid,surah,number,arabic,trans)
  VALUES (new.rowid, new.surah, new.number, new.arabic, new.trans);
END;
CREATE TRIGGER IF NOT EXISTS ayah_ad AFTER DELETE ON ayah BEGIN
  INSERT INTO ayah_fts(ayah_fts, rowid, surah, number, arabic, trans)
  VALUES ('delete', old.rowid, old.surah, old.number, old.arabic, old.trans);
END;
CREATE TRIGGER IF NOT EXISTS ayah_au AFTER UPDATE ON ayah BEGIN
  INSERT INTO ayah_fts(ayah_fts, rowid, surah, number, arabic, trans)
  VALUES ('delete', old.rowid, old.surah, old.number, old.arabic, old.trans);
  INSERT INTO ayah_fts(rowid,surah,number,arabic,trans)
  VALUES (new.rowid, new.surah, new.number, new.arabic, new.trans);
END;
