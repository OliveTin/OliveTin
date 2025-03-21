define delete-files
	python -c "import shutil;shutil.rmtree('$(1)', ignore_errors=True)"
endef

service:
	$(MAKE) -wC service

service-unittests:
	$(MAKE) -wC service unittests

it:
	$(MAKE) -wC integration-tests

go-tools:
	$(MAKE) -wC service go-tools

proto: grpc

grpc: go-tools
	$(MAKE) -wC proto

dist: protoc

protoc:
	protoc --go_out=. --go-grpc_out=. --grpc-gateway_out=. -I .:/usr/include/ OliveTin.proto

podman-image:
	buildah bud -t olivetin

podman-container:
	podman kill olivetin || true
	podman rm olivetin || true
	podman create --name olivetin -p 1337:1337 -v /etc/OliveTin/:/config:ro olivetin
	podman start olivetin

integration-tests-docker-image:
	docker rm -f olivetin && docker rmi -f olivetin
	docker build -t olivetin:latest .
	docker create --name olivetin -p 1337:1337 -v `pwd`/integration-tests/configs/:/config/ olivetin

devrun: compile
	killall OliveTin || true
	./OliveTin &

devcontainer: compile podman-image podman-container

webui-codestyle:
	cd webui.dev && npm install
	cd webui.dev && npx eslint main.js js/*
	cd webui.dev && npx stylelint style.css

webui-dist:
	$(call delete-files,webui)
	$(call delete-files,webui.dev/dist)
	cd webui.dev && npm install
	cd webui.dev && npx parcel build --public-url "."
	python -c "import shutil;shutil.move('webui.dev/dist', 'webui')"
	python -c "import shutil;import glob;[shutil.copy(f, 'webui') for f in glob.glob('webui.dev/*.png')]"

clean:
	$(call delete-files,dist)
	$(call delete-files,OliveTin)
	$(call delete-files,OliveTin.armhf)
	$(call delete-files,OliveTin.exe)
	$(call delete-files,reports)
	$(call delete-files,gen)

.PHONY: grpc proto service
