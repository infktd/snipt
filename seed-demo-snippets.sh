#!/usr/bin/env bash
# seed-demo-snippets.sh
# Seeds snipt with demo snippets for VHS recordings
#
# Usage:
#   bash seed-demo-snippets.sh          # clear + seed
#   bash seed-demo-snippets.sh clean    # just clear everything
#   bash seed-demo-snippets.sh seed     # just seed (without clearing)

set -euo pipefail

DB="$HOME/.local/share/snipt/snipt.db"

if [ ! -f "$DB" ]; then
  echo "Error: snipt database not found at $DB"
  echo "Run 'snipt' once first to initialize the database."
  exit 1
fi

ACTION="${1:-all}"

clean_db() {
  echo "🧹 Clearing all snippets..."
  sqlite3 "$DB" "DELETE FROM tags; DELETE FROM snippets;"
  echo "   Database is squeaky clean. Fresh start."
}

seed_db() {
  echo "🔪 Seeding snipt with demo snippets..."
  echo "   (a painless procedure, we promise)"

sqlite3 "$DB" <<'SEED'

-- ============================================================
-- Go snippets - the bread and butter
-- ============================================================

INSERT INTO snippets (id, title, content, language, description, source, pinned, use_count, created_at, updated_at) VALUES
('http-middleware-stack', 'HTTP server with middleware',
'func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("snip snip, still kicking"))
    })

    // wrap it up -- protection is important
    wrapped := corsMiddleware(
        loggingMiddleware(
            panicRecovery(mux),
        ),
    )

    log.Println("listening on :8080, no turning back now")
    http.ListenAndServe(":8080", wrapped)
}',
'go', 'HTTP server with protective middleware layers', '', 1, 47,
datetime('now'), datetime('now'));

INSERT INTO snippets (id, title, content, language, description, source, pinned, use_count, created_at, updated_at) VALUES
('retry-backoff', 'Retry with exponential backoff',
'// sometimes you gotta try more than once
// commitment issues? no -- resilience
func retry(attempts int, fn func() error) error {
    for i := 0; i < attempts; i++ {
        if err := fn(); err == nil {
            return nil // clean cut, first try
        }
        backoff := time.Duration(1<<i) * time.Second
        log.Printf("attempt %d failed, icing it for %v", i+1, backoff)
        time.Sleep(backoff)
    }
    return fmt.Errorf("failed after %d attempts -- should have quit sooner", attempts)
}',
'go', 'Exponential backoff -- keep trying until it takes', '', 0, 23,
datetime('now'), datetime('now'));

INSERT INTO snippets (id, title, content, language, description, source, pinned, use_count, created_at, updated_at) VALUES
('graceful-shutdown', 'Graceful shutdown handler',
'// the procedure: clean, quick, no drama
func gracefulShutdown(srv *http.Server) {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    <-quit // the moment of truth
    log.Println("shutting down gracefully... brief discomfort")

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("forced shutdown -- complications:", err)
    }
    log.Println("done. rest up. frozen peas recommended.")
}',
'go', 'A quick, painless shutdown procedure', '', 1, 31,
datetime('now'), datetime('now'));

INSERT INTO snippets (id, title, content, language, description, source, pinned, use_count, created_at, updated_at) VALUES
('error-wrap-pattern', 'Error wrapping pattern',
'func fetchUser(id string) (*User, error) {
    row := db.QueryRow("SELECT * FROM users WHERE id = ?", id)

    var u User
    if err := row.Scan(&u.ID, &u.Name, &u.Email); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            // snipped from the record entirely
            return nil, fmt.Errorf("user %s: cut from the lineup: %w", id, err)
        }
        return nil, fmt.Errorf("user %s: something went wrong on the table: %w", id, err)
    }
    return &u, nil // all clear, no complications
}',
'go', 'Wrap errors properly -- no loose ends', '', 0, 19,
datetime('now'), datetime('now'));

INSERT INTO snippets (id, title, content, language, description, source, pinned, use_count, created_at, updated_at) VALUES
('bubbletea-model', 'Bubbletea TUI boilerplate',
'// the little snip that could
type model struct {
    list     list.Model
    choice   string
    quitting bool
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyPressMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            m.quitting = true
            return m, tea.Quit // snip, done
        case "enter":
            i, ok := m.list.SelectedItem().(item)
            if ok {
                m.choice = string(i)
            }
            return m, tea.Quit
        }
    }
    var cmd tea.Cmd
    m.list, cmd = m.list.Update(msg)
    return m, cmd
}

