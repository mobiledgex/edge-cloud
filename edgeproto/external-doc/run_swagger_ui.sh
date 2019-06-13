#!/bin/sh
docker run -p 80:8080 -e SWAGGER_JSON=/doc/apidocs.swagger.json -v $PWD:/doc swaggerapi/swagger-ui
