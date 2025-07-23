# news-demo

This is a simple app that collects news from different sources. It consists of two parts:

- rssparser - parses RSS feeds and saves it
- apiserver - gives articles by HTTP

## How to run

First run sharded system. Here it's easy. Run:  `docker compose -f docker-compose-simple.yaml up`
 Then you have got sqpr-router and 2 postgresql16 shards 
Router will will start with config conf/simple-router-2-shards.yaml. Next it will be autoconfigured with conf/init2shards.sql. Here we create distributions with 2 key ranges and attach to it table `articles`.

Next you can run admin console to configure it:
`psql "host=localhost sslmode=allow user=user1 dbname=db1 port=17432"`

or you can use sql run commands in psql using router:
`psql "host=localhost sslmode=allow user=user1 dbname=db1 port=16432"`

First, connect to db1 and create an `articles` table:

```sql
CREATE TABLE articles (
    id SERIAL NOT NULL PRIMARY KEY,    
    url TEXT NOT NULL UNIQUE, 
    title TEXT NOT NULL,
    description TEXT NOT NULL
);
```

Then build and run demo apps:
```bash
make build
```
Run `rssparser` to fill sharded db articles
Run `apiserver` to get loaded articles http://localhost:8080/article/{id}.

