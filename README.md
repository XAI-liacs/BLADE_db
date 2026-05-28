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

## Building the database.
* With `pwd` as `.../migrations` run command:
    ```
        goose postgres "<access_key>" up
    ```
    * This will generate the schema and setup the database in you can check the database
    using `\d+` in postgres client.