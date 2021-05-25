FROM fedora

RUN mkdir -p /config /var/www/olivetin/ && \
    dnf install -y iputils && \
    dnf clean all && \
    rm -rf /var/cache/yum # install ping

EXPOSE 1337/tcp 

VOLUME /config

COPY OliveTin /usr/bin/OliveTin
COPY webui /var/www/olivetin/

USER 1001

ENTRYPOINT [ "/usr/bin/OliveTin" ]
