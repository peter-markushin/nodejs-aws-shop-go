package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3notifications"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssnssubscriptions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
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

	notificationEmail := awscdk.NewCfnParameter(stack, jsii.String("NotificationEmail"), &awscdk.CfnParameterProps{
		Type: jsii.String("String"),
	})

	addititonalNotificationEmail := awscdk.NewCfnParameter(stack, jsii.String("AdditionalNotificationEmail"), &awscdk.CfnParameterProps{
		Type: jsii.String("String"),
	})

	s3bucket := awss3.NewBucket(stack, jsii.String("RSAppBucket"), &awss3.BucketProps{
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
		AccessControl:     awss3.BucketAccessControl_BUCKET_OWNER_FULL_CONTROL,
		Cors: &[]*awss3.CorsRule{
			{
				AllowedMethods: &[]awss3.HttpMethods{
					awss3.HttpMethods_PUT,
				},
				AllowedOrigins: jsii.Strings("https://d1xaanpmmg0wvm.cloudfront.net"),
				AllowedHeaders: jsii.Strings("*"),
			},
		},
	})

	awscdk.NewCfnOutput(stack, jsii.String("BeBucket"), &awscdk.CfnOutputProps{
		Value: s3bucket.BucketArn(),
	})

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

	productImportQueue := awssqs.NewQueue(stack, jsii.String("ProductImportQueue"), &awssqs.QueueProps{
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		QueueName:     jsii.String("catalogBatchProcess"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("ProductImportQueueArn"), &awscdk.CfnOutputProps{
		Value: productImportQueue.QueueArn(),
	})

	importNotificationTopic := awssns.NewTopic(stack, jsii.String("Import_Notifications"), &awssns.TopicProps{
		TopicName: jsii.String("createProductTopic"),
	})

	importNotificationTopic.AddSubscription(awssnssubscriptions.NewEmailSubscription(
		notificationEmail.ValueAsString(),
		&awssnssubscriptions.EmailSubscriptionProps{Json: jsii.Bool(false)},
	))

	importNotificationTopic.AddSubscription(awssnssubscriptions.NewEmailSubscription(
		addititonalNotificationEmail.ValueAsString(),
		&awssnssubscriptions.EmailSubscriptionProps{
			Json: jsii.Bool(false),
			FilterPolicy: &map[string]awssns.SubscriptionFilter{
				"price": awssns.SubscriptionFilter_NumericFilter(&awssns.NumericConditions{GreaterThan: jsii.Number(10)}),
			},
		},
	))

	productsHandlerFunc := awslambda.NewFunction(stack, jsii.String("API_Products"), &awslambda.FunctionProps{
		Description: jsii.String("Products API function"),
		Runtime:     awslambda.Runtime_GO_1_X(),
		Handler:     jsii.String("productsHandler"),
		Code:        awslambda.Code_FromAsset(jsii.String("../tmp"), &awss3assets.AssetOptions{}),
		Environment: &map[string]*string{
			"DB_HOST":       dbInstance.InstanceEndpoint().Hostname(),
			"DB_PORT":       jsii.String(strings.Split(*dbInstance.InstanceEndpoint().SocketAddress(), ":")[1]),
			"DB_NAME":       jsii.String(DbName),
			"DB_USER":       jsii.String(DbUser),
			"DB_SSL_MODE":   jsii.String("require"),
			"DB_SECRET_ARN": dbInstance.Secret().SecretFullArn(),
		},
		Vpc:        vpc,
		VpcSubnets: &awsec2.SubnetSelection{SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS},
		SecurityGroups: &[]awsec2.ISecurityGroup{
			lambdaSecurityGroup,
		},
		Timeout: awscdk.Duration_Seconds(jsii.Number(15)),
	})

	dbInstance.Secret().GrantRead(productsHandlerFunc, jsii.Strings("AWSCURRENT"))

	getImportUploadURLFunc := awslambda.NewFunction(stack, jsii.String("API_Products_Import"), &awslambda.FunctionProps{
		Description: jsii.String("Import Products API function"),
		Runtime:     awslambda.Runtime_GO_1_X(),
		Handler:     jsii.String("getImportUploadURL"),
		Code:        awslambda.Code_FromAsset(jsii.String("../tmp"), &awss3assets.AssetOptions{}),
		Environment: &map[string]*string{
			"S3_BUCKET_NAME": s3bucket.BucketName(),
		},
		Vpc:        vpc,
		VpcSubnets: &awsec2.SubnetSelection{SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS},
		SecurityGroups: &[]awsec2.ISecurityGroup{
			lambdaSecurityGroup,
		},
		Timeout: awscdk.Duration_Seconds(jsii.Number(15)),
	})

	s3bucket.GrantReadWrite(getImportUploadURLFunc, "*")

	importFileParserFunc := awslambda.NewFunction(stack, jsii.String("Products_Import_Parser"), &awslambda.FunctionProps{
		Description: jsii.String("Parse imported products csv function"),
		Runtime:     awslambda.Runtime_GO_1_X(),
		Handler:     jsii.String("importFileParser"),
		Code:        awslambda.Code_FromAsset(jsii.String("../tmp"), &awss3assets.AssetOptions{}),
		Environment: &map[string]*string{
			"S3_BUCKET_NAME":   s3bucket.BucketName(),
			"IMPORT_QUEUE_URL": productImportQueue.QueueUrl(),
		},
		Vpc:        vpc,
		VpcSubnets: &awsec2.SubnetSelection{SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS},
		SecurityGroups: &[]awsec2.ISecurityGroup{
			lambdaSecurityGroup,
		},
		Timeout: awscdk.Duration_Seconds(jsii.Number(15)),
	})

	productImportQueue.GrantSendMessages(importFileParserFunc)
	s3bucket.GrantReadWrite(importFileParserFunc, jsii.String("*"))
	s3bucket.AddObjectCreatedNotification(
		awss3notifications.NewLambdaDestination(importFileParserFunc),
		&awss3.NotificationKeyFilter{
			Prefix: jsii.String("uploaded/"),
		},
	)

	catalogBatchProcessFunc := awslambda.NewFunction(stack, jsii.String("Products_Import_From_Queue"), &awslambda.FunctionProps{
		Description: jsii.String("Imported products from queue"),
		Runtime:     awslambda.Runtime_GO_1_X(),
		Handler:     jsii.String("catalogBatchProcess"),
		Code:        awslambda.Code_FromAsset(jsii.String("../tmp"), &awss3assets.AssetOptions{}),
		Environment: &map[string]*string{
			"DB_HOST":                   dbInstance.InstanceEndpoint().Hostname(),
			"DB_PORT":                   jsii.String(strings.Split(*dbInstance.InstanceEndpoint().SocketAddress(), ":")[1]),
			"DB_NAME":                   jsii.String(DbName),
			"DB_USER":                   jsii.String(DbUser),
			"DB_SSL_MODE":               jsii.String("require"),
			"DB_SECRET_ARN":             dbInstance.Secret().SecretFullArn(),
			"IMPORT_NOTIFICATION_TOPIC": importNotificationTopic.TopicArn(),
		},
		Vpc:        vpc,
		VpcSubnets: &awsec2.SubnetSelection{SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS},
		SecurityGroups: &[]awsec2.ISecurityGroup{
			lambdaSecurityGroup,
		},
		Timeout: awscdk.Duration_Seconds(jsii.Number(15)),
	})

	dbInstance.Secret().GrantRead(catalogBatchProcessFunc, jsii.Strings("AWSCURRENT"))
	productImportQueue.GrantConsumeMessages(catalogBatchProcessFunc)
	importNotificationTopic.GrantPublish(catalogBatchProcessFunc)

	productImportEventSource := awslambdaeventsources.NewSqsEventSource(productImportQueue, &awslambdaeventsources.SqsEventSourceProps{
		BatchSize:         jsii.Number(5),
		MaxBatchingWindow: awscdk.Duration_Seconds(jsii.Number(30)),
		Enabled:           jsii.Bool(true),
	})
	catalogBatchProcessFunc.AddEventSource(productImportEventSource)

	apigw := awsapigateway.NewLambdaRestApi(stack, jsii.String("API_GW"), &awsapigateway.LambdaRestApiProps{
		Handler: productsHandlerFunc,
		DefaultCorsPreflightOptions: &awsapigateway.CorsOptions{
			AllowOrigins: awsapigateway.Cors_ALL_ORIGINS(),
			AllowHeaders: awsapigateway.Cors_DEFAULT_HEADERS(),
			AllowMethods: awsapigateway.Cors_ALL_METHODS(),
		},
	})

	importResource := apigw.Root().AddResource(jsii.String("import"), &awsapigateway.ResourceOptions{})
	importResource.AddMethod(
		jsii.String("GET"),
		awsapigateway.NewLambdaIntegration(getImportUploadURLFunc, &awsapigateway.LambdaIntegrationOptions{}),
		&awsapigateway.MethodOptions{
			RequestParameters: &map[string]*bool{"method.request.querystring.name": jsii.Bool(true)},
		},
	)

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

	productsHandlerFunc.Role().AttachInlinePolicy(readDbSecret)
	catalogBatchProcessFunc.Role().AttachInlinePolicy(readDbSecret)

	awscdk.NewCfnOutput(stack, jsii.String("LambdaListProducts_arn"), &awscdk.CfnOutputProps{
		Value: productsHandlerFunc.FunctionArn(),
	})

	awscdk.NewCfnOutput(stack, jsii.String("LambdaImportProducts_arn"), &awscdk.CfnOutputProps{
		Value: getImportUploadURLFunc.FunctionArn(),
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
