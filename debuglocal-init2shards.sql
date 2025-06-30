create distribution ds1 column types integer;

alter distribution ds1 attach relation pgbench_branches distribution key bid;
alter distribution ds1 attach relation test1 distribution key id SCHEMA sh1;

CREATE KEY RANGE FROM 0 ROUTE TO sh1 FOR DISTRIBUTION ds1;

