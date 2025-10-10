default: test-install test-run

test-install:
	npm install --no-fund

test-run:
	npx mocha

find-flakey-tests:
	echo "Running test-run infinately"
	sh -c "while make test-run; do :; done"

nginx:
	podman-compose up -d nginx

clean:
	podman-compose down

getsnapshot:
	rm -rf /opt/OliveTin-snapshot/*
	gh run download -D /opt/OliveTin-snapshot/

.PHONY: default
