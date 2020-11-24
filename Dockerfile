FROM golang:1.15-buster

WORKDIR /code/chimera-admission
COPY . .
RUN go build


FROM debian:buster-slim
RUN adduser --uid 2000 chimera

COPY --from=0 /code/chimera-admission/chimera-admission /usr/bin/chimera-admission

ENTRYPOINT ["/usr/bin/chimera-admission"]
EXPOSE 8443
USER chimera
