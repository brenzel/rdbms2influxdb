# rdbms2influxdb

This ist a little go application for data transfer from relational database to influxdb.

I created this application, because it's so easy to use grafana with influxdb.

Configure your Influxdb- and RDBMS- Connetion in the app.toml configuration file.

There must be one column in the RDBMS - query called time. If there is no timestamp column in your table then put something like "localtimestamp as time" in the query.
