# simple-backend

## What it does

This is a simple gallery web application with the following features:
- **User Handling**: Sign up, sign in, sign out, and forgot password
- **Session Handling**: Using cookies
- **Gallery Handling**: Creating, updating, and deleting
- **Image Handling**: Showing, uploading, and deleting

## How it does on high-level

The web app uses the SSR (Server-Side Rendering), so the we app generates the content to browser on the backend side.

The architecture is the MVC (Model View Controller). There is no need to apply more complex architecture like Ports & Adatpers, or Clean Architectures.

## How it does on low-level

The used database is the PostgresSQL.

The MVC implemented this way (pretty obviuos):
- **Model** is under `models/`, this contains the SQL queries. The SQL table definition is under `migrations/`.
- **View** is under `views/`, this contains the content handling. The HTML templates are under `templates/`.
- **Controller** is under `controllers/`.

It is protected againts the CSRF attack.

The password hashing uses `bcrypt`.
The tokens are created by cryptographically secure pseudorandom number generator and hashed by `sha256`.

### SQL migration

SQL migration creates the SQL tables and it can apply changes on the DB based on how the product evolves.

The migration files are under `migrations/`.

The `goose` is a pretty good tool to support SQL migration in go. It has standalone binary, that you can install by executing this:
```
go install github.com/pressly/goose/v3/cmd/goose@v3
```

However, the web app does not require to be installed the `goose`. It is enough just import and use it from web app, so the web app holds all the needed functionalities.

You need to have the installed binary, if you want to add migration file. You can do it with following steps.
This creates the migration file, but the file name contains date that can be an issue:
```
goose create password_reset sql
```

You can change it to the simple version:
```
goose fix
```

Edit the file and that's it.

### Some design decisions

The SQL language is used instead of ORM (Object-Relational Mapping) in SQL queries. I think ORM is good but I wanted to keep the SQL language because it is clear what the query does.

The `log` package is used instead of `fmt` for logging purpose. Simply because the `fmt` is not concurrent safe.

The `custctx` package is created because I think this is a good approach to make sure the data handling in `context.Context` is safe.

The `errors` package is recreated because the `Wrap` method is missing from the standrad lib implementation.

### Live reload during development

I used the `wgo` to make the development easier.

It is here: https://github.com/bokwoon95/wgo

Just run this after installation::
```
wgo run -file=.html ./cmd/app
```

> [!NOTE]
> To make the DB available, run: `docker compose up -d`, and run `docker compose down` in order to uninstall it.

### How to deploy

Docker compose makes sure the web app and DB are deployed and the web app can see the DB. The web app is built by Docker.

```
docker compose -f docker-compose.yaml -f docker-compose.production.yaml up --build
```

This will remove the `app`:
```
docker compose -f docker-compose.yaml -f docker-compose.production.yaml rm app
```
