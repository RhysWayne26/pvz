update orders
set status = 'WAREHOUSED'
where id = $1;