func (m model) View() string {
    return m.list.View()
}',
'go', 'Bubbletea model -- the TUI that keeps on giving', '', 1, 42,
datetime('now'), datetime('now'));

INSERT INTO snippets (id, title, content, language, description, source, pinned, use_count, created_at, updated_at) VALUES
('concurrent-pipeline', 'Fan-out worker pipeline',
'// spawn workers, get results, no surprises
// like a well-run clinic -- in and out
func process(ctx context.Context, items []string, workers int) []Result {
    var wg sync.WaitGroup
    jobs := make(chan string, len(items))
    results := make(chan Result, len(items))

    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for item := range jobs {
                select {
                case <-ctx.Done():
                    return // procedure cancelled
                default:
                    results <- doWork(item)
                }
            }
        }()
    }

    for _, item := range items {
        jobs <- item
    }
    close(jobs) // no more patients today

    go func() { wg.Wait(); close(results) }()

    var out []Result
    for r := range results {
        out = append(out, r)
    }
    return out
}',
'go', 'Bounded concurrency -- controlled, precise, no loose threads', '', 0, 16,
datetime('now'), datetime('now'));

INSERT INTO snippets (id, title, content, language, description, source, pinned, use_count, created_at, updated_at) VALUES
('channel-semaphore', 'Channel-based semaphore',
'// limit concurrency -- because uncontrolled
// reproduction of goroutines is irresponsible
type Semaphore struct {
    ch chan struct{}
}

func NewSemaphore(max int) *Semaphore {
    return &Semaphore{ch: make(chan struct{}, max)}
}

func (s *Semaphore) Acquire() { s.ch <- struct{}{} }
func (s *Semaphore) Release() { <-s.ch }

// Usage: prevent goroutine proliferation
sem := NewSemaphore(10)
for _, task := range tasks {
    sem.Acquire()
    go func(t Task) {
        defer sem.Release()
        process(t)
    }(task)
}',
'go', 'Prevent uncontrolled goroutine reproduction', '', 0, 7,
datetime('now'), datetime('now'));

-- ============================================================
-- Nix snippets - declarative and reproducible, like a good doc
-- ============================================================

INSERT INTO snippets (id, title, content, language, description, source, pinned, use_count, created_at, updated_at) VALUES
('nix-flake-devshell', 'Nix flake dev shell',
'{
  description = "a sterile development environment";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { nixpkgs, ... }: let
    pkgs = nixpkgs.legacyPackages.x86_64-linux;
  in {
    devShells.x86_64-linux.default = pkgs.mkShell {
      packages = with pkgs; [
        go gopls gotools
        sqlite
        nodejs
      ];
    };
  };
}',
'nix', 'Clean, reproducible dev environment -- no surprises', '', 1, 31,
datetime('now'), datetime('now'));

INSERT INTO snippets (id, title, content, language, description, source, pinned, use_count, created_at, updated_at) VALUES
('nix-home-manager-git', 'Home Manager git config',
'{ pkgs, ... }: {
  programs.git = {
    enable = true;
    userName = "Jay";
    userEmail = "jay@infktd.dev";
    extraConfig = {
      init.defaultBranch = "main";
      push.autoSetupRemote = true;  # no pulling out mid-push
      pull.rebase = true;
      core.editor = "nvim";
      rerere.enabled = true;  # remember past resolutions
    };
    delta = {
      enable = true;  # see the diff clearly before you commit
      options.navigate = true;
    };
  };
}',
'nix', 'Git config -- commit with confidence', '', 0, 18,
datetime('now'), datetime('now'));

-- ============================================================
-- Bash - quick scripts, in and out
-- ============================================================

INSERT INTO snippets (id, title, content, language, description, source, pinned, use_count, created_at, updated_at) VALUES
('git-precommit-hook', 'Git pre-commit hook',
'#!/usr/bin/env bash
set -euo pipefail

echo "Pre-commit checks... quick consultation before you commit"

if ! go vet ./...; then
    echo "go vet: found issues. not ready to commit."
    exit 1
fi

if ! staticcheck ./...; then
    echo "staticcheck: nope, go home."
    exit 1
fi

if ! go test ./... -count=1 -short; then
    echo "tests: failed. no commitment today."
    exit 1
fi

echo "All clear. you may commit. its permanent though."',
'bash', 'Think twice before you commit -- checks first', '', 0, 15,
datetime('now'), datetime('now'));

