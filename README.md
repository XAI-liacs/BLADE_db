# Blade Database

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
        protocol://username:password@host:port/database
    ```
    * Let's caled this `access_key`.
    * Test it using `psql`, example:
        ```sql
            psql "postgres://username:@localhost:5432/blade"
        ```
        * It will open a postgres client with blade db active. Exit using `\q` command.

## Building / Destroying the database.
* In file alembic.ini, update the `sqlalchemy.url` on line 89, as 
    ```
    postgresql+psycopg://username:password@host:port/database
    ```
    * Then to update the schema to latest version, run:
        ```ssh
            alembic upgrade head
        ```
    * And to remove the database you can run:
        ```ssh
            alembic downgrade base
        ```
---