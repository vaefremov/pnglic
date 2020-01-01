OPENAPI_GENERATOR = ~/Tmp/SWAGGER/openapi-generator-cli.jar
JAVA = java

all: test build

build: 
	cd cmd/pnglic ; go build && mv pnglic ../..

validation:
	${JAVA} -jar ${OPENAPI_GENERATOR} validate -i ./pnglic.yaml 

api:
	${JAVA} -jar ${OPENAPI_GENERATOR} generate -i ./pnglic.yaml -g go-gin-server
	go fmt ./go/
	# puts the output into ./go and ./api directories
	# copy all the generated models from ./go to ./openapi
	cp ./go/model_*.go ./pkg/openapi/

test:
	go test ./...
clean:
	rm -rf ./go
	rm -rf ./api
	rm -f pnglic