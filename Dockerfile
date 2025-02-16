FROM golang:1.24 as golang_builder
MAINTAINER Jianwei Mao <maojianwei2012@126.com>

WORKDIR /

COPY "./" "/"
RUN /statically_linked_compilation.sh


FROM node:18.17.1 as webui_builder
MAINTAINER Jianwei Mao <maojianwei2012@126.com>

WORKDIR /

COPY "./" "/"
RUN /build_webui.sh


FROM scratch as product
MAINTAINER Jianwei Mao <maojianwei2012@126.com>

WORKDIR /

COPY --from=webui_builder "/resource/" "/resource/"
COPY --from=golang_builder "/MaoServerDiscovery" "/"

EXPOSE 28888 29999 39999

CMD ["/MaoServerDiscovery", "server"]
