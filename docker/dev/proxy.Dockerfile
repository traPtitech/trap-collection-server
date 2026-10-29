# syntax = docker/dockerfile:1

FROM caddy:2.11.6@sha256:3422ce6de165df66534f9b9ba50efaf457114ec961763cc52f5dbdaac2972d73

COPY Caddyfile /etc/caddy/Caddyfile
RUN wget -O - https://github.com/traPtitech/trap-collection-admin/releases/latest/download/dist.tar.gz \
  | tar zxv -C /usr/share/caddy --strip-components=2
