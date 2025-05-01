# Step 1
FROM  outsrkem/alpine:3.19.1-golng1.22.0-v1
WORKDIR /opt/apigw
ARG APIGW_VERSION
ARG APIGW_REVISION
ARG GO_VERSION="go1.22.0"

COPY . /opt/apigw

ARG LD_PATH="apigw/src/config"
ARG LD_FLAGS="-X $LD_PATH.Version=${APIGW_VERSION} -X $LD_PATH.GoVersion=${GO_VERSION} -X $LD_PATH.GitCommit=${APIGW_REVISION}"

RUN go build -trimpath -ldflags "-s -w $LD_FLAGS" -o output/apigw src/main/main.go

RUN output/apigw -version
RUN cp apigw.yaml output
RUN chmod +x docker-entrypoint.sh
RUN cp docker-entrypoint.sh output/entrypoint.sh

# Step 2
FROM alpine:3.19.1
ARG APIGW_REVISION
ARG APIGW_VERSION

COPY --from=0 /opt/apigw/output/ /usr/local/bin
ENV APIGW_REVISION=$APIGW_REVISION APIGW_VERSION=$APIGW_VERSION
ENTRYPOINT ["entrypoint.sh"]
