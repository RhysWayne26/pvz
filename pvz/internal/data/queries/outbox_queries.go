package queries

const (
	// CreateOutboxEventSQL is an SQL query string that inserts a payload into the `outbox` table.
	CreateOutboxEventSQL = `
insert into outbox (id,payload) values ($1, $2)
`

	// MarkAsProcessingSQL marks a limited number of CREATED events as PROCESSING and returns them
	MarkAsProcessingSQL = `
UPDATE outbox
SET status = 2, attempts = attempts + 1, last_attempt_at = now()
WHERE id IN (
	SELECT id FROM outbox
	WHERE status = 1
	  AND (last_attempt_at IS NULL OR last_attempt_at + ($1 * interval '1 second') <= now())
	ORDER BY created_at
	LIMIT $2
)
RETURNING
	id,
	payload,
	status,
	error,
	created_at,
	sent_at,
	attempts,
	last_attempt_at;
`

	// SetProcessingSQL is an SQL query string that updates the status and attempts of an outbox event to mark it as processing.
	SetProcessingSQL = `
update outbox 
set status = 2, attempts = attempts + 1, last_attempt_at = now() 
where id = $1;
`

	// SetCompletedSQL is an SQL query string that updates the status of an outbox event to completed and sets the sent timestamp.
	SetCompletedSQL = `
update outbox 
set status = 3, sent_at = $2
where id = $1;
`

	// SetFailedSQL is an SQL query string that updates an outbox event's status to failed and sets the error message.
	SetFailedSQL = `
update outbox
set status = 4, error = $2
where id = $1;
`

	UpdateErrorSQL = `
update outbox
set error = $2
where id = $1
`
)
