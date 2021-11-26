FROM --platform=linux/amd64 docker.io/amd64/fedora

RUN useradd -rm olivetin -u 1000 

RUN mkdir -p /config /var/www/olivetin \
    && \
    dnf install -y \ 
		iputils \
		openssh-clients \
		docker \
    && dnf clean all && \
    rm -rf /var/cache/dnf 

EXPOSE 1337/tcp 

VOLUME /config

COPY OliveTin /usr/bin/OliveTin
COPY webui /var/www/olivetin/

USER olivetin

ENTRYPOINT [ "/usr/bin/OliveTin" ]
