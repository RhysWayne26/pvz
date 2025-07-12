package queries

const (
	// CreateOutboxEventSQL is an SQL query string that inserts a payload into the `outbox` table.
	CreateOutboxEventSQL = `
insert into outbox (id,payload) values ($1, $2)
`

	// FetchPendingSQL is an SQL query string for retrieving pending events from the outbox table based on status and retry logic.
	FetchPendingSQL = `
select
    id,
    payload,
    status,
    error,
    created_at,
    sent_at,
    attempts,
    last_attempt_at
from outbox where status = $1
and (last_attempt_at is null or last_attempt_at + ($2 * interval '1 second') <= now())
order by created_at
limit $3;
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
