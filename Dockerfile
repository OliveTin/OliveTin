FROM fedora

USER 1001

CMD mkdir -p /config /var/www/olivetin/

EXPOSE 1337/tcp 
EXPOSE 1338/tcp 
EXPOSE 1339/tcp

VOLUME /config

COPY OliveTin /usr/bin/OliveTin
COPY webui /var/www/olivetin/

ENTRYPOINT [ "/usr/bin/OliveTin" ]
