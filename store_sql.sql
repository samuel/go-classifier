-- SQLite3

CREATE TABLE categories (
    id INTEGER PRIMARY KEY ASC,
    name TEXT NOT NULL,
    document_count INTEGER NOT NULL DEFAULT 0,
    UNIQUE(name)
);

CREATE TABLE tokens (
    id INTEGER PRIMARY KEY ASC,
    category_id INTEGER NOT NULL,
    token TEXT NOT NULL,
    count INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY(category_id) REFERENCES categories(id),
    UNIQUE(category_id, token)
);
