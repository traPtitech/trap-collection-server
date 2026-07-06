# syntax = docker/dockerfile:1

FROM caddy:2.11.4@sha256:af5fdcd76f2db5e4e974ee92f96ee8c0fc3edb55bd4ba5032547cbf3f65e486d

COPY Caddyfile /etc/caddy/Caddyfile
RUN wget -O - https://github.com/traPtitech/trap-collection-admin/releases/latest/download/dist.tar.gz \
  | tar zxv -C /usr/share/caddy --strip-components=2
