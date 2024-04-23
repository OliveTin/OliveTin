default: test-install test-run

test-install:
	npm install --no-fund

test-run:
	./node_modules/.bin/mocha -t 10000

nginx:
	podman-compose up -d nginx

clean:
	podman-compose down

getsnapshot:
	rm -rf /opt/OliveTin-snapshot/*
	gh run download -D /opt/OliveTin-snapshot/

.PHONY: default
