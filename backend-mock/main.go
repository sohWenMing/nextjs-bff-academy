// TeamPulse mock Go backend — the API your Next.js BFF talks to in this course.
//
// Run:  go run main.go   (listens on :8080, no external dependencies)
// See README.md in this folder for the full endpoint contract.
package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"math"
	"math/big"
	"net/http"
	"strings"
	"time"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // active | archived
	CreatedAt   time.Time `json:"createdAt"`
}

type Task struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"projectId"`
	Title     string    `json:"title"`
	Status    string    `json:"status"` // todo | doing | done
	Priority  int       `json:"priority"`
	DueDate   time.Time `json:"dueDate"`
	Assignee  string    `json:"assignee"`
}

var (
	users = []User{
		{"u1", "Wen", "admin"},
		{"u2", "Sarah", "member"},
	}
	projects = []Project{
		{"p1", "Billing Service", "Go rewrite of the billing pipeline", "active", mustDate("2026-05-01")},
		{"p2", "Mobile App", "Customer-facing iOS/Android app", "active", mustDate("2026-06-15")},
		{"p3", "Legacy Portal", "Old PHP portal, being retired", "archived", mustDate("2024-01-10")},
	}
	tasks = []Task{
		{"t1", "p1", "Implement invoice PDF export", "doing", 1, mustDate("2026-11-20"), "Wen"},
		{"t2", "p1", "Add retry to payment webhook", "todo", 2, mustDate("2026-11-25"), "Sarah"},
		{"t3", "p2", "Fix login screen keyboard overlap", "done", 1, mustDate("2026-11-01"), "Wen"},
		{"t4", "p2", "Push notification opt-in", "todo", 3, mustDate("2026-12-05"), "Sarah"},
		{"t5", "p1", "Load test checkout endpoint", "todo", 2, mustDate("2026-12-01"), "Wen"},
	}
	tokens = map[string]string{} // token -> userID
)

func mustDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// auth returns the user ID if the Authorization: Bearer <token> header is valid.
func auth(r *http.Request) (string, bool) {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return "", false
	}
	uid, ok := tokens[strings.TrimPrefix(h, "Bearer ")]
	return uid, ok
}

func requireAuth(w http.ResponseWriter, r *http.Request) (string, bool) {
	uid, ok := auth(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "missing or invalid token")
		return "", false
	}
	return uid, true
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "time": time.Now().UTC()})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "use POST")
		return
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	// Demo rule: any username, password must be "demo".
	if body.Password != "demo" {
		writeErr(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	uid := "u1" // always log in as Wen for the demo
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	token := "tok-" + hex.EncodeToString(b)
	tokens[token] = uid
	var user User
	for _, u := range users {
		if u.ID == uid {
			user = u
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "user": user})
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	uid, ok := requireAuth(w, r)
	if !ok {
		return
	}
	for _, u := range users {
		if u.ID == uid {
			writeJSON(w, http.StatusOK, u)
			return
		}
	}
	writeErr(w, http.StatusInternalServerError, "user vanished")
}

func handleProjects(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAuth(w, r); !ok {
		return
	}
	// /api/projects/{id}
	if id := strings.TrimPrefix(r.URL.Path, "/api/projects/"); id != r.URL.Path {
		for _, p := range projects {
			if p.ID == id {
				writeJSON(w, http.StatusOK, p)
				return
			}
		}
		writeErr(w, http.StatusNotFound, "project not found")
		return
	}
	writeJSON(w, http.StatusOK, projects)
}

func handleTasks(w http.ResponseWriter, r *http.Request) {
	uid, ok := requireAuth(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		projectID := r.URL.Query().Get("projectId")
		out := []Task{}
		for _, t := range tasks {
			if projectID == "" || t.ProjectID == projectID {
				out = append(out, t)
			}
		}
		writeJSON(w, http.StatusOK, out)
	case http.MethodPost:
		var body struct {
			ProjectID string `json:"projectId"`
			Title     string `json:"title"`
			Priority  int    `json:"priority"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		if body.ProjectID == "" || body.Title == "" {
			writeErr(w, http.StatusBadRequest, "projectId and title are required")
			return
		}
		n, _ := rand.Int(rand.Reader, big.NewInt(math.MaxInt32))
		t := Task{
			ID:        "t" + n.String(),
			ProjectID: body.ProjectID,
			Title:     body.Title,
			Status:    "todo",
			Priority:  body.Priority,
			DueDate:   time.Now().AddDate(0, 0, 7),
			Assignee:  uid,
		}
		tasks = append(tasks, t)
		writeJSON(w, http.StatusCreated, t)
	default:
		writeErr(w, http.StatusMethodNotAllowed, "use GET or POST")
	}
}

// handleFlaky fails ~30% of the time with a 500 — for the resilience lesson.
func handleFlaky(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAuth(w, r); !ok {
		return
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(10))
	if n.Int64() < 3 {
		writeErr(w, http.StatusInternalServerError, "upstream hiccup (intentional: 30% failure rate)")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "note": "this endpoint fails ~30% of the time on purpose"})
}

// handleSlow sleeps 2s — for the timeout lesson.
func handleSlow(w http.ResponseWriter, r *http.Request) {
	time.Sleep(2 * time.Second)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "note": "I slept 2 seconds on purpose"})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/login", handleLogin)
	mux.HandleFunc("/api/me", handleMe)
	mux.HandleFunc("/api/projects", handleProjects)
	mux.HandleFunc("/api/projects/", handleProjects)
	mux.HandleFunc("/api/tasks", handleTasks)
	mux.HandleFunc("/api/flaky", handleFlaky)
	mux.HandleFunc("/api/slow", handleSlow)

	log.Println("TeamPulse mock backend listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
