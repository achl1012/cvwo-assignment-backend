# CVWO Assignment Backend

## Getting Started

### Installing Go

Download and install Go by following the instructions [here](https://go.dev/doc/install).

### Running the app
1. [Fork](https://docs.github.com/en/get-started/quickstart/fork-a-repo#forking-a-repository) this repo.
2. [Clone](https://docs.github.com/en/get-started/quickstart/fork-a-repo#cloning-your-forked-repository) **your** forked repo.
3. Open your terminal and navigate to the directory containing your cloned project.
4.  **Install dependencies**

    You can install necessary Go packages via:
     ```bash
     go mod tidy
     ```

5. **Set Up PostgreSQL Locally**

   If you don't have PostgreSQL installed, download and install it from [here](https://www.postgresql.org/download/).

   After installing, you can create a local PostgreSQL database and user as follows:

- **Start the PostgreSQL server**  
   Run the following command to start the PostgreSQL server (adjust the command based on your OS and PostgreSQL installation):

   ```bash
   pg_ctl -D /usr/local/var/postgres start  # Example for macOS
   ```

- **Access the PostgreSQL command line**

     Enter the following command:
     ```bash
     psql postgres
     ```
- **Create a new database and user**
  
     Run the following SQL commands to create a new database and user:
     ```sql
     CREATE DATABASE forumflow;
     CREATE USER yourusername WITH PASSWORD 'yourpassword';
     ALTER ROLE yourusername SET client_encoding TO 'utf8';
     ALTER ROLE yourusername SET default_transaction_isolation TO 'read committed';
     ALTER ROLE yourusername SET timezone TO 'UTC';
     GRANT ALL PRIVILEGES ON DATABASE forumflow TO yourusername;
     ```

6. **Configure `.env` file**

   Create a .env file in the root of the backend folder if it doesn't already exist.

   Set the DATABASE_URL variable to your local database credentials:
   ```.env
   DATABASE_URL=postgres://yourusername:yourpassword@localhost:5432/forumflow?sslmode=disable
   ```

   Note: Replace yourusername and yourpassword with the credentials you used to set up the database. If you're using a different host or port, make sure to adjust the URL accordingly.
   
7. **Run the backend**
   
   Start the Go backend server:
   ```bash
   go run cmd/server/main.go
   ```
   
   Your backend should now be running at http://localhost:10000