INSERT INTO snippets (id, title, content, language, description, source, pinned, use_count, created_at, updated_at) VALUES
('find-and-process', 'Recursive find and process',
'#!/usr/bin/env bash
# snip through your codebase
find . -name "*.go" -type f | while read -r file; do
    echo "Examining: $file"
    lines=$(wc -l < "$file")
    todos=$(grep -c "TODO" "$file" || true)
    printf "  %d lines, %d unfinished procedures\n" "$lines" "$todos"
done
echo "Consultation complete."',
'bash', 'Quick examination of the whole codebase', '', 0, 28,
datetime('now'), datetime('now'));

-- ============================================================
-- SQL - precise cuts
-- ============================================================

INSERT INTO snippets (id, title, content, language, description, source, pinned, use_count, created_at, updated_at) VALUES
('fts5-search-setup', 'SQLite FTS5 search',
'-- full-text search: find what you snipped
CREATE VIRTUAL TABLE IF NOT EXISTS snippets_fts USING fts5(
    title, content, description, language,
    content=''snippets'',
    content_rowid=''rowid''
);

-- rebuild (like a follow-up appointment)
INSERT INTO snippets_fts(snippets_fts) VALUES(''rebuild'');

-- locate the snippet with surgical precision
SELECT s.*, rank
FROM snippets s
JOIN snippets_fts fts ON s.rowid = fts.rowid
WHERE snippets_fts MATCH ''http AND server''
ORDER BY rank
LIMIT 20;',
'sql', 'FTS5 search -- locate anything with surgical precision', '', 1, 12,
datetime('now'), datetime('now'));

INSERT INTO snippets (id, title, content, language, description, source, pinned, use_count, created_at, updated_at) VALUES
('upsert-pattern', 'SQLite upsert pattern',
'-- insert or update. either way, its going in
INSERT INTO snippets (id, title, language, content, updated_at)
VALUES (?, ?, ?, ?, datetime(''now''))
ON CONFLICT(id) DO UPDATE SET
    title = excluded.title,
    language = excluded.language,
    content = excluded.content,
    updated_at = datetime(''now'');
-- one way or another, the job gets done',
'sql', 'Upsert -- the procedure always completes', '', 0, 8,
datetime('now'), datetime('now'));

-- ============================================================
-- TypeScript - type-safe snippage
-- ============================================================

