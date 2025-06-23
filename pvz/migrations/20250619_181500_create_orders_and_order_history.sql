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