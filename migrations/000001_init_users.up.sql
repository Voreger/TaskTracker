
CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       email TEXT NOT NULL UNIQUE,
                       password TEXT NOT NULL,
                       created_at TIMESTAMP DEFAULT now()
);