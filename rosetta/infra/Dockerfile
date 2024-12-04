# Build
FROM golang:1.22.5 AS build

RUN apt update -y && \
    apt install libgmp-dev libssl-dev git -y && \
    apt -y clean all

ENV GOPATH=/root/go
ENV GO111MODULE=on
ENV ITC_PATH=${GOPATH}/src/github.com/intelchain-itc
RUN mkdir -p $ITC_PATH

WORKDIR $ITC_PATH

RUN git clone https://github.com/intelchain-itc/intelchain.git && \
    git clone https://github.com/intelchain-itc/bls.git && \
    git clone https://github.com/intelchain-itc/mcl.git

WORKDIR $ITC_PATH/intelchain

RUN go mod download

RUN go mod tidy

RUN make linux_static && \
    cp ./bin/intelchain /root/intelchain && \
    cp ./rosetta/infra/run.sh /root/run.sh && \
    cp ./rosetta/infra/rclone.conf /root/rclone.conf

RUN cp ./rosetta/infra/intelchain-pstn.conf /root/intelchain-pstn.conf && \
    cp ./rosetta/infra/intelchain-mainnet.conf /root/intelchain-mainnet.conf && \
    cp ./.itc/rosetta_local_fix.csv /root/rosetta_local_fix.csv

# Execution
FROM ubuntu:20.04

RUN apt update -y && \
    apt install libgmp-dev libssl-dev ca-certificates rclone -y && \
    apt -y clean all \

WORKDIR /root

COPY --from=build /root/intelchain /root/intelchain
COPY --from=build /root/run.sh /root/run.sh
COPY --from=build /root/rclone.conf /root/.config/rclone/rclone.conf
COPY --from=build /root/intelchain-pstn.conf /root/intelchain-pstn.conf
COPY --from=build /root/intelchain-mainnet.conf /root/intelchain-mainnet.conf
COPY --from=build /root/rosetta_local_fix.csv /root/rosetta_local_fix.csv

EXPOSE 9500/tcp
EXPOSE 9700/tcp
EXPOSE 9800/tcp
ENTRYPOINT ["/bin/bash","/root/run.sh"]
