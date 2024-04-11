FROM golang:1.22.2-bookworm
RUN apt-get update && apt-get install -y \
    curl \
    nano \
    jq \
    && apt-get clean
COPY . .
RUN go build -o /go/bin/ezshield
RUN wget https://github.com/anoma/namada/releases/download/v0.32.1/namada-v0.32.1-Linux-x86_64.tar.gz -O namada.tar.gz && tar -xzvf namada.tar.gz && mv namada-v0.32.1-Linux-x86_64/namada* /go/bin/ && rm -rf namada-v0.32.1-Linux-x86_64  namada.tar.gz
RUN wget https://github.com/osmosis-labs/osmosis/releases/download/v23.0.11/osmosisd-23.0.11-linux-amd64 -O /go/bin/osmosisd && chmod u+x /go/bin/osmosisd
ENV TERM=xterm-256color
WORKDIR /root
ENTRYPOINT ["/go/bin/ezshield"]