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