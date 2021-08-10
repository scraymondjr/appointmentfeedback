package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/labstack/echo/v4"

	"github.com/scraymondjr/appointment/datastore/neo4j"
	"github.com/scraymondjr/appointment/http"
)

// Create resources once in init so lambda instance will re-use the values for subsequent requests.

var e *echo.Echo

func init() {
	store := neo4j.New()
	e = http.Echo(store)
}

func main() {
	lambda.Start(Handler)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// If no name is provided in the HTTP request body, throw an error
	return echoadapter.New(e).ProxyWithContext(ctx, req)
}
