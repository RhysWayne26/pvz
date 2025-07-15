package queries

const (
	// CreateOutboxEventSQL is an SQL query string that inserts a payload into the `outbox` table.
	CreateOutboxEventSQL = `
insert into outbox (id, order_id, payload) values ($1, $2, $3)
`

	// SetProcessingSQL marks a limited number of CREATED events as PROCESSING
	SetProcessingSQL = `
with to_update as (
    select id
    from outbox
    where status = 1
      and (last_attempt_at is null or last_attempt_at + ($1 * interval '1 second') <= now())
    order by created_at
    limit $2
    for update skip locked
)
update outbox o
set status = 2,
    attempts = attempts + 1,
    last_attempt_at = now()
from to_update tu
where o.id = tu.id;
`

	// GetProcessingEventsSQL retrieves events with status PROCESSING that are ready for processing, considering retry delay and concurrency.
	GetProcessingEventsSQL = `
select id, order_id, payload, status, error, created_at, sent_at, attempts, last_attempt_at
from outbox
where status = 2
  and (last_attempt_at is null or last_attempt_at + ($1 * interval '1 second') <= now())
order by created_at, coalesce(last_attempt_at, created_at)
for update skip locked
limit $2;
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

	// UpdateErrorSQL is a SQL statement to update the error message of an outbox entry identified by its ID.
	UpdateErrorSQL = `
update outbox
set error = $2
where id = $1
`
)
