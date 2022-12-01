FROM ubuntu:22.04

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update \
	&& apt-get install -y --no-install-recommends \
	ca-certificates \
	debhelper \
	bash \
	apt-utils \
	build-essential \
	sudo \
	dpkg-dev \
	make \
	pkg-config \
	curl \
	libgtk-3-dev \
  libcairo2-dev \
  libglib2.0-dev \
  libvte-2.91-dev \
  libgirepository1.0-dev

RUN update-ca-certificates \
	&& mkdir -p /goroot \
	&& curl -L https://golang.org/dl/go1.19.2.linux-amd64.tar.gz | tar xvzf - -C /goroot --strip-components=1

ENV GOROOT /goroot
ENV GOPATH /go
ENV PATH $GOROOT/bin:$GOPATH/bin:$PATH
ENV GO_EXECUTABLE $GOROOT/bin/go

WORKDIR /build
