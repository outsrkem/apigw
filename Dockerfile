# Step 1
FROM alpine:3.19.1
WORKDIR /opt/apigw
ARG APIGW_VERSION
ARG APIGW_REVISION
ARG GO_VERSION="go1.22.0"
ARG LD_PATH="apigw/src/config"
ARG LD_FLAGS="-X $LD_PATH.Version=${APIGW_VERSION} -X $LD_PATH.GoVersion=${GO_VERSION} -X $LD_PATH.GitCommit=${APIGW_REVISION}"
RUN apk add --no-cache --virtual .build-depsgcc libc-dev g++ make

RUN wget https://dl.google.com/go/${GO_VERSION}.linux-amd64.tar.gz
RUN tar xf ${GO_VERSION}.linux-amd64.tar.gz -C /usr/local
ENV PATH=$PATH:/usr/local/go/bin
RUN go version

# Install upx 
RUN wget https://github.com/upx/upx/releases/download/v4.2.3/upx-4.2.3-amd64_linux.tar.xz
RUN tar xf upx-4.2.3-amd64_linux.tar.xz -C /usr/local
ENV PATH=$PATH:/usr/local/upx-4.2.3-amd64_linux
RUN upx -h

COPY . /opt/apigw

# -trimpath 移除源代码中的文件路径信息
# -ldflags -s：不生成符号表 -w：不生成DWARF调试信息
RUN go build -trimpath  -ldflags "-s -w $LD_FLAGS" -o output/apigw src/main/main.go
RUN upx -9 output/apigw
RUN output/apigw -version
RUN cp apigw.yaml output
RUN cp docker-entrypoint.sh output/entrypoint.sh

# Step 2
FROM alpine:3.19.1
ARG APIGW_REVISION
ARG APIGW_VERSION

COPY --from=0 /opt/apigw/output/* /usr/local/bin

ENV APIGW_REVISION=$APIGW_REVISION \
    APIGW_VERSION=$APIGW_VERSION

ENTRYPOINT ["entrypoint.sh"]
