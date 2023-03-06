FROM ubuntu

ADD client/bin/client-linux /client

ADD server/bin/server-linux /server

COPY runner.sh /scripts/runner.sh

RUN ["chmod", "+x", "/scripts/runner.sh"]

ENTRYPOINT ["/scripts/runner.sh"]