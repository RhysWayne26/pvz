-- +goose Up
alter table orders
    add column is_deleted boolean default false;

update orders
set
    is_deleted = true,
    status = 'RETURNED'
where status = 'WAREHOUSED';

alter table orders
    alter column status type varchar(32) using status::text,
    add constraint chk_orders_status
        check (status in ('ACCEPTED', 'RETURNED', 'ISSUED', 'WAREHOUSED'));

alter table orders
    alter column package type varchar(32),
    ADD CONSTRAINT chk_orders_package
        check (package in ('none', 'bag', 'box', 'film', 'bag+film', 'box+film'));

drop type if exists order_status;
drop type if exists package_type;



-- +goose Down
create type order_status as enum (
    'ACCEPTED',
    'RETURNED',
    'ISSUED',
    'WAREHOUSED'
);

create type package_type as enum (
    'none',
    'bag',
    'box',
    'film',
    'bag+film',
    'box+film'
);

update orders
set
    status = 'WAREHOUSED'
where is_deleted = true;


alter table orders
    drop constraint if exists chk_orders_status,
    drop constraint if exists chk_orders_package,
    alter column status type order_status using status::order_status,
    alter column package type package_type using package::package_type;

alter table orders
    drop column is_deleted;


