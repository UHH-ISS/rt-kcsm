# syntax=docker/dockerfile:1.7-labs
FROM node AS web-builder

RUN npm install -g typescript rollup

COPY /src/web/package.json /src/web/package.json
COPY /src/web/package-lock.json /src/web/package-lock.json
COPY /src/web/rollup.config.js /src/web/rollup.config.js
COPY /src/web/tsconfig.json /src/web/tsconfig.json

WORKDIR /src/web/
RUN npm install

COPY /src/web/src/ /src/web/src/
RUN tsc && rollup -c

COPY /src/static/ /src/static/

FROM golang:alpine AS builder

COPY --exclude=/src/web/* /src/ /src/
COPY --from=web-builder /src/static/ /src/static/
WORKDIR /src/

ARG TARGETOS TARGETARCH

RUN --mount=type=cache,target=/root/.cache/ GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /rtkcsm .

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