default: test-install prep test-run

test-install:
	npm ci --no-fund

codestyle: test-install
	npm run lint

codestyle-fix: test-install
	npm run lint:fix

prep:
ifneq ($(SKIP_WEBUI),1)
	$(MAKE) -wC .. webui-dist
endif
	$(MAKE) -wC ../service compile-currentenv

test-run:
    # GitHub Actions fails badly on the default timeout of 2000ms
	npm test

find-flakey-tests:
	echo "Running test-run infinately"
	sh -c "while make test-run; do :; done"

find-flakey-tests-inf: test-install
	node scripts/find-flakey-tests-inf.mjs

nginx:
	podman-compose up -d nginx

clean:
	podman-compose down

getsnapshot:
	rm -rf /opt/OliveTin-snapshot/*
	gh run download -D /opt/OliveTin-snapshot/

.PHONY: default test-install codestyle codestyle-fix test-run find-flakey-tests find-flakey-tests-inf prep
