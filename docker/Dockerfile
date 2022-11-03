FROM golang:1.18

# For showing Git version via 'pgwatch3 --version'
ARG GIT_HASH
ARG GIT_TIME
ENV GIT_HASH=${GIT_HASH}
ENV GIT_TIME=${GIT_TIME}

ADD src /pgwatch3
RUN cd /pgwatch3 && bash build_gatherer.sh


FROM ubuntu:22.04

RUN apt-get -q update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -qy curl ca-certificates gnupg \
    && echo "deb http://apt.postgresql.org/pub/repos/apt/ jammy-pgdg main" > /etc/apt/sources.list.d/pgdg.list \
    && curl -s https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - && apt-get -q update \
    && DEBIAN_FRONTEND=noninteractive apt-get -qy install postgresql-14 \
       postgresql-plpython3-14 libfontconfig python3-pip libssl-dev libpq-dev \
    && pip3 install -U pip \
    && locale-gen "en_US.UTF-8" && apt autoremove -y \
    && pg_dropcluster 14 main ; pg_createcluster --locale en_US.UTF-8 14 main \
    && echo "include = 'pgwatch_postgresql.conf'" >> /etc/postgresql/14/main/postgresql.conf

RUN curl -o grafana.deb https://dl.grafana.com/oss/release/grafana_8.5.4_amd64.deb \
    && dpkg -i grafana.deb && rm grafana.deb


# Add pgwatch3 sources
ADD src /pgwatch3
# Copy over the compiled gatherer
COPY --from=0 /pgwatch3/pgwatch3 /pgwatch3
ADD webpy /pgwatch3/webpy

# For showing Git versions via :8080/versions or 'pgwatch3 --version'
ARG GIT_HASH
ARG GIT_TIME
ENV GIT_HASH=${GIT_HASH}
ENV GIT_TIME=${GIT_TIME}
RUN echo "${GIT_HASH} ${GIT_TIME}" > /pgwatch3/build_git_version.txt

# Grafana config customizations, Web UI requirements, compilation of the Go gatherer
RUN cp /pgwatch3/bootstrap/grafana_custom_config.ini /etc/grafana/grafana.ini \
    && pip3 install -r /pgwatch3/webpy/requirements_pg_metrics.txt \
    && pip3 install psutil \
    && grafana-cli plugins install savantly-heatmap-panel \
    && pip3 install supervisor pyyaml && mkdir /var/log/supervisor

RUN apt-get -q update && DEBIAN_FRONTEND=noninteractive apt-get -qy install postgresql-14-pg-qualstats

ADD grafana_dashboards /pgwatch3/grafana_dashboards


# Set up supervisord [https://docs.docker.com/engine/admin/using_supervisord/]
COPY supervisord-postgres.conf /etc/supervisor/supervisord.conf

# NB! When security is a concern one should definitely alter "pgwatch3" password in change_pw.sql and maybe modify pg_hba.conf accordingly
COPY postgresql.conf /etc/postgresql/14/main/pgwatch_postgresql.conf
COPY pg_hba.conf /etc/postgresql/14/main/pg_hba.conf
COPY docker-launcher-postgres.sh postgresql.conf pg_hba.conf /pgwatch3/

ENV PW3_DATASTORE postgres
ENV PW3_PG_METRIC_STORE_CONN_STR postgresql://pgwatch3:pgwatch3admin@localhost:5432/pgwatch3_metrics
ENV PW3_PG_SCHEMA_TYPE metric-time
ENV PW3_AES_GCM_KEYPHRASE_FILE /pgwatch3/persistent-config/default-password-encryption-key.txt

# Admin UI for configuring servers to be monitored
EXPOSE 8080
# Gatherer healthcheck port / metric statistics (JSON)
EXPOSE 8081
# Postgres DB holding the pgwatch3 config DB / metrics
EXPOSE 5432
# Grafana UI
EXPOSE 3000
# Prometheus scraping port
EXPOSE 9187

### Volumes for easier updating to newer to newer pgwatch3 containers
### NB! Backwards compatibility is not 100% guaranteed so a backup
### using traditional means is still recommended before updating - see "Updating to a newer Docker version" from README

VOLUME /pgwatch3/persistent-config
VOLUME /var/lib/postgresql
VOLUME /var/lib/grafana

CMD ["/pgwatch3/docker-launcher-postgres.sh"]