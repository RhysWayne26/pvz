select id,
       user_id,
       status,
       created_at,
       expires_at,
       updated_status_at,
       package,
       weight,
       price
from orders
where id = $1;