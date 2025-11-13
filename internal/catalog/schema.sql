-- Conduit Model Catalog Database Schema
-- SQLite compatible, PostgreSQL ready

-- Models table - stores model metadata
CREATE TABLE IF NOT EXISTS models (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    domain TEXT NOT NULL,
    description TEXT,
    github_repo TEXT,
    license TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_models_name ON models(name);
CREATE INDEX IF NOT EXISTS idx_models_domain ON models(domain);

-- Model versions - tracks different versions of each model
CREATE TABLE IF NOT EXISTS model_versions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    model_id INTEGER NOT NULL,
    version TEXT NOT NULL,
    weights_uri TEXT NOT NULL,
    weights_size_gb REAL,
    checksum_sha256 TEXT,

    -- Runtime configuration
    framework TEXT NOT NULL,
    python_version TEXT NOT NULL,
    dependencies TEXT,
    custom_image TEXT,

    -- Inference configuration
    entrypoint TEXT NOT NULL,
    handler TEXT NOT NULL,

    -- Hardware requirements
    gpu_required BOOLEAN DEFAULT FALSE,
    recommended_instance TEXT,
    min_cpu INTEGER,
    min_memory_gb INTEGER,
    min_gpu_memory_gb INTEGER,

    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_latest BOOLEAN DEFAULT FALSE,

    FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE,
    UNIQUE(model_id, version)
);

CREATE INDEX IF NOT EXISTS idx_version_model ON model_versions(model_id);
CREATE INDEX IF NOT EXISTS idx_version_latest ON model_versions(model_id, is_latest);

-- Benchmarks - performance metrics for model versions
CREATE TABLE IF NOT EXISTS benchmarks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    model_version_id INTEGER NOT NULL,
    dataset TEXT NOT NULL,
    metric TEXT NOT NULL,
    result REAL NOT NULL,
    instance TEXT,
    cost_per_prediction TEXT,
    walltime_seconds REAL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (model_version_id) REFERENCES model_versions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_benchmarks_version ON benchmarks(model_version_id);
CREATE INDEX IF NOT EXISTS idx_benchmarks_dataset ON benchmarks(dataset);

-- Tags - for categorization and search
CREATE TABLE IF NOT EXISTS tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Model tags - many-to-many relationship
CREATE TABLE IF NOT EXISTS model_tags (
    model_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY (model_id, tag_id),
    FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- Citations - paper information for models
CREATE TABLE IF NOT EXISTS citations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    model_id INTEGER NOT NULL,
    paper_title TEXT,
    paper_url TEXT,
    doi TEXT,
    authors TEXT,
    year INTEGER,
    bibtex TEXT,

    FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_citations_model ON citations(model_id);

-- Model statistics - tracks usage and popularity
CREATE TABLE IF NOT EXISTS model_stats (
    model_id INTEGER PRIMARY KEY,
    total_deployments INTEGER DEFAULT 0,
    total_predictions INTEGER DEFAULT 0,
    view_count INTEGER DEFAULT 0,
    last_deployed_at TIMESTAMP,
    last_viewed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_stats_deployments ON model_stats(total_deployments);
CREATE INDEX IF NOT EXISTS idx_stats_predictions ON model_stats(total_predictions);
