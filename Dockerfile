# syntax = docker/dockerfile:1.3-labs
FROM alpine:latest AS build
ARG BUILDPLATFORM
ARG TARGETPLATFORM
ADD ./out /out
RUN <<EOS
set -ex
env > /out/log
echo "I am running on $BUILDPLATFORM, building for $TARGETPLATFORM"
mv /out/calc-hash-${TARGETPLATFORM/\//-} /calc-hash
EOS

FROM scratch
ARG BUILD_VERSION
ARG BUILD_GIT_HEAD
LABEL \
  calc-hash.version="${BUILD_VERSION}" \
  calc-hash.commit="${BUILD_GIT_HEAD}" \
  maintainer="allex_wang <allex.wxn@gmail.com>" \
  description="calc files hash, support md5,sha1,sha256,sha512"
COPY --from=0 /calc-hash /
ENTRYPOINT ["/calc-hash"]
