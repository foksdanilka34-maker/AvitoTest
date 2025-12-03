CREATE UNIQUE INDEX idx_op_pr ON pull_requests (request_name)
WHERE status = 'OPEN';