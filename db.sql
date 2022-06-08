show master status;

show binary logs;

use test;

select * from product;
select * from product2;

drop table if exists product;
drop table if exists product2;

CREATE TABLE product (
    product_name nvarchar(256) NOT NULL PRIMARY KEY,
    maker varchar(256) NOT NULL,
    category nvarchar(256) NOT NULL
);

insert into product values('dior prestige', 'dior', 'makeup');
update product
set maker='dior2'
where product_name='dior prestige';
delete from product where product_name='dior prestige';

insert into product values('ysl lipgloss', 'ysl', 'makeup');
insert into product values('charlie & keith handbag', 'CK', 'bag');
insert into product values('lancome yuex', 'lancome', 'makeup');
update product
set maker='me'
where category='makeup';
delete from product where maker='me';


CREATE TABLE product2 (
    product_name nvarchar(256) NOT NULL PRIMARY KEY,
    maker varchar(256) NOT NULL,
    category nvarchar(256) NOT NULL
);

SHOW BINLOG EVENTS;
SHOW BINLOG EVENTS in 'binlog.000033';
SHOW BINLOG EVENTS in 'binlog.000034';
