package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"crud-service/internal/model"
	"crud-service/internal/repository"
	"crud-service/internal/service"
)

// Handler holds all service dependencies.
type Handler struct {
	employees service.EmployeeService
	slots     service.SlotService
	appeals   service.AppealService
	subthemes service.SubthemeService
	clients   service.ClientService
	themes    service.ThemeService
	teams     service.TeamService
	workflows service.WorkflowService
}

// New returns a new Handler.
func New(
	employees service.EmployeeService,
	slots service.SlotService,
	appeals service.AppealService,
	subthemes service.SubthemeService,
	clients service.ClientService,
	themes service.ThemeService,
	teams service.TeamService,
	workflows service.WorkflowService,
) *Handler {
	return &Handler{
		employees: employees,
		slots:     slots,
		appeals:   appeals,
		subthemes: subthemes,
		clients:   clients,
		themes:    themes,
		teams:     teams,
		workflows: workflows,
	}
}

// InitRoutes registers all routes and returns the resulting ServeMux.
func (h *Handler) InitRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/employees", h.employeesCollection)
	mux.HandleFunc("/employees/", h.employeesResource)

	mux.HandleFunc("/slots", h.slotsCollection)
	mux.HandleFunc("/slots/", h.slotsResource)
	mux.HandleFunc("/slots/count", h.changeSlotsCount)

	mux.HandleFunc("/appeals", h.appealsCollection)
	mux.HandleFunc("/appeals/", h.appealsResource)

	mux.HandleFunc("/subthemes", h.subthemesCollection)
	mux.HandleFunc("/subthemes/", h.subthemesResource)

	mux.HandleFunc("/clients", h.clientsCollection)
	mux.HandleFunc("/clients/", h.clientsResource)

	mux.HandleFunc("/themes", h.themesCollection)
	mux.HandleFunc("/themes/", h.themesResource)

	mux.HandleFunc("/teams", h.teamsCollection)
	mux.HandleFunc("/teams/", h.teamsResource)

	mux.HandleFunc("/workflows", h.workflowsCollection)

	return corsMiddleware(mux)
}

