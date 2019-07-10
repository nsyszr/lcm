# Setup and run Device Control (Project Barkeeper)

## Run a PostgreSQL database server locally

```
docker run --name dev-postgres -e POSTGRES_PASSWORD=secret -p 5432:5432 -d postgres
```

## Create database

```
docker exec -it dev-postgres psql -U postgres -W postgres
```

```
create database barkeeperdev;
create user u4barkeeperdev with encrypted password 'pw4barkeeperdev';
alter database barkeeperdev owner to u4barkeeperdev;
grant all privileges on database barkeeperdev to u4barkeeperdev;
\q
```

## Create schema

```
docker exec -it dev-postgres psql -U u4barkeeperdev -W barkeeperdev
```

Copy content of file db/migrations/1.sql
