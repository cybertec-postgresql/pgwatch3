# Rollout sequence

Pick a storage schema and execute the according "roll_out_*.psql" file via the "psql" CLI tool coming with Postgres,
or manually first rollout the below files and then the chosen schema type's folder contents as superuser.

* "00_schema_base.sql" - schema type and listing of all known "dbname"-s are stored here
* "01_old_metrics_cleanup_procedures.sql" - used to list all unique dbnames and to delete/drop old metrics by the application (can also be used for manual cleanup).

The SQL files by default assume that "pgwatch3" user will owns the schema - to changge run the below command for example:

```bash
find . -name '*.sql' -exec sed -i 's/pgwatch3;/pgwatch3_owner;/g' {} \;
```

# Schema types


## metric-dbname-time

A single top level table for each distinct metric in the "public" schema + 2 levels of subpartitions ("dbname" + weekly time based) in the "subpartitions" schema.
Works on PG 11+ versions. Provides the fastest query runtimes when having long retention intervals / lots of metrics data or slow disks and accessing mostly only a single DB's metrics at a time.
Best used for 50+ monitored DBs.

Also note that when having extremely many hosts under monitoring it might be necessary to increase the "max_locks_per_transaction"
postgresql.conf parameter on the metrics DB for automatic old partition dropping to work. One could of course also drop old
data partitions with some custom script / Cron when increasing "max_locks_per_transaction" is not wanted, and actually this
kind of approach is also working behind the scenes for versions above v1.8.1.

Something like below will be done by the gatherer AUTOMATICALLY:
```sql
create table public."mymetric"
  (LIKE admin.metrics_template)
  PARTITION BY LIST (dbname);
COMMENT ON TABLE public."mymetric" IS 'pgwatch3-generated-metric-lvl';

create table subpartitions."mymetric_mydbname"
  PARTITION OF public."mymetric"
  FOR VALUES IN ('my-dbname') PARTITION BY RANGE (time);
COMMENT ON TABLE subpartitions."mymetric_mydbname" IS 'pgwatch3-generated-metric-dbname-lvl';

create table subpartitions."mymetric_mydbname_y2019w01" -- month calculated dynamically of course
  PARTITION OF subpartitions."mymetric_mydbname"
  FOR VALUES FROM ('2019-01-01') TO ('2019-01-07');
COMMENT ON TABLE subpartitions."mymetric_mydbname_y2019w01" IS 'pgwatch3-generated-metric-dbname-time-lvl';
```

## timescale

Most suitable storage schema when using long retention periods or hundreds of databases due to built-in extra compression.
Typical compression ratios vary from 3 to 10x and also querying of larger historical data sets is typically faster.

Assumes TimescaleDB (v1.7+) extension and "outsources" partition management for normal metrics to the extensions. Realtime
metrics still use the "metric-time" schema as sadly Timescale doesn't support unlogged tables. Additionally one can also
tune the chunking and historic data compression intervals - by default it's 2 days and 1 day. To change use the
admin.timescale_change_chunk_interval() and admin.timescale_change_compress_interval() functions.

Note that if wanting to store a deeper history of 6 months or a year then additionally using [Continous Aggregates](https://docs.timescale.com/latest/using-timescaledb/continuous-aggregates)
might be a good idea. This will though also require modifying the Grafana dashboards, so it's out of scope for pgwatch3.

Something like below will be done by the gatherer AUTOMATICALLY via the `admin.ensure_partition_timescale()` function:
```sql
CREATE TABLE public."some_metric"
  (LIKE admin.metrics_template INCLUDING INDEXES);
COMMENT ON TABLE public."some_metric" IS 'pgwatch3-generated-metric-lvl';

ALTER TABLE some_metric SET (
  timescaledb.compress,
  timescaledb.compress_segmentby = 'dbname'
);

SELECT add_compression_policy('some_metric', INTERVAL '1 day');
```

## metric (DEPRECATED)

A single / separate table for each distinct metric in the "public" schema. No partitioning. Works on all PG versions. Suitable for up to ~25 monitored DBs.

## metric-time (DEPRECATED)

A single top-level table for each distinct metric in the "public" schema + weekly partitions in the "subpartitions" schema.
Works on PG 11+ versions. Suitable for up to ~50 monitored DBs. Reduced IO compared to "metric" as old data partitions will be dropped, not deleted.

Default storage schema for the "pgwatch3" Docker image.

## custom (DEPRECATED)

For cases where the available presets are not satisfactory / applicable. All data inserted into "public.metrics" table and the user is responsible for re-routing with a trigger and possible partition management. In that case all table creations and data cleanup must be performed by the user.

# Data size considerations

When you're planning to monitor lots of databases or with very low intervals, i.e. generating a lot of data, but not selecting
all of it actively (alerting / Grafana) then it would make sense to consider BRIN indexes to save a lot on storage space. See
the according commented out line in the table template definition file.

# Notice on "realtime" metrics

Metrics that have the string 'realtime' in them are handled differently on storage level to draw less resources:

 * They're not normal persistent tables but UNLOGGED tables, meaning they're not WAL-logged and cleared on crash
 * Such subpartitions are dropped after 1d

# Notice on Grafana access to metric data and GRANT-s

For more security sensitive environments where a lot of people have access to metrics you'd want to secure things a bit by
creating a separate read-only user for queries generated by Grafana. And to make sure that this user, here "pgwatch3_grafana",
has access to all current and future tables in the metric DB you'd probably want to execute something like that:
```sql
ALTER DEFAULT PRIVILEGES FOR ROLE pgwatch3 IN SCHEMA public GRANT SELECT ON TABLES TO pgwatch3_grafana;
ALTER DEFAULT PRIVILEGES FOR ROLE pgwatch3 IN SCHEMA subpartitions GRANT SELECT ON TABLES TO pgwatch3_grafana;

GRANT USAGE ON SCHEMA public TO pgwatch3_grafana;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO pgwatch3_grafana;

GRANT USAGE ON SCHEMA admin TO pgwatch3_grafana;
GRANT SELECT ON ALL TABLES IN SCHEMA admin TO pgwatch3_grafana;

GRANT USAGE ON SCHEMA subpartitions TO pgwatch3_grafana;
GRANT SELECT ON ALL TABLES IN SCHEMA subpartitions TO pgwatch3_grafana;
```