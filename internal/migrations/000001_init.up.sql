CREATE TABLE users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at DATETIME
);

CREATE TABLE tasks (
    id INT PRIMARY KEY AUTO_INCREMENT,
    task_name VARCHAR(255),
    task_description TEXT,
    created_at DATETIME,
    updated_at DATETIME,
    user_id INT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);