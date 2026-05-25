# 🔗 ei23-FoundAItion – Manual / Anleitung

Kleine, schnelle Web-Oberfläche zum **Verwalten**, **Suchen** und **Teilen** von Links mit KI-Zusammenfassung (OpenAI). Eine einzige Binärdatei, Dark-Mode by default, Datenspeicher in **SQLite** (lokal, kein Server nötig).

---

## Features

- **SQLite-Datenbank** – Lokal, keine externen Dependencies
- **Bilingual UI** – Deutsch und English, umschaltbar über die UI
- Links aus der Tabelle `links` lesen (feste Tabelle, nicht konfigurierbar)
- Kompakte Listenansicht mit Titel, Datum, Status-Dots (Included, Marked, Read)
- Detailansicht mit gerendertem Markdown (Goldmark GFM: Tabellen, Code-Blöcke, Listen)
- **Suchfunktion** mit Echtzeit-Filter über Titel, URL, Zusammenfassung
- **Kategorie-Filter** und **Include/Marked/Read-Filter** als Toggle
- **Status-Dots** in der Liste: Included (grün), Marked (gelb), Read (blau)
- **"Read"-Flag**: Einträge als gelesen/ungelesen markieren, direkt in der Liste
- **Notizen editieren** in der Detailansicht
- **Link löschen** mit Bestätigungsdialog
- YouTube-Videos erkennen + Thumbnail anzeigen
- YouTube-Playlists: Alle Videos automatisch einzeln hinzufügen
- **KI-Zusammenfassung** (OpenAI / kompatible API) – automatisch beim Hinzufügen oder per Batch
- **RSS/Atom-Feed** (`/rss`) mit Action-Links (Mark, Include, Delete Summary)
- **Whisper-Integration** für YouTube-Transkription
- **Filterwörter** im Notiz-Feld: `kapitel`/`chapters` → Kapitel-Übersicht, `skip`/`überspringen` → Verarbeitung überspringen

---

## Android: Links per HTTP Shortcuts teilen

