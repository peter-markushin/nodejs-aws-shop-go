## [Rolling Scopes School](https://rs.school) AWS Developer 2023Q2 Backend


### Requirements
* [AWS Account](https://aws.amazon.com)
* [Go Lang](https://go.dev/doc/install)
* [AWS CDK](https://docs.aws.amazon.com/cdk/v2/guide/getting_started.html)
* [Docker](https://docs.docker.com/engine/install/) optional, for local development 
* GNU Make - to simplify things a bit

### Deploying Project

* Make sure you have configured access to your aws account with AWS CLI
* Run `cdk bootstrap` command to prepare CDK env in aws
* Run `make deploy email=your.email@example.com` - replace with your email for notifications

### Running the app for development

Good luck with that :)

You can copy `.env.example` into `.env` and adjust variables

Then run `docker compose build && docker compose up`

If you are lucky enough it will run handlers for `/products/...` endpoints at `http://localhost:9990/v1/products`

I didn't figure out a way to run lambdas handling S3 and queue in local environment.

Alternatively you can run `docker compose up db` to run postgres instance and `go run app/handlers/productsHandler/main.go`, but you'll need to configure env variables for it  to work, it doesn't read `.env` file
