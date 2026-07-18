define delete-files
	python -c "import shutil;shutil.rmtree('$(1)', ignore_errors=True)"
endef

codestyle:
	npm install
	npx eslint --fix main.js js/* resources/vue
	npx stylelint style.css

clean:
	$(call delete-files,dist)

deps:
	npm install

build:
	npx vite build

dist: deps clean build

.PHONY: codestyle