// corsMiddleware разрешает запросы с любого origin (нужно для браузерных клиентов).
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Preflight-запрос браузера
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func parseID(w http.ResponseWriter, r *http.Request, prefix string) (int, bool) {
	idStr := strings.TrimPrefix(r.URL.Path, prefix)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

func notFoundOrInternal(w http.ResponseWriter, err error) {
	if errors.Is(err, repository.ErrNotFound) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// ─── Employee ─────────────────────────────────────────────────────────────────

func (h *Handler) employeesCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := h.employees.GetAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var e model.Employee
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		created, err := h.employees.Create(e)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) employeesResource(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "/employees/")
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		item, err := h.employees.GetByID(id)
		if err != nil {
			notFoundOrInternal(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	case http.MethodPut:
		var e model.Employee
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		updated, err := h.employees.Update(id, e)
		if err != nil {
			notFoundOrInternal(w, err)
			return
		}
		writeJSON(w, http.StatusOK, updated)
	case http.MethodDelete:
		if err := h.employees.Delete(id); err != nil {
			notFoundOrInternal(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// ─── Slot ─────────────────────────────────────────────────────────────────────

func (h *Handler) slotsCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := h.slots.GetAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var s model.Slot
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		created, err := h.slots.Create(s)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) slotsResource(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "/slots/")
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		item, err := h.slots.GetByID(id)
		if err != nil {
			notFoundOrInternal(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	case http.MethodDelete:
		if err := h.slots.Delete(id); err != nil {
			notFoundOrInternal(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) changeSlotsCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody struct {
		EmployeeID int `json:"employeeId"`
		Count      int `json:"count"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	err := h.slots.UpdateCount(reqBody.EmployeeID, reqBody.Count)
	if err != nil {
		notFoundOrInternal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nil)
}

// ─── Appeal ───────────────────────────────────────────────────────────────────

func (h *Handler) appealsCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := h.appeals.GetAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var a model.Appeal
		if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		created, err := h.appeals.Create(a)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) appealsResource(w http.ResponseWriter, r *http.Request) {
	// POST /appeals/:id/close
	if strings.HasSuffix(r.URL.Path, "/close") {
		idStr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/appeals/"), "/close")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		err = h.appeals.Close(id)
		if err != nil {
			notFoundOrInternal(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	id, ok := parseID(w, r, "/appeals/")
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		item, err := h.appeals.GetByID(id)
		if err != nil {
			notFoundOrInternal(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	case http.MethodPut:
		var a model.Appeal
		if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		updated, err := h.appeals.Update(id, a)
		if err != nil {
			notFoundOrInternal(w, err)
			return
		}
		writeJSON(w, http.StatusOK, updated)
	case http.MethodDelete:
		if err := h.appeals.Delete(id); err != nil {
			notFoundOrInternal(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// ─── Subtheme ─────────────────────────────────────────────────────────────────

func (h *Handler) subthemesCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := h.subthemes.GetAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var s model.Subtheme
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		created, err := h.subthemes.Create(s)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) subthemesResource(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "/subthemes/")
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		item, err := h.subthemes.GetByID(id)
		if err != nil {
			notFoundOrInternal(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	case http.MethodPut:
		var s model.Subtheme
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		updated, err := h.subthemes.Update(id, s)
		if err != nil {
			notFoundOrInternal(w, err)
			return
		}
		writeJSON(w, http.StatusOK, updated)
	case http.MethodDelete:
		if err := h.subthemes.Delete(id); err != nil {
			notFoundOrInternal(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// ─── Client ───────────────────────────────────────────────────────────────────

func (h *Handler) clientsCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := h.clients.GetAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var c model.Client
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		created, err := h.clients.Create(c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) clientsResource(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "/clients/")
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		item, err := h.clients.GetByID(id)
		if err != nil {
			notFoundOrInternal(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	case http.MethodPut:
		var c model.Client
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		updated, err := h.clients.Update(id, c)
		if err != nil {
			notFoundOrInternal(w, err)
			return
		}
		writeJSON(w, http.StatusOK, updated)
	case http.MethodDelete:
		if err := h.clients.Delete(id); err != nil {
			notFoundOrInternal(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// ─── Theme ────────────────────────────────────────────────────────────────────

func (h *Handler) themesCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := h.themes.GetAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var t model.Theme
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		created, err := h.themes.Create(t)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) themesResource(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "/themes/")
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		item, err := h.themes.GetByID(id)
		if err != nil {
			notFoundOrInternal(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	case http.MethodPut:
		var t model.Theme
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		updated, err := h.themes.Update(id, t)
		if err != nil {
			notFoundOrInternal(w, err)
			return
		}
		writeJSON(w, http.StatusOK, updated)
	case http.MethodDelete:
		if err := h.themes.Delete(id); err != nil {
			notFoundOrInternal(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// ///////////////////////////////////////////////// Team ////////////////////////////////////////////////////////////////////////////
func (h *Handler) teamsCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := h.teams.GetAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var t model.Team
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		created, err := h.teams.Create(t)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) teamsResource(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "/teams/")
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		item, err := h.teams.GetByID(id)
		if err != nil {
			notFoundOrInternal(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	case http.MethodPut:
		var t model.Team
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		updated, err := h.teams.Update(id, t)
		if err != nil {
			notFoundOrInternal(w, err)
			return
		}
		writeJSON(w, http.StatusOK, updated)
	case http.MethodDelete:
		if err := h.teams.Delete(id); err != nil {
			notFoundOrInternal(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) workflowsCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := h.workflows.GetAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var w model.Workflow
		if err := json.NewDecoder(r.Body).Decode(&w); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		created, err := h.workflows.Create(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
