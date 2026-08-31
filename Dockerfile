FROM ubuntu:latest AS build

RUN apt-get update && apt-get install -y git make wget curl sudo unzip golang shellcheck s3cmd util-linux fuse3

COPY --link . /src

WORKDIR /src

ENV CGO_ENABLED=0

RUN make setup
RUN make build

FROM scratch
COPY --from=build /src/tigrisfs /
