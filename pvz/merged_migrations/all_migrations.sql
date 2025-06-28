-- Source: /mnt/c/Users/danma/GolandProjects/homework/pvz/migrations/20250619_181500_create_orders_and_order_history.sql
-- +goose Up
create type order_status as enum (
    'ACCEPTED',
    'RETURNED',
    'ISSUED',
    'WAREHOUSED'
);

create type event_type as enum (
    'ACCEPTED',
    'ISSUED',
    'RETURNED_BY_CLIENT',
    'RETURNED_TO_WAREHOUSE'
);

create type package_type as enum (
    'none',
    'bag',
    'box',
    'film',
    'bag+film',
    'box+film'
);

create table if not exists orders (
      id bigint primary key,
      user_id bigint not null,
      status order_status not null,
      created_at timestamptz not null default now(),
      expires_at timestamptz not null,
      updated_status_at timestamptz not null default now(),
      package package_type not null,
      weight real not null,
      price real not null
);

create index if not exists idx_orders_user_id ON orders(user_id);

create index if not exists idx_orders_status ON orders(status);

create index if not exists idx_orders_created_at ON orders(created_at);

create index if not exists idx_orders_id_created_at ON orders(id, created_at);

create index if not exists idx_orders_status_updated_status_at_desc on orders(status, updated_status_at desc);

create table if not exists order_history (
     id bigserial primary key,
     order_id bigint not null references orders(id),
     event event_type not null,
     timestamp timestamptz not null default now(),
     constraint uq_order_history_order_ts unique (order_id, timestamp)
);

create index if not exists idx_order_history_order_id_ts on order_history(order_id, timestamp desc);

-- +goose Down
drop table if exists order_history;
drop table if exists orders;

drop type if exists event_type;
drop type if exists order_status;
drop type if exists package_type;

-- Source: /mnt/c/Users/danma/GolandProjects/homework/pvz/migrations/20250623104500_orders_soft_delete_and_drop_order_enums.sql
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




-- Source: /mnt/c/Users/danma/GolandProjects/homework/pvz/migrations/20250623110300_drop_history_event_type_enum.sql
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


-- Source: /mnt/c/Users/danma/GolandProjects/homework/pvz/migrations/20250623112000_alter_indexes_for_soft_delete.sql
-- +goose Up
drop index if exists idx_orders_user_id;
drop index if exists idx_orders_status;
drop index if exists idx_orders_created_at;
drop index if exists idx_orders_id_created_at;
drop index if exists idx_orders_status_updated_status_at_desc;

create index if not exists idx_orders_user_id_active on orders(user_id) where is_deleted = false;
create index if not exists idx_orders_status_active on orders(status) where is_deleted = false;
create index if not exists idx_orders_created_at_active on orders(created_at) where is_deleted = false;
create index if not exists idx_orders_id_created_at_active on orders(id, created_at) where is_deleted = false;
create index if not exists idx_orders_status_updated_status_at_active on orders(status, updated_status_at desc) where is_deleted = false;
create index if not exists idx_orders_id_is_deleted on orders(id, is_deleted);

-- +goose Down
drop index if exists idx_orders_user_id_active;
drop index if exists idx_orders_status_active;
drop index if exists idx_orders_created_at_active;
drop index if exists idx_orders_id_created_at_active;
drop index if exists idx_orders_status_updated_status_at_active;
drop index if exists idx_orders_id_is_deleted;

create index if not exists idx_orders_user_id on orders(user_id);
create index if not exists idx_orders_status on orders(status);
create index if not exists idx_orders_created_at on orders(created_at);
create index if not exists idx_orders_id_created_at on orders(id, created_at);
create index if not exists idx_orders_status_updated_status_at_desc on orders(status, updated_status_at desc);

-- Source: /mnt/c/Users/danma/GolandProjects/homework/pvz/migrations/20250624013900_orders_alter_varchars_to_ints.sql
-- +goose Up

alter table orders drop constraint if exists chk_orders_status;
alter table orders drop constraint if exists chk_orders_package;

update orders set status = case
   when status = 'ACCEPTED' then '1'
   when status = 'RETURNED' then '2'
   when status = 'ISSUED' then '3'
end;

update orders set package = case
    when package = 'none' then '0'
    when package = 'bag' then '1'
    when package = 'box' then '2'
    when package = 'film' then '3'
    when package = 'bag+film' then '4'
    when package = 'box+film' then '5'
end;

alter table orders
    alter column status type integer using status::integer,
    add constraint chk_orders_status check (status in (1, 2, 3));

alter table orders
    alter column package type integer using package::integer,
    add constraint chk_orders_package check (package in (0, 1, 2, 3, 4, 5));

-- +goose Down
alter table orders
    drop constraint if exists chk_orders_status,
    alter column status type varchar(32) using status::varchar;


alter table orders
    drop constraint if exists chk_orders_package,
    alter column package type varchar(32) using package::varchar;

update orders set status = case
   when status = '1' then 'ACCEPTED'
   when status = '2' then 'RETURNED'
   when status = '3' then 'ISSUED'
end;

update orders set package = case
    when package = '0' then 'none'
    when package = '1' then 'bag'
    when package = '2' then 'box'
    when package = '3' then 'film'
    when package = '4' then 'bag+film'
    when package = '5' then 'box+film'
end;

alter table orders
    add constraint chk_orders_status check (status in ('ACCEPTED', 'RETURNED', 'ISSUED'));

alter table orders
    add constraint chk_orders_package check (package in ('none', 'bag', 'box', 'film', 'bag+film', 'box+film'));

-- Source: /mnt/c/Users/danma/GolandProjects/homework/pvz/migrations/20250624015000_history_alter_varchars_to_ints.sql
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

-- Source: /mnt/c/Users/danma/GolandProjects/homework/pvz/migrations/20250627230300_drop_constraints.sql
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

