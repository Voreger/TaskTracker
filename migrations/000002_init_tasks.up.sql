CREATE TABLE IF NOT EXISTS tasks (
                                     id SERIAL PRIMARY KEY,
                                     title TEXT NOT NULL,
                                     description TEXT,
                                     status TEXT NOT NULL DEFAULT 'todo',
                                     user_id INT REFERENCES users(id) on delete cascade,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
    );