INSERT INTO snippets (id, title, content, language, description, source, pinned, use_count, created_at, updated_at) VALUES
('ts-fetch-wrapper', 'Type-safe fetch wrapper',
'// a protective wrapper. always use protection.
async function api<T>(
  endpoint: string,
  options?: RequestInit
): Promise<T> {
  const res = await fetch(`${BASE_URL}${endpoint}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
  });

  if (!res.ok) {
    const error = await res.text().catch(() => "complications arose");
    throw new Error(`${res.status}: ${error}`);
  }

  return res.json() as Promise<T>;
}

// wrap it before you ship it
// const user = await api<User>("/users/me");',
'typescript', 'Always wrap your requests -- be safe out there', '', 0, 14,
datetime('now'), datetime('now'));

INSERT INTO snippets (id, title, content, language, description, source, pinned, use_count, created_at, updated_at) VALUES
('ts-debounce', 'Debounce with generics',
'// dont fire too often. show some restraint.
function debounce<T extends (...args: any[]) => any>(
  fn: T,
  ms: number
): (...args: Parameters<T>) => void {
  let timer: ReturnType<typeof setTimeout>;

  return (...args: Parameters<T>) => {
    clearTimeout(timer); // cancel the last one
    timer = setTimeout(() => fn(...args), ms);
  };
}

// one clean execution, no repeats
const search = debounce((query: string) => {
  fetch(`/api/search?q=${query}`);
}, 300);',
'typescript', 'One clean execution -- no unnecessary repeats', '', 0, 11,
datetime('now'), datetime('now'));

-- ============================================================
-- YAML & Config - keeping things in order
-- ============================================================

INSERT INTO snippets (id, title, content, language, description, source, pinned, use_count, created_at, updated_at) VALUES
('docker-compose-dev', 'Docker Compose dev stack',
'# local dev: isolated, contained, no cross-contamination
services:
  app:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - .:/app
    environment:
      - DATABASE_URL=postgres://dev:dev@db:5432/app
      - REDIS_URL=redis://cache:6379
    depends_on:
      db:
        condition: service_healthy  # health check first

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: dev
      POSTGRES_PASSWORD: dev
      POSTGRES_DB: app
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U dev"]
      interval: 2s
      retries: 10  # patience is a virtue

  cache:
    image: redis:7-alpine',
'yaml', 'Dev stack -- isolated containers, no complications', '', 0, 22,
datetime('now'), datetime('now'));

INSERT INTO snippets (id, title, content, language, description, source, pinned, use_count, created_at, updated_at) VALUES
('niri-window-rules', 'Niri window rules',
'window-rules {
    // keep everything tidy and contained
    rule {
        match is-floating=true
        opacity 0.95
        shadow enable=true size=12
    }

    rule {
        match app-id="foot|kitty|alacritty"
        default-column-width { proportion 0.5; }
    }

    // browser gets its own space, no overlap
    rule {
        match app-id="firefox|chromium"
        open-on-workspace 2
        default-column-width { proportion 0.75; }
    }
}',
'kdl', 'Window management -- everything in its own space', '', 0, 9,
datetime('now'), datetime('now'));

-- ============================================================
-- Tags - categorize everything
-- ============================================================

INSERT INTO tags (snippet_id, tag) VALUES
('http-middleware-stack', 'http'),
('http-middleware-stack', 'server'),
('http-middleware-stack', 'middleware'),
('retry-backoff', 'resilience'),
('retry-backoff', 'retry'),
('retry-backoff', 'patterns'),
('graceful-shutdown', 'server'),
('graceful-shutdown', 'signals'),
('graceful-shutdown', 'lifecycle'),
('error-wrap-pattern', 'errors'),
('error-wrap-pattern', 'database'),
('error-wrap-pattern', 'patterns'),
('bubbletea-model', 'tui'),
('bubbletea-model', 'bubbletea'),
('bubbletea-model', 'boilerplate'),
('concurrent-pipeline', 'concurrency'),
('concurrent-pipeline', 'goroutines'),
('concurrent-pipeline', 'pipeline'),
('channel-semaphore', 'concurrency'),
('channel-semaphore', 'goroutines'),
('channel-semaphore', 'semaphore'),
('nix-flake-devshell', 'nix'),
('nix-flake-devshell', 'flake'),
('nix-flake-devshell', 'devshell'),
('nix-home-manager-git', 'nix'),
('nix-home-manager-git', 'home-manager'),
('nix-home-manager-git', 'git'),
('git-precommit-hook', 'git'),
('git-precommit-hook', 'hooks'),
('git-precommit-hook', 'lint'),
('git-precommit-hook', 'ci'),
('find-and-process', 'bash'),
('find-and-process', 'find'),
('find-and-process', 'scripting'),
('fts5-search-setup', 'sql'),
('fts5-search-setup', 'sqlite'),
('fts5-search-setup', 'fts5'),
('fts5-search-setup', 'search'),
('upsert-pattern', 'sql'),
('upsert-pattern', 'sqlite'),
('upsert-pattern', 'upsert'),
('ts-fetch-wrapper', 'typescript'),
('ts-fetch-wrapper', 'fetch'),
('ts-fetch-wrapper', 'api'),
('ts-debounce', 'typescript'),
('ts-debounce', 'utils'),
('ts-debounce', 'performance'),
('docker-compose-dev', 'docker'),
('docker-compose-dev', 'compose'),
('docker-compose-dev', 'postgres'),
('docker-compose-dev', 'redis'),
('niri-window-rules', 'niri'),
('niri-window-rules', 'window-manager'),
('niri-window-rules', 'config');

SEED

  echo ""
  echo "✅ Seeded $(sqlite3 "$DB" "SELECT COUNT(*) FROM snippets;") snippets into snipt"
  echo ""
  echo "Snippets by language:"
  sqlite3 "$DB" "SELECT language, COUNT(*) as count FROM snippets GROUP BY language ORDER BY count DESC;" | while IFS='|' read -r lang count; do
    echo "   $lang: $count"
  done
  echo ""
  echo "Tags: $(sqlite3 "$DB" "SELECT COUNT(DISTINCT tag) FROM tags;") unique"
  echo ""
  echo "A painless procedure. Run 'snipt' to see them!"
}

case "$ACTION" in
  clean)
    clean_db
    echo "Done. Everything has been removed. Irreversible."
    ;;
  seed)
    seed_db
    ;;
  all|*)
    clean_db
    seed_db
    ;;
esac
