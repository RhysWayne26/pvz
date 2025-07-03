-- +goose Up
alter table orders
    drop constraint if exists chk_orders_status,
    drop constraint if exists chk_orders_package;

alter table order_history drop constraint if exists chk_history_event;

-- +goose Down
alter table orders
    add constraint chk_orders_status check (status  in (1,2,3)),
    add constraint chk_orders_package check (package in (0,1,2,3,4,5));

alter table order_history
    add constraint chk_history_event check (event   in (1,2,3,4));