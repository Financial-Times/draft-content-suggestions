FROM golang:1 as builder

ENV PROJECT="draft-content-suggestions"
ENV BUILDINFO_PACKAGE="github.com/Financial-Times/service-status-go/buildinfo."

COPY . /${PROJECT}
WORKDIR /${PROJECT}

RUN VERSION="version=$(git describe --tag --always 2> /dev/null)" \
  && DATETIME="dateTime=$(date -u +%Y%m%d%H%M%S)" \
  && REPOSITORY="repository=$(git config --get remote.origin.url)" \
  && REVISION="revision=$(git rev-parse HEAD)" \
  && BUILDER="builder=$(go version)" \
  && LDFLAGS="-X '"${BUILDINFO_PACKAGE}$VERSION"' -X '"${BUILDINFO_PACKAGE}$DATETIME"' -X '"${BUILDINFO_PACKAGE}$REPOSITORY"' -X '"${BUILDINFO_PACKAGE}$REVISION"' -X '"${BUILDINFO_PACKAGE}$BUILDER"'" \
  && echo "Build flags: $LDFLAGS" \
  && CGO_ENABLED=0 go build -mod=readonly -a -o /artifacts/${PROJECT} -ldflags="${LDFLAGS}"

FROM scratch
WORKDIR /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /artifacts/* /
COPY ./_ft/api.yml /
COPY ./config.yml /

CMD [ "/draft-content-suggestions" ]
