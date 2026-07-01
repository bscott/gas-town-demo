# Gas Town Demo

**SlackLite** — a tiny Slack-style chat backend in Go, used as a demo project.

It exposes a small REST API for channels and messages plus a WebSocket
endpoint for live message broadcast. Messages and channels are kept **in
memory**, so state resets every time the server restarts.

## Requirements

- Go 1.24+

## Run

```bash
go run .
```

The server listens on **http://localhost:8080**:

```
SlackLite server starting on :8080
```

## API

### Channels

| Method | Path                 | Body                 | Description             |
|--------|----------------------|----------------------|-------------------------|
| GET    | `/api/channels`      | –                    | List all channels       |
| POST   | `/api/channels`      | `{"name":"general"}` | Create a channel        |
| GET    | `/api/channels/{id}` | –                    | Get a single channel    |

### Messages

| Method | Path                          | Body                                | Description                 |
|--------|-------------------------------|-------------------------------------|-----------------------------|
| GET    | `/api/channels/{id}/messages` | –                                   | List messages (paginated)   |
| POST   | `/api/channels/{id}/messages` | `{"content":"hi","author":"alice"}` | Post a message to a channel |

`GET .../messages` accepts `?page=` (default `1`) and `?limit=` (default `20`)
query parameters and returns `{ messages, page, limit, total }`.

### WebSocket

| Path               | Description                                                   |
|--------------------|--------------------------------------------------------------|
| `/ws?channel={id}` | Subscribe to a channel; receive posted messages in real time |

WebSocket frames use the JSON shape:

```json
{
  "type": "message",
  "channel_id": "1",
  "author": "alice",
  "content": "hello",
  "created_at": "2026-07-01T15:30:54Z"
}
```

## Example

```bash
# Start the server
go run .

# In another terminal:
curl -X POST localhost:8080/api/channels \
  -H 'Content-Type: application/json' \
  -d '{"name":"general"}'
# => {"id":"1","name":"general","created_at":"..."}

curl -X POST localhost:8080/api/channels/1/messages \
  -H 'Content-Type: application/json' \
  -d '{"content":"hello","author":"alice"}'

curl localhost:8080/api/channels/1/messages
# => {"messages":[...],"page":1,"limit":20,"total":1}
```

## Notes

This is a demo. Storage is in memory (no database is required to run), and the
`static/` directory holds a simple browser frontend for the API.

## License

[MIT](LICENSE)
