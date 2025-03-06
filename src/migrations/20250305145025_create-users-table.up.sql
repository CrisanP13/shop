CREATE TABLE users(
	id int primary key auto_increment not null,
	name varchar(64) not null,
	email varchar(64) not null,
	password binary(60) not null
)
