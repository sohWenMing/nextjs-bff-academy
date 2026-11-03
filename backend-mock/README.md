# TeamPulse mock Go backend

The API your Next.js BFF talks to throughout this course. Zero external
dependencies — stdlib only.

## Run it

```bash
cd backend-mock
go run main.go
# TeamPulse mock backend listening on http://localhost:8080
```

Requires Go 1.22+ (`go version`). Leave it running in a dedicated terminal tab
while you work — every project session assumes `http://localhost:8080` is up.

## Endpoints

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/api/health` | no | Liveness check: `{"status":"ok","time":"..."}` |
| POST | `/api/login` | no | Body `{"username":"any","password":"demo"}` → returns `token` + `user` |
| GET | `/api/me` | yes | Current user: `{"id","name","role"}` |
| GET | `/api/projects` | yes | All projects |
| GET | `/api/projects/{id}` | yes | One project (404 if unknown) |
| GET | `/api/tasks?projectId=p1` | yes | Tasks, optionally filtered |
| POST | `/api/tasks` | yes | Body `{"projectId","title","priority"}` → 201 + created task |
| GET | `/api/flaky` | yes | Fails ~30% of the time with 500 (resilience practice) |
| GET | `/api/slow` | no | Always takes 2 seconds (timeout practice) |

**Auth:** send `Authorization: Bearer <token>` where `<token>` came from
`/api/login`. Password is always `demo`; you always log in as user `u1` ("Wen", admin).

Data is in-memory and resets on restart.

### Example

```bash
curl -s http://localhost:8080/api/health

# log in and capture the token (python3 parses the JSON; no jq needed)
TOKEN=$(curl -s -X POST localhost:8080/api/login \
  -H 'Content-Type: application/json' \
  -d '{"username": "wen", "password": "demo"}' \
  | python3 -c 'import sys,json;print(json.load(sys.stdin)["token"])')

curl -s http://localhost:8080/api/projects -H "Authorization: Bearer $TOKEN"
```

## The data model (what your TS types must mirror)

```go
User    { ID, Name, Role string }                        // json: id, name, role
Project { ID, Name, Description, Status string; CreatedAt time.Time }
Task    { ID, ProjectID, Title, Status, Assignee string; Priority int; DueDate time.Time }
```

`status` values: Project `active|archived` · Task `todo|doing|done`.
Errors are always `{"error": "message"}` with an appropriate HTTP status code.

## Reading the source

`main.go` is deliberately short and commented. Lessons 24–28 ask you to read
specific parts of it — you do **not** need to know Go to follow them; treat it
as a contract you consume, exactly like the real team's Go backend.
