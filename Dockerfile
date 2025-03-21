FROM --platform=linux/amd64 registry.fedoraproject.org/fedora-minimal:40-x86_64 AS olivetin-tmputils

RUN sudo microdnf -y install dnf-plugins-core
RUN sudo dnf-3 config-manager --add-repo https://download.docker.com/linux/fedora/docker-ce.repo
RUN sudo microdnf install docker-ce-cli docker-compose-plugin

FROM --platform=linux/amd64 registry.fedoraproject.org/fedora-minimal:40-x86_64

LABEL org.opencontainers.image.source https://github.com/OliveTin/OliveTin
LABEL org.opencontainers.image.title OliveTin

RUN mkdir -p /config /config/entities/ /var/www/olivetin \
    && \
	microdnf install -y --nodocs --noplugins --setopt=keepcache=0 --setopt=install_weak_deps=0 \
		iputils \
		openssh-clients \
		shadow-utils \
		apprise \
		jq \
		git \
	&& microdnf clean all

COPY --from=olivetin-tmputils \
	/usr/bin/docker \
	/usr/bin/docker

COPY --from=olivetin-tmputils \
	/usr/libexec/docker/cli-plugins/docker-compose \
	/usr/libexec/docker/cli-plugins/docker-compose

RUN useradd --system --create-home olivetin -u 1000

EXPOSE 1337/tcp

COPY config.yaml /config
COPY var/entities/* /config/entities/
VOLUME /config

COPY OliveTin /usr/bin/OliveTin
COPY webui /var/www/olivetin/
COPY var/helper-actions/* /usr/bin/

USER olivetin

ENTRYPOINT [ "/usr/bin/OliveTin" ]
