# mr-weasel

## commands

### ping

Ping pong answer.

### car

Track gas / lease/ tax expenses for different cars.

### holiday

Track holiday / vacation days for different jobs.


### download

Download videos from youtube / vk / twitch clips.

## sqlx cheatsheet

```go
db = sqlx.Open("sqlite3", ":memory:")
err = db.Ping()

// Open and Ping shortcut
db, err = sqlx.Connect("sqlite3", ":memory:")
db = sqlx.MustConnect("sqlite3", ":memory:")

res, err := Exec("INSERT INTO place (country, telcode) VALUES (?, ?)", "Singapore", 65) 
res := MustExec("INSERT INTO place (country, telcode) VALUES (?, ?)", "Singapore", 65)

rows, err := Query("SELECT country, city, telcode FROM place") 
for rows.Next() { err = rows.Scan(&country, &city, &telcode) }
err = rows.Err()

rows, err := Queryx("SELECT country, city, telcode FROM place") 
for rows.Next() { err = rows.Scan(&p) }
err = rows.Err()

err := db.QueryRow("SELECT * FROM place WHERE telcode=?", 50).Scan(&telcode)
err := db.QueryRowx("SELECT city, telcode FROM place LIMIT 1").StructScan(&p)

err = db.Get(&p, "SELECT * FROM place LIMIT 1")
err = db.Select(&pp, "SELECT * FROM place WHERE telcode > ?", 50)

tx, err := db.Begin()
err = tx.Exec(...)
err = tx.Commit()

stmt, err := db.Prepare(`SELECT * FROM place WHERE telcode=?`)
```

