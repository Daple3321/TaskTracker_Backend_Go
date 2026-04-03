-- Widen columns for real titles and markdown descriptions (VARCHAR(45) was too small).
ALTER TABLE tasks
    MODIFY COLUMN task_name VARCHAR(255),
    MODIFY COLUMN task_description TEXT;
