# Sqlite Examples

db=db-sqlite|name=createtable
```sql
CREATE TABLE todo (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    is_done INTEGER
);
```

result:
|  |
||

db=db-sqlite|name=inserttodo
```sql
INSERT INTO todo (name, is_done) VALUES 
('first todo', true),
('second todo', false);
```

result:
|  |
||

db=db-sqlite|name=todoschema
```sql
PRAGMA table_info(todo);
```

result:
| cid | name    | type    | notnull | dflt_value | pk |
|-----|---------|---------|---------|------------|----|
| 0   | id      | INTEGER | 0       |            | 1  |
| 1   | name    | TEXT    | 0       |            | 0  |
| 2   | is_done | INTEGER | 0       |            | 0  |


```sql
SELECT * from todo; 
```

result:
| id | name        | is_done |
|----|-------------|---------|
| 1  | first todo  | 1       |
| 2  | second todo | 0       |
