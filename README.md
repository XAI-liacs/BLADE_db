# Blade Database HTTP

* This is a web interface with blade database project that translates the postgresql commands to http commands for easy use as the backend in the BladeDB.

## Database architecture:
![Database](./src/1.jpeg)

## Setting Up PostgreSQL:
* Run the following command:
    * macOS:
        ```bash
            psql postgres
        ```
    * Linux:
        ```bash
            sudo -u postgres psql
        ```
* Create a new database in posgres client represented by:
    ```bash
        postgres=#
    ```
    * To create a new database  use command:
    ```sql
        CREATE DATABASE blade;
    ```
    * Then connect to the new database:
        ```sql
            \c blade
        ```
        * This will update the cli prompt to:
            ```bash
                blade=#
            ```
    * Linux only Set the user password:
        ```sql
            ALTER USER postgres PASSWORD 'postgres';
        ```
* Get your connection string of the format:
    ```
        protocol://<username>:<password>@host:port/database
    ```
    * Let's call this `access_key`.
    * Test it using `psql`, example:
        ```sql
            psql "postgres://username:password@localhost:5432/blade"
        ```
        * It will open a postgres client with blade db active. Exit using `\q` command.

## Building / Destroying the database.
* For managing the schema of the database, we are using `goose`:
* Use command:
  ```bash
    $: go install github.com/pressly/goose/v3/cmd/goose@latest
  ```
* Once installed, there are schema management files available in directory `/sql/schema`, when in that directory use command:
  ```bash
    $: goose postgres "<access_key>" up
  ```
  to build the database in the PostgreSQL, and:

  ```bash
    $: goose postgres "<access_key>" down
  ```

  to undo the build.

## Safe Data Injection.
* For implementing data injection into database, with proper input cleaning, we are using `SQLC`.
* Install it using:
  ```bash
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
  ```
---
