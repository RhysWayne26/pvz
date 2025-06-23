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