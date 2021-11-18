FROM node:16 AS node-build
WORKDIR /build
ADD Makefile /build
# we add what's needed to run install node packages so that dependencies can be cached in a dedicate layer
ADD web/frontend/package.json web/frontend/package-lock.json /build/web/frontend/
RUN make web-deps
ADD web/frontend /build/web/frontend
RUN make web-assets

FROM golang:1.16 as go-build
WORKDIR /build
# we add what's needed to download go modules so that dependencies can be cached in a dedicate layer
ADD go.mod go.sum /build/
RUN go mod download
ADD . /build
COPY --from=node-build /build /build
RUN make build

FROM python:3.7-slim AS trento-runner
RUN ln -s /usr/local/bin/python /usr/bin/python \
    && /usr/bin/python -m venv /venv \
    && /venv/bin/pip install 'ansible~=4.6.0' 'ara~=1.5.7' 'rpm==0.0.2' 'pyparsing~=2.0' \
    && apt-get update && apt-get install -y --no-install-recommends \
      ssh \
    && apt-get purge -y --auto-remove -o APT::AutoRemove::RecommendsImportant=false \
    && rm -rf /var/lib/apt/lists/*

ENV PATH="/venv/bin:$PATH"
ENV PYTHONPATH=/venv/lib/python3.7/site-packages

# Add Tini
ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini

COPY --from=go-build /build/trento /app/trento
LABEL org.opencontainers.image.source="https://github.com/trento-project/trento"
ENTRYPOINT ["/tini", "--", "/app/trento"]

FROM gcr.io/distroless/base:debug AS trento-web
COPY --from=go-build /build/trento /app/trento
LABEL org.opencontainers.image.source="https://github.com/trento-project/trento"
EXPOSE 8080/tcp
ENTRYPOINT ["/app/trento"]
