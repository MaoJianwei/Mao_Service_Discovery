FROM golang:1.18.3 as builder
MAINTAINER Jianwei Mao <maojianwei2012@126.com>

WORKDIR /

COPY "./" "/"
RUN /statically_linked_compilation.sh

FROM scratch as product
MAINTAINER Jianwei Mao <maojianwei2012@126.com>

WORKDIR /

COPY "resource" "/resource/"
COPY --from=builder "/MaoServerDiscovery" "/"

EXPOSE 28888 29999

CMD ["/MaoServerDiscovery", "server", "--report_server_addr", "::"]