Mit der kostenlosen App **[HTTP Shortcuts](https://github.com/Waboodoo/HTTP-Shortcuts)** (sehr empfehlenswert!) kannst du Links direkt aus dem Android-Teilen-Menü an FoundAItion senden:

1. **HTTP Shortcuts** installieren und einen neuen Shortcut anlegen
2. Methode: `POST` · URL: `http://<deine-instanz>:8080/linkshare`
3. **Content-Type**: `application/json`
4. **Body** (JSON):
   ```json
   {"url": "{{param:clipboard|url}}", "note": "geteilt"}
   ```
5. Im Shortcut **„Teilen“ aktivieren**, damit er im Android-Teilen-Dialog erscheint
6. Beliebig oft erweitern – z. B. mit einem zweiten Shortcut für YouTube-Playlists:
   ```json
   {"url": "{{param:clipboard|url}}", "note": "kapitel"}
   ```
   (dank des Filterworts `kapitel` werden Kapitelüberschriften erstellt)

So landen Artikel, Videos oder ganze Playlists mit einem Klick in FoundAItion und werden automatisch zusammengefasst.

---

## Anforderungen

| Was | Warum |
|-----|-------|
| **Go ≥ 1.21** | Kompilieren der Binary |
| **SQLite** | Datenbank (lokal, pure Go via modernc.org/sqlite, no CGO) |

### Summarizer (optional)
| Was | Warum |
|-----|-------|
| **OpenAI API Key** | KI-Zusammenfassungen (GPT-4o-mini oder kompatibel) |
| **yt-dlp** | YouTube-Audio herunterladen |
| **Whisper Server** | Lokaler Self-hosted Whisper für Audio-Transkription |
| **Colly + html-to-markdown** | Reines Go – Webseiten crawlen, HTML → Markdown |

Die Tabelle `links` wird beim ersten Start **automatisch angelegt** (`CREATE TABLE IF NOT EXISTS`).

### Tabellenschema (15 Spalten)

| Column | Type | Default | Description |
|--------|------|---------|-------------|
| `id` | INTEGER AUTOINCREMENT | – | Primary key |
| `timestamp` | TEXT | – | Creation timestamp (ISO-8601) |
| `url` | TEXT | – | The link URL |
| `title` | TEXT | – | Page/video title |
| `note` | TEXT | – | Notes + filter keywords (`chapters`, `skip`) |
| `summary` | TEXT | – | AI-generated summary (Markdown) |
| `content` | TEXT | – | Crawled content / transcript |
| `category` | TEXT | – | Category for filtering |
| `included` | INTEGER | `0` | Include in export/view |
| `marked` | INTEGER | `0` | Marked flag |
| `read` | INTEGER | `0` | Read/Unread flag |
---

## Schnellstart

```bash
# .env erstellen:
cat > .env << 'EOF'
LISTEN_PORT=8080
OPENAI_API_KEY=sk-...
YTDLP_PATH=/usr/local/bin/yt-dlp
EOF

go mod tidy && go build -o foundaition
./foundaition
# SQLite-Datei: ./foundaition.db (auto-created)
```

Dann `http://localhost:8080` im Browser öffnen. Port über `LISTEN_PORT=9000` änderbar.

---

## Bedienung

### Listenansicht (Startseite)

- **„+ Link hinzufügen"**: Ausklappbares Formular zum Einfügen neuer URLs
- **„⚙️ Prozess starten"**: Batch-Verarbeitung aller Einträge ohne Zusammenfassung
- Nummerierte Liste aller Links, neuester zuerst (paginiert: 20/Seite)
- **YouTube-Links** zeigen ein Thumbnail
- Status-Dots: Included (grün), Marked (gelb), Read (blau) – klickbar
- Kategorie-Badges rechts
- Filter-Toggles für Included/Marked/Read (An/Aus)
- **Suchfeld oben**: Echtzeit-Suche über Titel, URL, Zusammenfassung

### Detailansicht (`#link/123`)

- Titel + Metadaten (URL, Datum, Notiz)
- **Markdown-Zusammenfassung** gerendert via Goldmark (GFM)
- **Editierbereich**: Notiz bearbeiten, Included/Marked/Read togglen
- **Aktionen:**
  - 📋 URL kopieren
  - 📝 Zusammenfassung als Markdown kopieren
  - ↗️ Im Browser öffnen
- 🗑️ **Eintrag löschen** (mit Bestätigungsdialog)

### Sprachumschaltung

- Im Burger-Menü (oben links) → Sprache wählen: **Deutsch** oder **English**
- Die UI wechselt sofort alle Texte
- Prompt-Templates werden automatisch in der passenden Sprache geladen
- Einstellung wird in `.env` gespeichert (`UI_LANGUAGE=de` / `UI_LANGUAGE=en`)

---

## Umgebungsvariablen

| Variable | Default | Description |
|----------|---------|-------------|
| `LISTEN_PORT` | `8080` | HTTP server port |
| `DB_FILE` | `./foundaition.db` | SQLite database file path |
| `UI_LANGUAGE` | `de` | Interface language (`de` or `en`) |
| `OPENAI_API_KEY` | *(empty)* | OpenAI API key |
| `OPENAI_MODEL` | `gpt-4o-mini` | AI model for summaries |
| `OPENAI_MODELS` | `gpt-4o-mini,gpt-4o` | Comma-separated available models |
| `OPENAI_BASE_URL` | `https://api.openai.com/v1` | OpenAI-compatible API endpoint |
| `SUMMARY_MAX_TOKENS` | `64000` | Max tokens for summary content (×3.5 chars/token) |
| `AUTO_SUMMARIZE` | `true` | Auto-generate summary on link add |
| `WHISPER_URL` | `http://localhost:8081` | Local Whisper server URL |
| `WHISPER_MODEL` | `large-v3` | Whisper model |
| `CRAWL_TIMEOUT` | `60` | Timeout per crawl request (seconds) |
| `YTDLP_PATH` | auto-detect | Path to yt-dlp binary |
| `MAX_PROCESS_DAYS` | `14` | Max days backwards for batch processing |
| `RSS_BASE_URL` | *(empty)* | External base URL for action links in Atom feed |
| `RSS_ITEM_COUNT` | `30` | Max entries in Atom feed |

---

## API-Endpunkte

| Methode | Pfad | Beschreibung |
|---------|------|-------------|
| GET | `/` | Index page (HTML) |
| GET | `/api/links` | Links auflisten – Filter: `?page=N`, `?q=...` (Titel/URL/Summary/Kategorie), `?url_like=...` (nur URL), `?category=...`, `?note=...`, `?included=true`, `?marked=true`, `?read=true/false` · Einzellink: `?id=123` |
| PATCH | `/api/links?id=X` | Felder aktualisieren (`?note=...&included=true/false&marked=true/false&read=true/false`) |
| DELETE | `/api/links?id=X` | Delete link |
| GET | `/api/categories` | List categories with counts |
| GET/POST | `/api/share` | Get link(s) by ID |
| POST | `/linkshare` | Add link (`{"url":"...","note":"..."}`) |
| GET | `/process-entries` | Process pending entries (batch) |
| GET | `/api/manual` | Get this manual (raw Markdown) |
| GET | `/api/config` | Get configuration |
| POST | `/api/config` | Save configuration |
| GET/POST | `/api/language` | Get/set UI language (`{"language":"de"}`) |
| GET | `/rss` | Atom 1.0 feed (`?page=N&q=...&url_like=...&category=...&note=...&included=true&marked=true&read=true/false`) |
| GET | `/api/feed/mark?id=X` | Set marked=1 |
| GET | `/api/feed/include?id=X` | Set included=1 |
| GET | `/api/feed/delete-summary?id=X` | Clear summary + content |

---

## Prompts

Die Prompt-Templates liegen im `prompts/`-Verzeichnis und werden sprachspezifisch geladen:

```
prompts/
├── summary_DE.md        # Zusammenfassung (Deutsch)
├── summary_EN.md        # Summary (English)
├── chapters_DE.md       # Kapitel-Übersicht (Deutsch)
└── chapters_EN.md       # Chapter overview (English)
```

Die App wählt das Template basierend auf der aktuellen `UI_LANGUAGE` aus.

---

## Projektstruktur

```
├── main.go              # Entry point: DB, API, HTTP server
├── app.go               # App struct, prompt loading, renderPrompt
├── config.go            # Configuration loading + token estimation
├── configapi.go         # Config API schema + handlers
├── handlers.go          # HTTP handlers (CRUD, RSS, Feed, Language)
├── types.go             # Data structures, scanLink, NullBool
├── crawl.go             # Web crawling + markdown conversion
├── openai.go            # OpenAI API + Whisper integration
├── youtube.go           # YouTube subtitle extraction
├── db.go                # SQLite connection + table creation
├── go.mod / go.sum      # Dependencies
├── prompts/
│   ├── summary_DE.md
│   ├── summary_EN.md
│   ├── chapters_DE.md
│   └── chapters_EN.md
├── templates/
│   └── index.html       # Frontend: HTML/CSS/JS (bilingual)
├── .env                 # Configuration (auto-created)
├── .gitignore
├── manual.md            # ← this file
├── README.md            # GitHub project description
└── agent.md             # For AI assistants
```

---

## Lizenz

Intern / privat.
