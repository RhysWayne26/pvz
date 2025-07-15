-- +goose Up
create table if not exists outbox(
    id bigint primary key,
    order_id bigint not null,
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

create index if not exists idx_outbox_create_at_retry_at on outbox(created_at, last_attempt_at);

-- +goose Down
drop table if exists outbox;