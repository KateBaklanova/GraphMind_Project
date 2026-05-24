-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Таблица графов
CREATE TABLE IF NOT EXISTS graphs (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    summary TEXT,
    is_public BOOLEAN DEFAULT FALSE,
    view_count INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Таблица документов
CREATE TABLE IF NOT EXISTS documents (
    id VARCHAR(36) PRIMARY KEY,
    graph_id VARCHAR(36) NOT NULL REFERENCES graphs(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    content TEXT,
    size BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Таблица узлов
CREATE TABLE IF NOT EXISTS nodes (
    id VARCHAR(36) NOT NULL,
    graph_id VARCHAR(255) NOT NULL REFERENCES graphs(id) ON DELETE CASCADE,
    label VARCHAR(255) NOT NULL,
    type VARCHAR(50),
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (graph_id, id)
);

-- Таблица связей
CREATE TABLE IF NOT EXISTS edges (
    id VARCHAR(36) PRIMARY KEY,

    graph_id VARCHAR(36) NOT NULL REFERENCES graphs(id) ON DELETE CASCADE,

    from_id VARCHAR(36) NOT NULL,
    to_id VARCHAR(36) NOT NULL,

    label VARCHAR(255),
    weight FLOAT DEFAULT 1.0,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT different_nodes
    CHECK (from_id != to_id),

    CONSTRAINT edges_from_fk
    FOREIGN KEY (graph_id, from_id)
    REFERENCES nodes(graph_id, id)
    ON DELETE CASCADE,

    CONSTRAINT edges_to_fk
    FOREIGN KEY (graph_id, to_id)
    REFERENCES nodes(graph_id, id)
    ON DELETE CASCADE,

    CONSTRAINT unique_edge
    UNIQUE(graph_id, from_id, to_id, label)
);

-- Индексы для производительности
CREATE INDEX IF NOT EXISTS idx_graphs_user_id ON graphs(user_id);
CREATE INDEX IF NOT EXISTS idx_graphs_public ON graphs(is_public, view_count DESC);
CREATE INDEX IF NOT EXISTS idx_nodes_graph_id ON nodes(graph_id);
CREATE INDEX IF NOT EXISTS idx_edges_graph_id ON edges(graph_id);