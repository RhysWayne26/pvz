-- +goose Up
alter table order_history drop constraint if exists chk_history_event;

update order_history
set event = case
    when event = 'ACCEPTED' then '1'
    when event = 'ISSUED' then '2'
    when event = 'RETURNED_BY_CLIENT' then '3'
    when event = 'RETURNED_TO_WAREHOUSE' then '4'
end;

alter table order_history
    alter column event type integer using event::integer,
    add constraint chk_history_event check (event in (1,2,3,4));

-- +goose Down
alter table order_history
    drop constraint if exists chk_history_event,
    alter column event type varchar(32) using event::varchar;


update order_history
set event = case
    when event = '1' then 'ACCEPTED'
    when event = '2' then 'ISSUED'
    when event = '3' then 'RETURNED_BY_CLIENT'
    when event = '4' then 'RETURNED_TO_WAREHOUSE'
end;

alter table order_history
    alter column event type varchar(32) using event::varchar,
    add constraint chk_history_event check (event in (
      'ACCEPTED',
      'ISSUED',
      'RETURNED_BY_CLIENT',
      'RETURNED_TO_WAREHOUSE'
    ));