ARG GOPATH_DEFAULT=/root/go
ARG SRC_PATH=src/github.com/zennittians
ARG BUILD_BRANCH=main

FROM golang:1.22.5-bullseye as builder

ARG GOPATH_DEFAULT
ARG SRC_PATH
ARG BUILD_BRANCH

ENV GOPATH=${GOPATH_DEFAULT}
ENV GO111MODULE=on
ENV ITC_PATH=${GOPATH}/${SRC_PATH}
ENV BRANCH=${BUILD_BRANCH}

ENV PACKAGES libgmp-dev libssl-dev curl git \
	psmisc dnsutils jq make gcc g++ bash tig tree

# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y $PACKAGES --no-install-recommends

WORKDIR ${ITC_PATH}

RUN git clone --branch ${BRANCH} https://github.com/zennittians/intelchain.git intelchain \
  && git clone https://github.com/zennittians/bls.git bls \
  && git clone https://github.com/zennittians/mcl.git mcl \
  && git clone https://github.com/zennittians/go-sdk.git go-sdk

WORKDIR ${ITC_PATH}/intelchain
RUN make linux_static

# Pull intelchain into a second stage deploy alpine container
FROM alpine:3.16.0

ARG GOPATH_DEFAULT
ARG SRC_PATH

ARG INTELCHAIN_USER=intelchain
ARG INTELCHAIN_USER_UID=1000
ARG INTELCHAIN_USER_GID=1000

ENV INTELCHAIN_HOME=/intelchain
ENV HOME=${INTELCHAIN_HOME}

# hadolint ignore=DL3018
RUN apk add --no-cache bash bind-tools tini curl sed \
    && rm -rf /var/cache/apk/* \
    && addgroup -g ${INTELCHAIN_USER_GID} ${INTELCHAIN_USER} \
    && adduser -u ${INTELCHAIN_USER_UID} -G ${INTELCHAIN_USER} --shell /sbin/nologin --no-create-home -D ${INTELCHAIN_USER} \
    && addgroup ${INTELCHAIN_USER} tty \
    && sed -i -e "s/bin\/sh/bin\/bash/" /etc/passwd

RUN echo "[ ! -z \"\$TERM\" -a -r /etc/motd ] && cat /etc/motd" >> /etc/bash/bashrc

WORKDIR ${INTELCHAIN_HOME}

COPY --from=builder ${GOPATH_DEFAULT}/${SRC_PATH}/intelchain/bin/intelchain /usr/local/bin/

RUN chmod +x  /usr/local/bin/intelchain  \
    && mkdir -p /data \
    && chown -R ${INTELCHAIN_USER_UID}:${INTELCHAIN_USER_GID} ${INTELCHAIN_HOME} /data

VOLUME /data

USER ${INTELCHAIN_USER_UID}:${INTELCHAIN_USER_GID}

# default base port, rpc port and rest port
EXPOSE 9000/tcp 9500/tcp 9800/tcp 6000/tcp

ENTRYPOINT ["tini", "--"]
CMD ["intelchain", "-c", "intelchain.conf"]
