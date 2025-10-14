package handlers

import (
	"GoProjects/TaskTracker/internal/models"
	"GoProjects/TaskTracker/internal/queue"
	"GoProjects/TaskTracker/internal/realtime"
	"GoProjects/TaskTracker/internal/store"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strconv"
)

type TaskHandler struct {
	Store  *store.TaskStore
	Hub    *realtime.Hub
	Broker *queue.Broker
}

func RegisterTaskRoutes(r chi.Router, s *store.TaskStore, hub *realtime.Hub, broker *queue.Broker) {
	h := &TaskHandler{Store: s, Hub: hub, Broker: broker}

	r.Route("/tasks", func(r chi.Router) {
		r.Get("/", h.ListTasks)
		r.Post("/", h.CreateTask)
		r.Get("/{id}", h.GetTask)
		r.Put("/{id}", h.UpdateTask)
		r.Delete("/{id}", h.DeleteTask)
	})
}

// ListTasks godoc
// @Summary      Get all tasks for current user
// @Description  Returns list of tasks belonging to the authenticated user
// @Tags         tasks
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   models.Task
// @Failure      401  {string}  string "unauthorized"
// @Failure      500  {string}  string "internal error"
// @Router       /tasks [get]
func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	tasks, err := h.Store.List(context.Background(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(tasks)
	if err != nil {
		return
	}
}

// CreateTask godoc
// @Summary      Create task
// @Description  Creates a new task for the authenticated user
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        task  body      models.Task  true  "Task info"
// @Security 	 BearerAuth
// @Success      201   {object}  models.Task
// @Failure      400   {string}  string "invalid input"
// @Failure      401   {string}  string "unauthorized"
// @Failure      500   {string}  string "internal error"
// @Router       /tasks [post]
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	task.UserID = userID

	if err := h.Store.Create(context.Background(), &task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err := json.NewEncoder(w).Encode(task)
	if err != nil {
		return
	}
	h.Hub.Broadcast(realtime.Message{
		Type: "task_created",
		Data: task,
	})

	go func() {
		msg := queue.EventMessage{
			Type:    queue.EventTaskCreated,
			Payload: task,
		}

		body, _ := json.Marshal(msg)
		err := h.Broker.Publish(body)
		if err != nil {
			log.Println("Publish error:", err)
		}
	}()
}

// GetTask godoc
// @Summary      Get task by ID
// @Description  Returns a single task by its ID
// @Tags         tasks
// @Produce      json
// @Param        id   path      int  true  "Task ID"
// @Security 	 BearerAuth
// @Success      200  {object}  models.Task
// @Failure      400  {string}  string "invalid id"
// @Failure      401  {string}  string "unauthorized"
// @Failure      500  {string}  string "internal error"
// @Router       /tasks/{id} [get]
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	task, err := h.Store.Get(context.Background(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(task)
	if err != nil {
		return
	}
}

// UpdateTask godoc
// @Summary      Update task
// @Description  Updates an existing task for the authenticated user
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id    path      int         true  "Task ID"
// @Param        task  body      models.Task true  "Task info"
// @Security 	 BearerAuth
// @Success      200   {object}  models.Task
// @Failure      400   {string}  string "invalid input"
// @Failure      401   {string}  string "unauthorized"
// @Failure      500   {string}  string "internal error"
// @Router       /tasks/{id} [put]
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var t models.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	t.ID = id
	updated, err := h.Store.Update(context.Background(), &t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(updated)
	if err != nil {
		return
	}
	h.Hub.Broadcast(realtime.Message{
		Type: "task_updated",
		Data: updated,
	})

	go func() {
		msg := queue.EventMessage{
			Type:    queue.EventTaskUpdated,
			Payload: updated,
		}

		body, _ := json.Marshal(msg)
		err := h.Broker.Publish(body)
		if err != nil {
			log.Println("Publish error:", err)
		}
	}()
}

// DeleteTask godoc
// @Summary      Delete task
// @Description  Deletes a task by ID for the authenticated user
// @Tags         tasks
// @Param        id   path      int  true  "Task ID"
// @Security 	 BearerAuth
// @Success      204  {string}  string "no content"
// @Failure      400  {string}  string "invalid id"
// @Failure      401  {string}  string "unauthorized"
// @Failure      500  {string}  string "internal error"
// @Router       /tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	err = h.Store.Delete(context.Background(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	h.Hub.Broadcast(realtime.Message{
		Type: "task_deleted",
		Data: map[string]int{"id": id},
	})

	go func() {
		msg := queue.EventMessage{
			Type:    queue.EventTaskDeleted,
			Payload: id,
		}

		body, _ := json.Marshal(msg)
		err := h.Broker.Publish(body)
		if err != nil {
			log.Println("Publish error:", err)
		}
	}()
}
