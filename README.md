# indexer-jackett

Jackett/Prowlarr indexer connector via Torznab API for MuxCore.

## Env Vars

- `MUXCORE_JACKETT_ADDR` — Jackett/Prowlarr URL (default: `http://localhost:9117`)
- `MUXCORE_JACKETT_APIKEY` — API key

## Usage

```go
import "github.com/Muxcore-Media/indexer-jackett"

mod := jackett.NewModule()
mgr.Register(mod, nil)
```
