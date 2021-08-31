# Migrations

## How do I write a migration file?

[See here](https://bun.uptrace.dev/guide/migrations.html#migration-names)

As a template, take one of the existing migration files and modify it. It will be automatically loaded and handled by Bun.

## File format

Bun requires a very specific format: 14 digits, then letters or underscores.

You can use the following bash command on your branch to generate a suitable migration filename.

```bash
echo "$(date --utc +%Y%m%H%M%S%N | head -c 14)_$(git rev-parse --abbrev-ref HEAD).go"
```

## Rules of thumb

1. **DON'T DROP TABLES**!!!!!!!!
