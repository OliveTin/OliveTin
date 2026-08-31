define delete-files
	python3 -c "import shutil;shutil.rmtree('$(1)', ignore_errors=True)"
endef

codestyle:
	npm install
	npx eslint --fix main.js js/* resources/vue
	npx stylelint style.css

unittests: deps
	npm test

clean:
	$(call delete-files,dist)

deps:
	npm install

build:
	npx vite build

dist: deps clean build

.PHONY: codestyle unittests
