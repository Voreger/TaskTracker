package store

import (
	"GoProjects/TaskTracker/internal/logger"
	"GoProjects/TaskTracker/internal/models"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type UserStore struct {
	Pool *pgxpool.Pool
}

func NewUserStore(pool *pgxpool.Pool) *UserStore {
	return &UserStore{Pool: pool}
}

// Create
func (s *UserStore) Create(ctx context.Context, t *models.User) error {
	query := `INSERT INTO users (email, password) values ($1, $2) returning id, created_at`
	return s.Pool.QueryRow(ctx, query, t.Email, t.Password).Scan(&t.ID, &t.CreatedAt)
}

// Get by id
func (s *UserStore) Get(ctx context.Context, id int) (*models.User, error) {
	u := &models.User{}
	query := `SELECT id, email, password, created_at FROM users WHERE id = $1;`
	err := s.Pool.QueryRow(ctx, query, id).Scan(&u.ID, &u.Email, &u.Password, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// Get by email
func (s *UserStore) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	u := &models.User{}
	query := `SELECT id, email, password, created_at FROM users WHERE email = $1;`
	err := s.Pool.QueryRow(ctx, query, email).Scan(&u.ID, &u.Email, &u.Password, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// Get all users
func (s *UserStore) List(ctx context.Context) ([]*models.User, error) {
	query := `SELECT id, email, password, created_at FROM users;`
	rows, err := s.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*models.User
	for rows.Next() {
		u := &models.User{}
		err := rows.Scan(&u.ID, &u.Email, &u.Password, &u.CreatedAt)
		if err != nil {
			logger.Log.Error("scan error", zap.Error(err))
			continue
		}
		users = append(users, u)
	}
	return users, nil
}

// Update
func (s *UserStore) Update(ctx context.Context, t *models.User) (*models.User, error) {
	query := `UPDATE users 
		 	  SET email = $1, password = $2 
		 	  WHERE id = $3
			  returning id, email, password, created_at;`
	var updated models.User
	err := s.Pool.QueryRow(ctx, query, t.Email, t.Password, t.ID).
		Scan(&updated.ID, &updated.Email, &updated.Password, &updated.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

// Delete
func (s *UserStore) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := s.Pool.Exec(ctx, query, id)
	return err
}
