FROM fedora
RUN dnf install go make -y --exclude=iputils
COPY Makefile ./
RUN make go-tools

