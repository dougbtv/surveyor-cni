# This Dockerfile is used to build the image available on DockerHub
FROM golang:1.17.1 as build

# Add everything
ADD . /usr/src/surveyor-cni

RUN  cd /usr/src/surveyor-cni && \
     ./hack/build-go.sh

FROM dougbtv/chainsaw-baseimage:latest
LABEL org.opencontainers.image.source https://github.com/dougbtv/surveyor-cni
COPY --from=build /usr/src/surveyor-cni/bin /usr/src/surveyor-cni/bin
COPY --from=build /usr/src/surveyor-cni/LICENSE /usr/src/surveyor-cni/LICENSE
WORKDIR /

ADD ./deployments/entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
