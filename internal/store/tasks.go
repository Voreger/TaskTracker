package store

import (
	"GoProjects/TaskTracker/internal/logger"
	"GoProjects/TaskTracker/internal/models"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type TaskStore struct {
	Pool *pgxpool.Pool
}

func NewTaskStore(pool *pgxpool.Pool) *TaskStore {
	return &TaskStore{Pool: pool}
}

// Create
func (s *TaskStore) Create(ctx context.Context, t *models.Task) error {
	query := `INSERT INTO tasks (title, description, status, user_id) 
			  VALUES ($1, $2, $3, $4) returning id, user_id, created_at, updated_at;`
	return s.Pool.QueryRow(ctx, query, t.Title, t.Description, t.Status, t.UserID).Scan(&t.ID, &t.UserID, &t.CreatedAt, &t.UpdatedAt)
}

// Get by id
func (s *TaskStore) Get(ctx context.Context, id int) (t *models.Task, err error) {
	t = &models.Task{}
	query := `SELECT id, title, description, status, user_id, created_at, updated_at FROM tasks WHERE id = $1;`
	err = s.Pool.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.Title, &t.Description, &t.Status, &t.UserID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// List all
func (s *TaskStore) List(ctx context.Context, userID int) ([]*models.Task, error) {
	query := `SELECT id, title, description, status, user_id, created_at, updated_at FROM tasks where user_id = $1`
	rows, err := s.Pool.Query(ctx, query, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []*models.Task{}
	for rows.Next() {
		t := &models.Task{}
		err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.UserID, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			logger.Log.Error("Scan error", zap.Error(err))
			continue
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// Update
func (s *TaskStore) Update(ctx context.Context, t *models.Task) (*models.Task, error) {
	query := `
        UPDATE tasks
        SET title=$1, description=$2, status=$3, updated_at=now()
        WHERE id=$4
        RETURNING id, title, description, status, user_id, created_at, updated_at
    `
	var updated models.Task
	err := s.Pool.QueryRow(ctx, query, t.Title, t.Description, t.Status, t.ID).
		Scan(&updated.ID, &updated.Title, &updated.Description, &updated.Status, &updated.UserID, &updated.CreatedAt, &updated.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

// Delete
func (s *TaskStore) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM tasks WHERE id=$1`
	_, err := s.Pool.Exec(ctx, query, id)
	return err
}
