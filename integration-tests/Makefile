default:
	npm install --no-fund
	./node_modules/.bin/mocha

nginx:
	podman-compose up -d nginx

clean:
	podman-compose down

getsnapshot:
	rm -rf /opt/OliveTin-snapshot/*
	gh run download -D /opt/OliveTin-snapshot/

.PHONY: default
