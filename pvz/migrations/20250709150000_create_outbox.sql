-- +goose Up
create table if not exists outbox(
    id bigserial primary key,
    payload jsonb not null,
    status integer not null default 1,
    error text not null default '',
    created_at timestamptz not null default now(),
    sent_at timestamptz,
    attempts integer not null default 0,
    last_attempt_at timestamptz
);

create index if not exists idx_outbox_status_created on outbox(status, created_at);

create index if not exists idx_outbox_retry_at on outbox(status, last_attempt_at);

-- +goose Down
drop table if exists outbox;