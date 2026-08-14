# syntax = docker/dockerfile:1

FROM caddy:2.11.4@sha256:df7f1c2fb114453b951de51a98efc010db1655a92c2e86be6706714e2417a78d

COPY Caddyfile /etc/caddy/Caddyfile
RUN wget -O - https://github.com/traPtitech/trap-collection-admin/releases/latest/download/dist.tar.gz \
  | tar zxv -C /usr/share/caddy --strip-components=2
