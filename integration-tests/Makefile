default: test-install test-run

test-install:
	npm install --no-fund

test-run:
    # GitHub Actions fails badly on the default timeout of 2000ms
	npx mocha -t 10000

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
