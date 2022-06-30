FROM scratch
MAINTAINER Jianwei Mao <maojianwei2012@126.com>

WORKDIR /

COPY "resource" "/resource/"
COPY "MaoServerDiscovery" "/"

EXPOSE 28888 29999

CMD ["/MaoServerDiscovery", "server", "--report_server_addr", "::"]
