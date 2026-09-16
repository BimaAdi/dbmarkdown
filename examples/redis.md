# Redis Example

db=db-redis|name=set-hello
```redis
SET hello world
```

result:
---
```
OK
```
---

db=db-redis|name=set-foo
```redis
SET foo bar 
```

result:
---
```
OK
```
---


db=db-redis|name=get-all
```redis
KEYS *
```

result:
---
```
hello
foo
```
---


db=db-redis|name=get-hello
```redis
GET hello 
```

result:
---
```
world
```
---


