create table notifications(
    id serial primary key,
    telegram_id bigint,
    message text not null,
    email varchar(255),
    status varchar(50) not null default ('pending'),
    dt timestamp not null
);