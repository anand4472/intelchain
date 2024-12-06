FROM ubuntu:18.04

ARG TARGETARCH
ARG GOLANG_VERSION="1.19"

SHELL ["/bin/bash", "-c"]

ENV GOPATH=/root/go
ENV GO111MODULE=on
ENV ITC_PATH=${GOPATH}/src/github.com/intelchain-itc
ENV OPENSSL_DIR=/usr/lib/ssl
ENV MCL_DIR=${ITC_PATH}/mcl
ENV BLS_DIR=${ITC_PATH}/bls
ENV CGO_CFLAGS="-I${BLS_DIR}/include -I${MCL_DIR}/include"
ENV CGO_LDFLAGS="-L${BLS_DIR}/lib"
ENV LD_LIBRARY_PATH=${BLS_DIR}/lib:${MCL_DIR}/lib
ENV GIMME_GO_VERSION=${GOLANG_VERSION}
ENV PATH="/root/bin:${PATH}"

RUN apt update && apt upgrade -y && \
	apt install libgmp-dev libssl-dev curl git \
	psmisc dnsutils jq make gcc g++ bash tig tree sudo vim \
	silversearcher-ag unzip emacs-nox nano bash-completion -y

RUN mkdir ~/bin && \
	curl -sL -o ~/bin/gimme \
	https://raw.githubusercontent.com/travis-ci/gimme/master/gimme && \
	chmod +x ~/bin/gimme

RUN eval "$(~/bin/gimme ${GIMME_GO_VERSION})"

RUN git clone https://github.com/intelchain-itc/intelchain.git ${ITC_PATH}/intelchain

RUN git clone https://github.com/intelchain-itc/bls.git ${ITC_PATH}/bls

RUN git clone https://github.com/intelchain-itc/mcl.git ${ITC_PATH}/mcl

RUN git clone https://github.com/intelchain-itc/go-sdk.git ${ITC_PATH}/go-sdk

RUN cd ${ITC_PATH}/bls && make -j8 BLS_SWAP_G=1

RUN touch /root/.bash_profile && \
	gimme ${GIMME_GO_VERSION} >> /root/.bash_profile && \
	echo "GIMME_GO_VERSION='${GIMME_GO_VERSION}'" >> /root/.bash_profile && \
	echo "GO111MODULE='on'" >> /root/.bash_profile && \
	echo ". ~/.bash_profile" >> /root/.profile && \
	echo ". ~/.bash_profile" >> /root/.bashrc

ENV PATH="/root/.gimme/versions/go${GIMME_GO_VERSION}.linux.${TARGETARCH:-amd64}/bin:${GOPATH}/bin:${PATH}"

RUN . ~/.bash_profile; \
	go install golang.org/x/tools/cmd/goimports; \
	go install golang.org/x/lint/golint ; \
	go install github.com/rogpeppe/godef ; \
	go install github.com/go-delve/delve/cmd/dlv; \
	go install github.com/golang/mock/mockgen; \
	go install github.com/stamblerre/gocode; \
	go install golang.org/x/tools/...; \
	go install honnef.co/go/tools/cmd/staticcheck/...

WORKDIR ${ITC_PATH}/intelchain

RUN scripts/install_build_tools.sh

RUN go mod tidy

RUN scripts/go_executable_build.sh -S

RUN cd ${ITC_PATH}/go-sdk && make -j8 && cp itc /root/bin

ARG K1=
ARG K2=
ARG K3=

ARG KS1=
ARG KS2=
ARG KS3=

RUN itc keys import-private-key ${KS1} && \
	itc keys import-private-key ${KS2} && \
	itc keys import-private-key ${KS3} && \
	itc keys generate-bls-key > keys.json

RUN jq  '.["encrypted-private-key-path"]' -r keys.json > /root/keypath && cp keys.json /root && \
	echo "export BLS_KEY_PATH=$(cat /root/keypath)" >> /root/.bashrc && \
	echo "export BLS_KEY=$(jq '.["public-key"]' -r keys.json)" >> /root/.bashrc && \
	echo "printf '${K1}, ${K2}, ${K3} are imported accounts in itc for local dev\n\n'" >> /root/.bashrc && \
	echo "printf 'test with: itc blockchain validator information ${K1}\n\n'" >> /root/.bashrc && \
	echo "echo "$(jq '.["public-key"]' -r keys.json)" is an extern bls key" >> /root/.bashrc && \
	echo ". /etc/bash_completion" >> /root/.bashrc && \
	echo ". <(itc completion)" >> /root/.bashrc
