insert into orders(
                   id,
                   user_id,
                   status,
                   created_at,
                   expires_at,
                   updated_status_at,
                   package,
                   weight,
                   price)
values (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9
)
ON CONFLICT (id) DO UPDATE SET
user_id            = EXCLUDED.user_id,
status             = EXCLUDED.status,
created_at         = LEAST(orders.created_at, EXCLUDED.created_at),
expires_at         = EXCLUDED.expires_at,
updated_status_at  = EXCLUDED.updated_status_at,
package            = EXCLUDED.package,
weight             = EXCLUDED.weight,
price              = EXCLUDED.price;