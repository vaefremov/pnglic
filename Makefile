OPENAPI_GENERATOR = ~/Tmp/SWAGGER/openapi-generator-cli.jar
JAVA = java
validate:
	${JAVA} -jar ${OPENAPI_GENERATOR} validate -i ./pnglic.yaml 

generate:
	${JAVA} -jar ${OPENAPI_GENERATOR} generate -i ./pnglic.yaml -g go-gin-server
	go fmt ./go/
	# puts the output into ./go and ./api directories
	# copy all the generated models from ./go to ./openapi
	cp ./go/model_*.go ./openapi/

clean:
	rm -r ./go
