FROM golang:1.17 as builder

RUN mkdir /usr/src/ds-switch
WORKDIR /usr/src/ds-switch
COPY . /usr/src/ds-switch/

RUN go mod tidy
RUN make build

FROM alpine:latest
LABEL version=0.1.0

RUN apk update && apk add --no-cache openssh-client bash

COPY --from=builder ./usr/src/ds-switch/bin/ds-switch /
COPY --from=builder ./usr/src/ds-switch/conf/ds.yml /conf/
COPY --from=builder ./usr/src/ds-switch/entrypoint.sh /

ENV DS_SWITCH_MONITOR_INTERVAL 5
ENV DS_SWITCH_EVALUATE_INTERVAL 60
ENV PROMETHEUS_END_POINT "http://localhost:30099"
ENV PROMETHEUS_MONITOR_API_TYPE "GET"
ENV PROMETHEUS_MONITOR_API_METHOD "/-/healthy"
#ENV GRAFANA_BEARER ""
ENV GRAFANA_END_POINT "http://localhost:31947"
ENV GRAFANA_DS_UPDATE_API_TYPE "PUT" 
ENV GRAFANA_DS_UPDATE_API_METHOD "api/datasources" 
ENV GRAFANA_DS_GET_API_TYPE "GET" 
ENV GRAFANA_DS_GET_API_METHOD "api/datasources"
#ENV GRAFANA_DS_NAME_PRIMARY ""
#ENV GRAFANA_DS_NAME_BACKUP ""

ENTRYPOINT ["sh", "/entrypoint.sh"]