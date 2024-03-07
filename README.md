# Fullstack Next.js with PostgreSQL & Golang all Dockerized

## Prerequisites

- Docker installed
- GO installed

## How to run

- Clone the repository
- Run `docker compose build` in the root directory
- Run `docker-compose up -d goapp` in the root directory
- Run `docker-compose up -d nextapp` in the `frontend` directory
- Run `docker exec -it db psql -U postgres` in the terminal to access the database
- Type `\l` to list all databases
- Type `\dt` to list all tables
- Type `select * from users;` to see the users table

## How to stop

- Run `docker-compose down` in the root directory
- Type `exit` in the terminal to exit the database
