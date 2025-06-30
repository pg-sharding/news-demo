docker-compose-pg2.yaml - подъём двух нод postgre
debuglocal-newrouter.yaml - поднимаем ноду роутера с подключением к поднятым нодам postgre как к шардам
debuglocal-init2shards.sql - простейший init с одним распределением и присоединением к нему в т.ч. таблицы articles (см. ниже)


RUN - apiserver
# news-demo

This is a simple app that collects news from different sources. It consists of two parts:

- rssparser - parses RSS feeds and saves it
- apiserver - gives articles by HTTP

## How to run

First, create `news` database and `news` user. Connect to it and create an `articles` table:

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
export DATABASE_URL="postgres://news@localhost:5432/news?sslmode=disable&prefer_simple_protocol=true"
make
./rssparser
./apiserver
```

Read the latest news from http://localhost:1323.

## How to shard it

Let's say that we have two PostgreSQL clusters and want to shard the `articles` table. We should run the router, configure it and connect to the router insted of a PostgreSQL cluster directly.

### Run the router

Build SPQR from the [source code](https://github.com/pg-sharding/spqr/tree/1f90d39654b81d4c56e6fd4790adab3ed3be9c3d), call `spqr-router run -c router.yaml` with config:

```yaml
host: 'localhost'
router_port: '6432'
admin_console_port: '7432'
grpc_api_port: '7000'
router_mode: PROXY
show_notice_messages: true
frontend_rules:
  - usr: news
    db: news
    pool_mode: TRANSACTION
    auth_rule:
      auth_method: ok
backend_rules:
  - usr: news
    db: news
    pool_discard: true
    pool_rollback: true
    auth_rule:
      auth_method: md5
      password: password
shards:
  shard01:
    db: news
    usr: news
    pwd: password
    type: DATA
    tls:
      sslmode: "required"
      cert_file: "/Users/denchick/.postgresql/root.crt"
    hosts:
      - 'sas-tcwn5bde6pvo9r0k.db.yandex.net:6432'
  shard02:
    db: news
    usr: news
    pwd: password
    type: DATA
    tls:
      sslmode: "required"
      cert_file: "/Users/denchick/.postgresql/root.crt"
    hosts:
      - 'sas-6puuyjs5sqhcozq9.db.yandex.net:6432'

```

### Configure SPQR

SPQR has an administrative console. This is an app that works by PostgreSQL protocol and you can connect to it by usual `psql`.

```bash
➜  news-demo git:(main) ✗ psql "host=localhost sslmode=disable user=news dbname=news port=7432"

                SQPR router admin console
        Here you can configure your routing rules
------------------------------------------------
        You can find documentation here 
https://github.com/pg-sharding/spqr/tree/master/docs

psql (14.5 (Homebrew), server console)
Type "help" for help.

news=> SHOW shards;
    listing data shards    
---------------------------
 datashard with ID shard01
 datashard with ID shard02
(2 rows)
```

As you can see, SPQR knows about shards but that's not all. To make it work, we should create routing rules: specify a sharding column and key ranges. Let's do that:

```bash
news=> ADD SHARDING RULE rule1 COLUMNS id;
      add sharding rule       
------------------------------
 created sharding column [id]
(1 row)

news=> ADD KEY RANGE krid1 FROM 1 TO 1073741823  ROUTE TO shard01;
             add key range              
----------------------------------------
 created key range from 1 to 1073741823
(1 row)

news=> ADD KEY RANGE krid2 FROM 1073741824 TO 2147483647  ROUTE TO shard02;
                  add key range                  
-------------------------------------------------
 created key range from 1073741824 to 2147483647
(1 row)

news=> SHOW sharding_rules;
          listing sharding rules           
-------------------------------------------
 sharding rule rule1 with column set: [id]
(1 row)

news=>  SHOW key_ranges;
 Key range ID | Shard ID | Lower bound | Upper bound 
--------------+----------+-------------+-------------
 krid1        | shard01  | 1           | 1073741823
 krid2        | shard02  | 1073741824  | 2147483647
(2 rows)
```

### Connect to SPQR router

Now we can connect to proxy a.k.a. router and play with it:

```bash
➜  news-demo git:(main) ✗ psql "host=localhost sslmode=disable user=news dbname=news port=6432"
psql (13.3, server 9.6.22)
Type "help" for help.

news=> CREATE TABLE articles (
    id SERIAL NOT NULL PRIMARY KEY,
    url TEXT NOT NULL UNIQUE, 
    title TEXT NOT NULL,
    description TEXT NOT NULL
);
NOTICE: send query to shard(s) : shard01,shard02
CREATE TABLE
news=> INSERT INTO articles (id, url, title, description) VALUES ('1235', 'https://www.nature.com/articles/d41586-022-01516-2', 'Science needs more research software engineers', 'nope');
NOTICE: send query to shard(s) : shard01
INSERT 0 1
news=> INSERT INTO articles (id, url, title, description) VALUES ('1073741825', 'https://news.ycombinator.com/item?id=29201000', 'Scalable PostgreSQL Connection Pooler (github.com/yandex)', 'nope');
news=> INSERT INTO articles (id, url, title, description) VALUES (3803855397, 'https://tiramisu.bearblog.dev/coffee-gift/', 'An ode to that “coffee friend”', 'Comments');
NOTICE: send query to shard(s) : shard02
INSERT 0 1
```

> NOTICE messages are disabled by default, specify `show_notice_messages` setting in the router config to enable them

You could check now that each shard has only one record:

```bash
news=> SELECT * FROM articles WHERE id > 1;
NOTICE: send query to shard(s) : shard01
  id  |                        url                         |                     title                      | description 
------+----------------------------------------------------+------------------------------------------------+-------------
 1235 | https://www.nature.com/articles/d41586-022-01516-2 | Science needs more research software engineers | nope
(1 row)

news=> SELECT * FROM articles WHERE id < 2147483647;
NOTICE: send query to shard(s) : shard02
     id     |                      url                      |                           title                           | description 
------------+-----------------------------------------------+-----------------------------------------------------------+-------------
 1073741825 | https://news.ycombinator.com/item?id=29201000 | Scalable PostgreSQL Connection Pooler (github.com/yandex) | nope
(1 row)
```

SPQR can handle such queries as `SELECT * FROM table` but we don't recommend using it. This feature is implemented in a non-transactional way.

```bash
news=> SELECT * FROM articles;
NOTICE: send query to shard(s) : shard01,shard02
     id     |                        url                         |                           title                           | description 
------------+----------------------------------------------------+-----------------------------------------------------------+-------------
       1235 | https://www.nature.com/articles/d41586-022-01516-2 | Science needs more research software engineers            | nope
 1073741825 | https://news.ycombinator.com/item?id=29201000      | Scalable PostgreSQL Connection Pooler (github.com/yandex) | nope
(2 rows)
```

## How to run the app again

And that's it! Check out that now everything works: rebuild the parser and the server, run it with `DATABASE_URL="postgres://news@localhost:6432/news?sslmode=disable&prefer_simple_protocol=true"`, play with it.