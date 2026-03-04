# This file is used for the reproducible research platform codeocean
# hash:sha256:4ca5a4a40e2f74f77e92405086fd650f1685b251736410391946e8b436eea0ed
FROM registry.codeocean.com/codeocean/ubuntu:20.04.2

ARG DEBIAN_FRONTEND=noninteractive

RUN apt update && apt install -y \
    software-properties-common \
    curl && \
    rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://deb.nodesource.com/setup_23.x | bash - && \
    apt install -y nodejs && \
    rm -rf /var/lib/apt/lists/*

ARG TARGETARCH

RUN curl -L https://go.dev/dl/go1.24.2.linux-$TARGETARCH.tar.gz -o golang.tar.gz && \
    tar -xf golang.tar.gz -C /usr/local && \
    rm golang.tar.gz

RUN apt update && \
    apt install -y build-essential zlib1g-dev libncurses5-dev libgdbm-dev \
    libnss3-dev libssl-dev libreadline-dev libffi-dev libsqlite3-dev \
    wget libbz2-dev && \
    cd /tmp && \
    wget https://www.python.org/ftp/python/3.13.0/Python-3.13.0.tgz && \
    tar xzf Python-3.13.0.tgz && \
    cd Python-3.13.0 && \
    ./configure --enable-optimizations --with-ensurepip=install && \
    make -j$(nproc) && \
    make altinstall && \
    cd / && \
    rm -rf /tmp/Python-3.13.0* && \
    apt purge -y build-essential wget && \
    apt autoremove -y && \
    apt clean && \
    rm -rf /var/lib/apt/lists/*

RUN curl -sS https://bootstrap.pypa.io/get-pip.py | /usr/local/bin/python3.13

RUN pip install pipenv

RUN rm $(which python3) && ln -s $(which python3.13) /usr/local/bin/python3

ENV GOBIN "/usr/local/go/bin/"
ENV PATH "$PATH:/usr/local/go/bin/"