package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"strings"
)

type CdkStackProps struct {
	awscdk.StackProps
}

func NewCdkStack(scope constructs.Construct, id string, props *CdkStackProps) awscdk.Stack {
	const DbName = "rs_school"
	const DbUser = "postgres"

	var sprops awscdk.StackProps

	if props != nil {
		sprops = props.StackProps
	}

	stack := awscdk.NewStack(scope, &id, &sprops)

	vpc := awsec2.NewVpc(stack, jsii.String("RSAppVPC"), &awsec2.VpcProps{
		VpcName:     jsii.String("rs-app-vpc"),
		MaxAzs:      jsii.Number(3),
		NatGateways: jsii.Number(1),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				Name:       jsii.String("public-1"),
				SubnetType: awsec2.SubnetType_PUBLIC,
				CidrMask:   jsii.Number(24),
			},
			{
				Name:       jsii.String("private-1"),
				SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
				CidrMask:   jsii.Number(24),
			},
		},
	})

	lambdaSecurityGroup := awsec2.NewSecurityGroup(stack, jsii.String("LambdaSecurityGroup"), &awsec2.SecurityGroupProps{
		SecurityGroupName: jsii.String("productsFunc-security-group"),
		Vpc:               vpc,
	})

	dbSecurityGroup := awsec2.NewSecurityGroup(stack, jsii.String("DBSecurityGroup"), &awsec2.SecurityGroupProps{
		SecurityGroupName: jsii.String("db-security-group"),
		Vpc:               vpc,
	})

	dbSecurityGroup.AddIngressRule(
		lambdaSecurityGroup,
		awsec2.Port_Tcp(jsii.Number(5432)),
		jsii.String("ALLOW Lambda to RDS"),
		jsii.Bool(false),
	)

	dbInstance := awsrds.NewDatabaseInstance(stack, jsii.String("RSAppDB"), &awsrds.DatabaseInstanceProps{
		Vpc:          vpc,
		VpcSubnets:   &awsec2.SubnetSelection{SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS},
		Engine:       awsrds.DatabaseInstanceEngine_Postgres(&awsrds.PostgresInstanceEngineProps{Version: awsrds.PostgresEngineVersion_VER_15()}),
		InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_BURSTABLE3, awsec2.InstanceSize_MICRO),
		Credentials: awsrds.Credentials_FromGeneratedSecret(
			jsii.String(DbUser),
			&awsrds.CredentialsBaseOptions{SecretName: jsii.String("DBUserCredentials")},
		),

		MultiAz:                   jsii.Bool(false),
		AllowMajorVersionUpgrade:  jsii.Bool(false),
		AutoMinorVersionUpgrade:   jsii.Bool(true),
		BackupRetention:           awscdk.Duration_Days(jsii.Number(0)),
		DeleteAutomatedBackups:    jsii.Bool(true),
		DeletionProtection:        jsii.Bool(false),
		RemovalPolicy:             awscdk.RemovalPolicy_DESTROY,
		PubliclyAccessible:        jsii.Bool(false),
		EnablePerformanceInsights: jsii.Bool(false),
		DatabaseName:              jsii.String(DbName),
		AllocatedStorage:          jsii.Number(10),
		SecurityGroups: &[]awsec2.ISecurityGroup{
			dbSecurityGroup,
		},
	})

	awscdk.NewCfnOutput(stack, jsii.String("DBEndpoint"), &awscdk.CfnOutputProps{
		Value: dbInstance.InstanceEndpoint().SocketAddress(),
	})

	productsFunc := awslambda.NewFunction(stack, jsii.String("API_Products"), &awslambda.FunctionProps{
		Description: jsii.String("Products API function"),
		Runtime:     awslambda.Runtime_GO_1_X(),
		Handler:     jsii.String("lambdaHandler"),
		Code:        awslambda.Code_FromAsset(jsii.String("../tmp"), &awss3assets.AssetOptions{}),
		Environment: &map[string]*string{
			"DB_HOST":       dbInstance.InstanceEndpoint().Hostname(),
			"DB_PORT":       jsii.String(strings.Split(*dbInstance.InstanceEndpoint().SocketAddress(), ":")[1]),
			"DB_NAME":       jsii.String(DbName),
			"DB_USER":       jsii.String(DbUser),
			"DB_SECRET_ARN": dbInstance.Secret().SecretFullArn(),
		},
		Vpc:        vpc,
		VpcSubnets: &awsec2.SubnetSelection{SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS},
		SecurityGroups: &[]awsec2.ISecurityGroup{
			lambdaSecurityGroup,
		},
		Timeout: awscdk.Duration_Seconds(jsii.Number(15)),
	})

	dbInstance.Secret().GrantRead(productsFunc, jsii.Strings("AWSCURRENT"))

	apigw := awsapigateway.NewLambdaRestApi(stack, jsii.String("API_GW"), &awsapigateway.LambdaRestApiProps{
		Handler: productsFunc,
		DefaultCorsPreflightOptions: &awsapigateway.CorsOptions{
			AllowOrigins: awsapigateway.Cors_ALL_ORIGINS(),
			AllowHeaders: awsapigateway.Cors_DEFAULT_HEADERS(),
			AllowMethods: awsapigateway.Cors_ALL_METHODS(),
		},
	})

	readDbSecretPolicy := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect:  awsiam.Effect_ALLOW,
		Actions: jsii.Strings("secretsmanager:DescribeSecret", "secretsmanager:GetSecretValue"),
		Resources: &[]*string{
			dbInstance.Secret().SecretArn(),
		},
	})

	readDbSecret := awsiam.NewPolicy(stack, jsii.String("LambdaReadDbSecret"), &awsiam.PolicyProps{
		Statements: &[]awsiam.PolicyStatement{readDbSecretPolicy},
	})

	productsFunc.Role().AttachInlinePolicy(readDbSecret)

	awscdk.NewCfnOutput(stack, jsii.String("LambdaListProducts_arn"), &awscdk.CfnOutputProps{
		Value: productsFunc.FunctionArn(),
	})

	awscdk.NewCfnOutput(stack, jsii.String("API_GW_URL"), &awscdk.CfnOutputProps{
		Value: apigw.Url(),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewCdkStack(app, "CdkBEStack", &CdkStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}
