​​CREATE TABLE tasks IF NOT EXISTS (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT,
    description TEXT,
    status TEXT,
    created_at DATETIME
);
