.PHONY: docker clean

build: $(shell find . -iname '*.go')
	cd client && make build
	cd server && make build

docker: client/bin/client-linux server/bin/server-linux
	docker image build -t kingkoia06/client-image:v$(version) client/
	docker push kingkoia06/client-image:v$(version)

	docker image build -t kingkoia06/server-image:v$(version) server/
	docker push kingkoia06/server-image:v$(version) 

clean:
	rm -rf bin