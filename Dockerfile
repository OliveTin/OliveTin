FROM fedora

USER 1001

CMD mkdir -p /config /var/www/olivetin/

CMD dnf install -y iputils && dnf clean all && rm -rf /var/cache/yum # install ping

EXPOSE 1337/tcp 
EXPOSE 1338/tcp 
EXPOSE 1339/tcp

VOLUME /config

COPY OliveTin /usr/bin/OliveTin
COPY webui /var/www/olivetin/

ENTRYPOINT [ "/usr/bin/OliveTin" ]
