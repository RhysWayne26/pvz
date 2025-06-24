-- +goose Up
alter table order_history
    alter column event type varchar(32) using event::text,
    add constraint chk_history_event
        check (event in ('ACCEPTED', 'ISSUED', 'RETURNED_BY_CLIENT', 'RETURNED_TO_WAREHOUSE'));

drop type if exists event_type;

-- +goose Down
create type event_type as enum (
    'ACCEPTED',
    'ISSUED',
    'RETURNED_BY_CLIENT',
    'RETURNED_TO_WAREHOUSE'
);

alter table order_history
    drop constraint if exists chk_history_event,
    alter column event type event_type using event::event_type;
