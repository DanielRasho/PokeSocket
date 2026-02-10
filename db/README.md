# 🗄️ DB

**I believe databases should be reproducible.** Aside from backups, I'm a fan of migration files for mutating the DB. If you want to spin up a new database anywhere else, just run the migration files and you're ready to go — just like Git!

## Schema

This project uses one database for now. Its schema definitions are stored in [`./game/migrations`](./game/migrations/).

## Quick Start

```bash
# Start the database via Docker Compose (from root)
moon run db:dev
```

Migrations are applied automatically when the container starts for the first time.
