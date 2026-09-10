define delete-files
	python3 -c "import shutil;shutil.rmtree('$(1)', ignore_errors=True)"
endef

codestyle: deps
	npm run lint

codestyle-fix: deps
	npm run lint:fix

unittests: deps
	npm test

clean:
	$(call delete-files,dist)

deps:
	npm ci

build:
	npm run build

dist: deps clean build

.PHONY: codestyle codestyle-fix unittests deps build dist clean
