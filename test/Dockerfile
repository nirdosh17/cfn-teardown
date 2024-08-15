FROM ubuntu

ARG GO_VERSION=1.22.0

RUN set -ex \
  && apt-get update \
  && apt-get install wget git python3 python3-pip unzip jq -y

RUN wget -q -o /dev/null "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" \
  && unzip awscli-exe-linux-x86_64.zip -d /awscli > /dev/null \
  && ./awscli/aws/install

WORKDIR /usr/local
RUN wget -q -o /dev/null https://dl.google.com/go/go$GO_VERSION.linux-amd64.tar.gz
RUN tar -xzf go$GO_VERSION.linux-amd64.tar.gz
ENV GOROOT=/usr/local/go
ENV GOPATH=/root/go
ENV PATH=${GOROOT}/bin:${GOPATH}/bin:${PATH}

WORKDIR /app

COPY . /app

CMD ["sh", "-c", "/app/test/test_runner.sh"]
