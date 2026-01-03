FROM postgres:16 AS builder

# Install build deps
RUN apt-get update && apt-get install -y \
    build-essential git postgresql-server-dev-16 flex bison \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /tmp

# 1. Build Apache AGE (Graph Database)
RUN git clone --branch release/PG16/1.5.0 https://github.com/apache/age.git age \
    && cd age && make install

# 2. Build pgvector (Embeddings)
RUN git clone --branch v0.7.0 https://github.com/pgvector/pgvector.git pgvector \
    && cd pgvector && make && make install

# Final Stage
FROM postgres:16
COPY --from=builder /usr/lib/postgresql/16/lib/age.so /usr/lib/postgresql/16/lib/
COPY --from=builder /usr/share/postgresql/16/extension/age* /usr/share/postgresql/16/extension/
COPY --from=builder /usr/lib/postgresql/16/lib/vector.so /usr/lib/postgresql/16/lib/
COPY --from=builder /usr/share/postgresql/16/extension/vector* /usr/share/postgresql/16/extension/
COPY ./init /docker-entrypoint-initdb.d/