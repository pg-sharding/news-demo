# news-demo

This is a simple app that collects news from different sources. It consists of two parts:

- rssparser - parses RSS feeds and saves it
- apiserver - gives articles by HTTP

## How to run

First, create an article table. Connect to your database and create an `articles` table:

```sql
CREATE TABLE articles (
    id SERIAL NOT NULL PRIMARY KEY,
    url TEXT NOT NULL UNIQUE, 
    title TEXT NOT NULL,
    description TEXT NOT NULL
);
```

Then build and run the app:

```bash
export DATABASE_URL="postgres://postgres@localhost:5432/postgres?sslmode=disable&prefer_simple_protocol=true"
make
./rssparser
./apiserver
```

Read the latest news from http://localhost:1323.

## How to shard it

Let's say that we have two PostgreSQL clusters and want to shard the `articles` table. Briefly, we should run spqr, configure it and modify a bit our app.

### Run router

Build SPQR from the [source code](https://github.com/pg-sharding/spqr/tree/1f90d39654b81d4c56e6fd4790adab3ed3be9c3d), call `spqr-rr run -c cfg.yaml` with config:

```yaml
addr: '[::1]:6432'
adm_addr: '[::1]:7432'
proto: tcp6
http_addr: '[::1]:7001'
log_level: DEBUG5
qrouter:
  qrouter_type: PROXY
rules:
  frontend_rules:
    - route_key_cfg:
        usr: username
        db: dbname
      pooling_mode: TRANSACTION
      auth_rule:
        auth_method: ok
  proto: tcp6
  world_shard_fallback: true
  shard_mapping:
    shard1:
      conn_db: dbname
      conn_usr: username
      passwd: password
      shard_type: DATA
      hosts:
        - conn_addr: 'host1:5432'
    shard2:
      conn_db: dbname
      conn_usr: username
      passwd: '123456'
      shard_type: DATA
      hosts:
        - conn_addr: 'host2:5432'
  backend_rules:
    - route_key_cfg:
        usr: username
        db: dbname
      pool_discard: true
      pool_rollback: true
```

### Configure SPQR

SPQR has an administrative console. This is an app that works by PostgreSQL protocol and you can connect to it by usual `psql`.

```bash
➜  news-demo git:(main) ✗ psql "host=localhost sslmode=disable user=username dbname=dbname port=7432"

                SQPR router admin console
        Here you can configure your routing rules
------------------------------------------------
        You can find documentation here 
https://github.com/pg-sharding/spqr/tree/master/doc/router

psql (13.3, server console)
Type "help" for help.

dbname=?> SHOW SHARDS;
                                                         listing data shards                                                         
-------------------------------------------------------------------------------------------------------------------------------------
 datashard with ID &{sh1 %!s(*config.ShardCfg=&{[0xc000140be0] dbname username password DATA {  } <nil>})}
 datashard with ID &{sh2 %!s(*config.ShardCfg=&{[0xc000140c00] dbname username password DATA {  } <nil>})}
(2 rows)
```

As you can see, SPQR knows about shards but that's not all. To make it work, we should create routing rules: specify a sharding column and key ranges. Let's do that:

```bash
dbname=> CREATE SHARDING COLUMN id;
      add sharding rule       
------------------------------
 created sharding column [id]
(1 row)

dbname=> ADD KEY RANGE 1 1073741823 sh1 krid1;
             add key range              
----------------------------------------
 created key range from 1 to 1073741823
(1 row)

dbname=> ADD KEY RANGE 1073741824 2147483647 sh2 krid2;
                  add key range                  
-------------------------------------------------
 created key range from 1073741824 to 2147483647
(1 row)

dbname=> show key_ranges;
 Key range ID | Shard ID | Lower bound | Upper bound 
--------------+----------+-------------+-------------
 krid2        | sh2      | 1073741824  | 2147483647
 krid1        | sh1      | 1           | 1073741823
(2 rows)

```

### Connect to SPQR router

Now we can connect to proxy a.k.a. router and play with it:

```bash
➜  news-demo git:(main) ✗ psql "host=localhost sslmode=disable user=username dbname=dbname port=6432"
psql (13.3, server 9.6.22)
Type "help" for help.

dbname=> CREATE TABLE articles (
dbname(>     id SERIAL NOT NULL PRIMARY KEY,
dbname(>     url TEXT NOT NULL UNIQUE, 
dbname(>     title TEXT NOT NULL,
dbname(>     description TEXT NOT NULL
dbname(> );
CREATE TABLE
dbname=> INSERT INTO articles (id, url, title, description) VALUES ('1235', 'https://www.nature.com/articles/d41586-022-01516-2', 'Science needs more research software engineers', 'nope');
INSERT 0 1
dbname=> INSERT INTO articles (id, url, title, description) VALUES ('2147483644', 'https://www.nature.com/articles/d41586-022-01516-2', 'Science needs more research software engineers', 'nope');
INSERT 0 1
```

You could check now that each shard has only one record:

```bash
opgejrqr=> select * from articles where id > 1;
  id  |                        url                         |                     title                      | description 
------+----------------------------------------------------+------------------------------------------------+-------------
 1235 | https://www.nature.com/articles/d41586-022-01516-2 | Science needs more research software engineers | nope
(1 row)

opgejrqr=> select * from articles where id < 2147483647;
     id     |                        url                         |                     title                      | description 
------------+----------------------------------------------------+------------------------------------------------+-------------
 2147483644 | https://www.nature.com/articles/d41586-022-01516-2 | Science needs more research software engineers | nope
(1 row)
```

### Modify and run the app again

Unfortunately, spqr router does not know handle such queries as `SELECT * FROM articles` (at least for now), so we should modify `GetAll()` method in repo/repo.go:

```golang
func (repo *ArticlesRepository) GetAll() ([]*Article, error) {
	articles := []*Article{}
	if err := repo.scan(&articles, 1); err != nil {
		return nil, err
	}
	if err := repo.scan(&articles, 1073741824); err != nil {
		return nil, err
	}

	return articles, nil
}

func (repo *ArticlesRepository) scan(articles *[]*Article, id int) error {
	rows, err := repo.pool.Query(context.Background(), "SELECT * FROM articles WHERE id > $1", id)
	if err != nil {
		return fmt.Errorf("unable to SELECT: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		a := Article{}
		err := rows.Scan(&a.ID, &a.URL, &a.Title, &a.Description)
		if err != nil {
			return err
		}
		*articles = append(*articles, &a)
	}
	return nil
}
```

And that's it! Check out that now everything works: rebuild parser and server, run it, play with it.