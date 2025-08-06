FROM node AS web-builder

RUN npm install -g typescript rollup

COPY /src/ /src/

WORKDIR /src/web/
RUN npm install
RUN tsc && rollup -c

FROM golang:alpine AS builder

COPY /src/ /src/
COPY --from=web-builder /src/static/ /src/static/
WORKDIR /src/

ARG TARGETOS TARGETARCH

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /rtkcsm .

FROM golang:alpine AS passwd

RUN echo "nobody:x:65534:65534:Nobody:/:" > /etc_passwd

FROM scratch

COPY --from=passwd /etc_passwd /etc/passwd
COPY --from=builder /rtkcsm /rtkcsm

ENV GIN_MODE=release
WORKDIR /

USER nobody
ENV PATH="/"

ENTRYPOINT ["rtkcsm"]
CMD ["--reader", "suricata", "--transport", "tcp", "--listen", ":9000", "--server", ":8080"]