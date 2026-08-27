# Performance log

## 2026-08-27

Measurements were repeated under the same local conditions on an Intel Xeon Silver 4210R. The SSH benchmark uses a fixed 20 ms delay per remote command so command-round-trip changes are deterministic.

| Area | Baseline | Result | Verdict |
| --- | --- | --- | --- |
| Container creation, synchronous path | 82.26 ms/op median (82.17–82.31) | 40.83 ms/op median (40.77–40.91) | Kept: 50.4% faster by resolving credentials once, probing ports concurrently, and letting `docker run -v` create the named volume. |
| Initial JS transfer through Nginx | 111,166 B | 43,336 B | Kept: gzip reduced transfer by 61.0%. |
| Lazy-loaded xterm JS transfer through Nginx | 282,768 B | 69,904 B | Kept: gzip reduced transfer by 75.3%. |
| Hashed asset cache policy | No explicit cache header | `public, max-age=31536000, immutable` | Kept: repeat visits can reuse versioned assets. |
| Frontend initial JS bundle | 111.17 kB / 43.41 kB gzip | Unchanged | No code change: existing route splitting is already well below the 200 kB gzip budget. |

Reproduce the backend measurement:

```bash
cd backend
GOROOT=/usr/local/go PATH=/usr/local/go/bin:$PATH \
  go test ./internal/service -run '^$' \
  -bench '^BenchmarkCreateContainerSSHLatency$' -benchtime=5x -count=5
```

Reproduce the frontend bundle baseline:

```bash
cd frontend
npm run build
```

The Nginx transfer measurements used `curl --raw` with `Accept-Encoding: gzip` against the production Nginx image/configuration. Real SSH latency and browser Core Web Vitals depend on deployment conditions; no production RUM data was available for this run.

Verification passed with `go test ./...`, `go test -race ./internal/service`, `go vet ./...`, `npm run build`, `nginx -t`, and `docker compose config --quiet`.
