CREATE TABLE users (
    id INT AUTO_INCREMENT,
    user text NOT NULL,
    pass text NOT NULL
);
-- name: CreateUser :exec
INSERT INTO users (user, pass) VALUES (?, ?);