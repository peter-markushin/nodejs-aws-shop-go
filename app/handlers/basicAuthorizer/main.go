package main

import (
	"context"
	"encoding/base64"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/peterm-itr/nodejs-aws-shop-go/config"
)

var appConfig *config.Configuration

func main() {
	var err error
	appConfig, err = config.GetConfig()

	if err != nil {
		log.Println(err.Error())
	}

	lambda.Start(Handler)
}

func Handler(ctx context.Context, event events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	token := event.AuthorizationToken

	if len(token) < 6 || token[:5] != "Basic" {
		log.Println("Not basic auth token")

		return generatePolicy("user", "Deny", event.MethodArn), nil
	}

	decodedToken, err := base64.StdEncoding.DecodeString(token[6:])

	if token == "" || err != nil {
		log.Printf("Token not decoded: %s, %+v", token, err)

		return generatePolicy("user", "Deny", event.MethodArn), nil
	}

	s := strings.SplitN(string(decodedToken), ":", 2)

	if len(s) < 2 {
		log.Printf("Token invalid: %+v", s)

		return generatePolicy("user", "Deny", event.MethodArn), nil
	}

	user, passwd := s[0], s[1]

	if user == appConfig.AppUser && passwd == appConfig.AppPassword {
		return generatePolicy("user", "Allow", event.MethodArn), nil
	}

	return generatePolicy("user", "Deny", event.MethodArn), nil
}

func generatePolicy(principalId, effect, resource string) events.APIGatewayCustomAuthorizerResponse {
	authResponse := events.APIGatewayCustomAuthorizerResponse{PrincipalID: principalId}

	if effect != "" && resource != "" {
		authResponse.PolicyDocument = events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   effect,
					Resource: []string{resource},
				},
			},
		}
	}

	return authResponse
}
