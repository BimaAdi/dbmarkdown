# How to use
Create a database configuration file. By default, Db Markdown looks for `db.json` in the current directory. The configuration name is used by Markdown query blocks to select a connection:

```json
{
	"db-sqlite": "sqlite://todo.db",
	"db-postgres": "postgresql://username:password@localhost:5432/database_name?sslmode=disable"
}
```
then you can use it to Run Markdown or Execute Query Directly
## Run markdown
Add a named SQL query to a Markdown file. The `db` value must match a configuration name, and `name` is the name passed to the CLI (NOTE: don't include \ on markdown. it added to escape the triple backtick):

```markdown
# filename todo.md
## Todo items

db=db-sqlite|name=getalltodo
\```sql
SELECT id, name, is_done FROM todo;
\```

```

Run the named query and write the result back into the same Markdown file:

```sh
dbmd run getalltodo todos.md
```

it will add result on the markdown
```markdown
# filename todo.md
## Todo items

db=db-sqlite|name=getalltodo
\```sql
SELECT id, name, is_done FROM todo;
\```

result:
---
| id | name       | is_done |
|----|------------|---------|
| 1  | first todo | 1       |
---
```

Running the same command again replaces the result between the `---` delimiters. To keep the existing result and add another one, use `--append`:

```sh
dbmd run getalltodo todos.md --append
```

To keep the source file unchanged and write the result to another file, use `--output-file`:

```sh
dbmd run getalltodo todos.md --output-file todos-result.md
```

Use `--output-to shell` to print the result instead of writing Markdown:

```sh
dbmd run getalltodo examples/todos.md --output-to shell
| id | name       | is_done |
|----|------------|---------|
| 1  | first todo | 1       |
```

Use a different configuration file with the global `--conf` option:

```sh
dbmd --conf config/db.json run getalltodo examples/todos.md
```

## Execute Query Directly
Execute a SQL query directly (no markdown needed). The first argument is the configuration name, followed by the query. Results are printed to the shell:

```sh
dbmd exec db-sqlite "SELECT id, name, is_done FROM todo;"
| id | name       | is_done |
|----|------------|---------|
| 1  | first todo | 1       |
```

For queries containing shell-sensitive characters or multiple lines, use single quotes or a shell variable:

```sh
dbmd exec db-postgres 'SELECT id, name FROM todo WHERE is_done = true;'
```

The `exec` command also uses `db.json` by default and accepts `--conf`:

```sh
dbmd --conf config/db.json exec db-sqlite "SELECT COUNT(*) FROM todo;"
```
