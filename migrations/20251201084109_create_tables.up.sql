CREATE TYPE request_status AS ENUM ('OPEN', 'MERGED');

CREATE TABLE IF NOT EXISTS teams (
    team_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_name VARCHAR(100) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_name VARCHAR(100) UNIQUE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    team_id UUID REFERENCES teams(team_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS pull_requests (
    request_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_name TEXT NOT NULL,

    creator_id UUID REFERENCES users(user_id) ON DELETE SET NULL,
    status request_status DEFAULT 'OPEN',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    merged_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS pr_reviewers (
    request_id UUID REFERENCES pull_requests(request_id) ON DELETE CASCADE,
    reviewer_id UUID REFERENCES users(user_id) ON DELETE CASCADE,

    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY(request_id, reviewer_id)
);