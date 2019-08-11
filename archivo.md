Para arrancar el backend:
 - `make docker`
 - `docker-compose up -d postgres redis`
 - `docker-compose up server`


Para ejecutar los tests:
 - `docker-compose up -d postgres redis`
 - `make go.test`
