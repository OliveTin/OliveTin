FROM --platform=linux/amd64 registry.fedoraproject.org/fedora-minimal:36-x86_64

RUN mkdir -p /config /var/www/olivetin \
    && microdnf install -y --nodocs --noplugins --setopt=keepcache=0 --setopt=install_weak_deps=0 \ 
		iputils \
		openssh-clients \
		shadow-utils \
		docker \
	&& microdnf clean all

RUN useradd --system --create-home olivetin -u 1000 

EXPOSE 1337/tcp 

VOLUME /config

COPY OliveTin /usr/bin/OliveTin
COPY webui /var/www/olivetin/

USER olivetin

ENTRYPOINT [ "/usr/bin/OliveTin" ]
