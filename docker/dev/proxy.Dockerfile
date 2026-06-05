# syntax = docker/dockerfile:1

FROM caddy:2.11.4@sha256:cb9d71ad83182011b79355cd57692686374bd78d6fe327efe0ff8507da03ab13

COPY Caddyfile /etc/caddy/Caddyfile
RUN wget -O - https://github.com/traPtitech/trap-collection-admin/releases/latest/download/dist.tar.gz \
  | tar zxv -C /usr/share/caddy --strip-components=2
