-- River Queue tables migration
-- This migration creates the necessary tables for River Queue job processing

-- River jobs table
CREATE TABLE IF NOT EXISTS river_job (
    id bigserial PRIMARY KEY,
    args jsonb NOT NULL,
    attempt smallint NOT NULL DEFAULT 0,
    attempted_at timestamptz,
    attempted_by text[],
    created_at timestamptz NOT NULL DEFAULT NOW(),
    errors jsonb[],
    finalized_at timestamptz,
    kind text NOT NULL,
    max_attempts smallint NOT NULL,
    metadata jsonb NOT NULL DEFAULT '{}',
    priority smallint NOT NULL DEFAULT 1,
    queue text NOT NULL DEFAULT 'default',
    scheduled_at timestamptz NOT NULL DEFAULT NOW(),
    state text NOT NULL DEFAULT 'available',
    tags text[] NOT NULL DEFAULT '{}',
    unique_key text,
    unique_states bit(8),
    CONSTRAINT finalized_or_finalized_at_null CHECK (
        (finalized_at IS NULL AND state NOT IN ('cancelled', 'completed', 'discarded')) OR
        (finalized_at IS NOT NULL AND state IN ('cancelled', 'completed', 'discarded'))
    ),
    CONSTRAINT kind_length CHECK (char_length(kind) > 0 AND char_length(kind) < 128),
    CONSTRAINT max_attempts_is_positive CHECK (max_attempts > 0),
    CONSTRAINT priority_in_range CHECK (priority >= 1 AND priority <= 4),
    CONSTRAINT queue_length CHECK (char_length(queue) > 0 AND char_length(queue) < 128),
    CONSTRAINT state_in_set CHECK (state IN ('available', 'cancelled', 'completed', 'discarded', 'pending', 'retryable', 'running', 'scheduled'))
);

-- Indexes for River job queries
CREATE INDEX IF NOT EXISTS river_job_args_index ON river_job USING GIN (args);
CREATE INDEX IF NOT EXISTS river_job_kind ON river_job (kind);
CREATE INDEX IF NOT EXISTS river_job_state_and_finalized_at_index ON river_job (state, finalized_at) WHERE finalized_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS river_job_prioritized_fetching_index ON river_job (state, queue, priority, scheduled_at, id) WHERE state IN ('available', 'retryable');
CREATE INDEX IF NOT EXISTS river_job_scheduled_at_index ON river_job (scheduled_at) WHERE state = 'scheduled';

-- Unique constraint for unique jobs
CREATE UNIQUE INDEX IF NOT EXISTS river_job_unique_idx ON river_job (unique_key) WHERE unique_key IS NOT NULL AND unique_states IS NOT NULL AND (unique_states & (b'00000001'::bit(8) << state::int)) != b'00000000'::bit(8);

-- River leader table
CREATE TABLE IF NOT EXISTS river_leader (
    elected_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    leader_id text NOT NULL,
    name text PRIMARY KEY,
    CONSTRAINT name_length CHECK (char_length(name) > 0 AND char_length(name) < 128),
    CONSTRAINT leader_id_length CHECK (char_length(leader_id) > 0 AND char_length(leader_id) < 128)
);

-- River migration table
CREATE TABLE IF NOT EXISTS river_migration (
    id bigserial PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    version bigint NOT NULL,
    CONSTRAINT version_gte_1 CHECK (version >= 1)
);

CREATE UNIQUE INDEX IF NOT EXISTS river_migration_version_idx ON river_migration (version);

-- River queue table for queue configuration
CREATE TABLE IF NOT EXISTS river_queue (
    name text PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    metadata jsonb NOT NULL DEFAULT '{}',
    paused_at timestamptz,
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    CONSTRAINT name_length CHECK (char_length(name) > 0 AND char_length(name) < 128)
);

-- Insert default queue
INSERT INTO river_queue (name) VALUES ('default') ON CONFLICT DO NOTHING;

-- Insert River schema version
INSERT INTO river_migration (version) VALUES (1) ON CONFLICT DO NOTHING;

-- Comments for documentation
COMMENT ON TABLE river_job IS 'River Queue jobs table for background job processing';
COMMENT ON TABLE river_leader IS 'River Queue leader election table';
COMMENT ON TABLE river_migration IS 'River Queue schema version tracking';
COMMENT ON TABLE river_queue IS 'River Queue configuration and state tracking';
COMMENT ON COLUMN river_job.kind IS 'Job type identifier (e.g., deliver_message)';
COMMENT ON COLUMN river_job.state IS 'Current job state: available, pending, running, completed, cancelled, discarded, retryable, scheduled';
COMMENT ON COLUMN river_job.scheduled_at IS 'When the job should be executed';
COMMENT ON COLUMN river_job.args IS 'JSON arguments for the job';