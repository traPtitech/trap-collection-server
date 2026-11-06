# syntax = docker/dockerfile:1

FROM caddy:2.11.7@sha256:36c24ad91a7ff7bc4c87fe22bfa8df3c4b18da0773d73fc12ee2c52b98c0bb7e

COPY Caddyfile /etc/caddy/Caddyfile
RUN wget -O - https://github.com/traPtitech/trap-collection-admin/releases/latest/download/dist.tar.gz \
  | tar zxv -C /usr/share/caddy --strip-components=2
