# ei23-FoundAItion – Manual

A small, fast web interface for **managing**, **searching** and **sharing** links with AI summaries (OpenAI). Single binary, dark mode by default, **SQLite** storage (local, no server needed).

---

## Features

- **SQLite database** – Local, no external dependencies
- **Bilingual UI** – German and English, switchable via the UI
- Reads from the `links` table (fixed, not configurable)
- Compact list view with title, date, status dots (Included, Marked, Read)
- Detail view with rendered Markdown (Goldmark GFM: tables, code blocks, lists)
- **Search** with real-time filter across title, URL, summary
- **Category filter** and **Include/Marked/Read toggles**
- **Status dots** in list view: Included (green), Marked (amber), Read (blue)
- **"Read" flag**: Mark entries as read/unread directly in the list
- **Edit notes** in the detail view
- **Delete entries** with confirmation dialog
- YouTube video detection + thumbnail display
- YouTube playlists: All videos added individually
- **AI summary** (OpenAI / compatible API) – automatic on add or via batch
- **RSS/Atom feed** (`/rss`) with action links (Mark, Include, Delete Summary)
- **Whisper integration** for YouTube transcription
- **Filter words** in the note field: `chapters` → chapter overview, `skip` → skip processing

---

## Requirements

| What | Why |
|-----|-------|
| **Go ≥ 1.21** | Compile the binary |
| **SQLite** | Database (local, pure Go via modernc.org/sqlite, no CGO) |

### Summarizer (optional)
| What | Why |
|-----|-------|
| **OpenAI API Key** | AI summaries (GPT-4o-mini or compatible) |
| **yt-dlp** | YouTube audio download |
| **Whisper Server** | Local self-hosted Whisper for audio transcription |
| **Colly + html-to-markdown** | Pure Go – web crawling, HTML → Markdown |

The `links` table is **auto-created** on first run (`CREATE TABLE IF NOT EXISTS`).

### Schema (15 columns)

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
| `created_at` | TEXT | – | Auto-generated timestamp |
| `updated_at` | TEXT | – | Auto-generated timestamp |
| `created_by` | TEXT | – | Creator ID (reserved) |
| `updated_by` | TEXT | – | Editor ID (reserved) |

---

## Quick Start

```bash
# Create .env:
cat > .env << 'EOF'
LISTEN_PORT=8080
OPENAI_API_KEY=sk-...
YTDLP_PATH=/usr/local/bin/yt-dlp
EOF

go mod tidy && go build -o foundaition
./foundaition
# SQLite file: ./foundaition.db (auto-created)
```

Open `http://localhost:8080` in your browser. Change port via `LISTEN_PORT=9000`.

---

## Usage

### List view (start page)

- **"+ Add Link"**: Expandable form to insert new URLs
- **"⚙️ Process"**: Batch process all entries without summary
- Numbered list of all links, newest first (paginated: 20/page)
- **YouTube links** show a thumbnail
- Status dots: Included (green), Marked (amber), Read (blue) – clickable
- Category badges on the right
- Toggle filters for Included/Marked/Read (On/Off)
- **Search field**: Real-time search across title, URL, summary

### Detail view (`#link/123`)

- Title + metadata (URL, date, note)
- **Markdown summary** rendered via Goldmark (GFM)
- **Edit area**: Edit note, toggle Included/Marked/Read
- **Actions:**
  - 📋 Copy URL
  - 📝 Copy summary as Markdown
  - ↗️ Open in browser
- 🗑️ **Delete entry** (with confirmation dialog)

### Language switching

- In the burger menu (top left) → select **Deutsch** or **English**
- The UI switches all texts immediately
- Prompt templates are loaded in the matching language
- Setting is saved to `.env` (`UI_LANGUAGE=de` / `UI_LANGUAGE=en`)

---

## Environment Variables

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

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Index page (HTML) |
| GET | `/api/links` | List links (`?page=N&q=...&category=...&included=true&marked=true&read=true/false`) |
| PATCH | `/api/links?id=X` | Update fields (`?note=...&included=true/false&marked=true/false&read=true/false`) |
| DELETE | `/api/links?id=X` | Delete link |
| GET | `/api/categories` | List categories with counts |
| GET/POST | `/api/share` | Get link(s) by ID |
| POST | `/linkshare` | Add link (`{"url":"...","note":"..."}`) |
| GET | `/process-entries` | Process pending entries (batch) |
| GET | `/api/manual` | Get this manual (raw Markdown, `?lang=de` or `?lang=en`) |
| GET | `/api/config` | Get configuration |
| POST | `/api/config` | Save configuration |
| GET/POST | `/api/language` | Get/set UI language (`{"language":"de"}`) |
| GET | `/rss` | Atom 1.0 feed |
| GET | `/api/feed/mark?id=X` | Set marked=1 |
| GET | `/api/feed/include?id=X` | Set included=1 |
| GET | `/api/feed/delete-summary?id=X` | Clear summary + content |

---

## Prompts

Prompt templates are stored in the `prompts/` directory and loaded by language:

```
prompts/
├── summary_DE.md        # Summary (German)
├── summary_EN.md        # Summary (English)
├── chapters_DE.md       # Chapters (German)
└── chapters_EN.md       # Chapters (English)
```

The app selects the template based on the current `UI_LANGUAGE`.

---

## Project Structure

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
├── manual_DE.md         # Manual (German)
├── manual_EN.md         # Manual (English)
├── README.md            # GitHub project description
└── agent.md             # For AI assistants
```

---

## License

Internal / private. No claim.
