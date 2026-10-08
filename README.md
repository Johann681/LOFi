# Student Matching API

A Fiber v2 API backed by MongoDB. The server requires `MONGO_URI`; `MONGO_DATABASE` and `PORT` are optional.

## Run

```powershell
$env:MONGO_URI = "mongodb://localhost:27017"
go run ./cmd/server
```

The default database is `student_matching`, and the default HTTP port is `3000`.

## Endpoints

- `POST /api/students/register` creates a student. Required JSON fields: `name`, `gender`, `orientation`, `age`, `height`, `department`, `class`, `verificationProof`, and both `preferences.targetGender` and `preferences.orientation`. `likes` and `dislikes` are optional arrays of strings.
- `GET /api/students/{id}` fetches a student by its 24-character MongoDB ObjectID hex string.
- `POST /api/match/find` accepts `{"studentId":"<object-id>"}`. Returns `matched` with the partner profile and match ID, or `searching` with HTTP 202 while waiting in the MongoDB-backed queue.
- `GET /api/match/history/{student_id}` returns past and active matches with each partner's profile.
- `GET /ws/chat?student_id={object-id}` upgrades to a WebSocket. Send chat JSON as `{"matchId":"<object-id>","content":"Hello"}`. Messages are accepted only for an active match containing that student, saved asynchronously in `messages`, and delivered to the partner and sender when connected.

The matching service creates indexes on startup. It filters candidates against each other's gender and orientation preferences (`any` is a wildcard), then scores shared interests and department/class overlap. It prevents self-matches, repeat pairs, and simultaneous active matches using MongoDB unique indexes. Registration requires the student's own `gender` and `orientation` as well as both preference fields.

WebSocket identity currently comes directly from the `student_id` query parameter because authentication is not enabled. Do not expose this endpoint publicly until that value is bound to an authenticated session or token.

Example registration body:

```json
{
  "name": "Alex Student",
  "gender": "woman",
  "orientation": "straight",
  "age": 20,
  "height": "170 cm",
  "department": "Computer Science",
  "class": "2028",
  "verificationProof": "https://example.edu/verify/abc123",
  "preferences": {
    "targetGender": "any",
    "orientation": "straight"
  },
  "likes": ["music", "hiking"],
  "dislikes": ["smoking"]
}
```
