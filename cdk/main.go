package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticache"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53targets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type JwtAppStackProps struct {
	awscdk.StackProps
}

func NewJwtAppStack(scope constructs.Construct, id string, props *JwtAppStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Create VPC for Lambda, RDS, and ElastiCache
	vpc := awsec2.NewVpc(stack, jsii.String("JwtAppVPC"), &awsec2.VpcProps{
		MaxAzs: jsii.Number(2),
		NatGateways: jsii.Number(1),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				Name:       jsii.String("Public"),
				SubnetType: awsec2.SubnetType_PUBLIC,
				CidrMask:   jsii.Number(24),
			},
			{
				Name:       jsii.String("Private"),
				SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
				CidrMask:   jsii.Number(24),
			},
			{
				Name:       jsii.String("Isolated"),
				SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
				CidrMask:   jsii.Number(24),
			},
		},
	})

	// Security group for Lambda
	lambdaSG := awsec2.NewSecurityGroup(stack, jsii.String("LambdaSecurityGroup"), &awsec2.SecurityGroupProps{
		Vpc:              vpc,
		Description:      jsii.String("Security group for JWT App Lambda function"),
		AllowAllOutbound: jsii.Bool(true),
	})

	// Security group for RDS
	rdsSG := awsec2.NewSecurityGroup(stack, jsii.String("RDSSecurityGroup"), &awsec2.SecurityGroupProps{
		Vpc:              vpc,
		Description:      jsii.String("Security group for JWT App RDS database"),
		AllowAllOutbound: jsii.Bool(true),
	})
	rdsSG.AddIngressRule(lambdaSG, awsec2.Port_Tcp(jsii.Number(5432)), jsii.String("Allow Lambda to access RDS"), jsii.Bool(false))

	// Security group for ElastiCache
	redisSG := awsec2.NewSecurityGroup(stack, jsii.String("RedisSecurityGroup"), &awsec2.SecurityGroupProps{
		Vpc:              vpc,
		Description:      jsii.String("Security group for JWT App Redis cluster"),
		AllowAllOutbound: jsii.Bool(true),
	})
	redisSG.AddIngressRule(lambdaSG, awsec2.Port_Tcp(jsii.Number(6379)), jsii.String("Allow Lambda to access Redis"), jsii.Bool(false))

	// Create RDS PostgreSQL database
	dbSecret := awssecretsmanager.NewSecret(stack, jsii.String("DBSecret"), &awssecretsmanager.SecretProps{
		SecretName: jsii.String("jwtapp-db-credentials"),
		GenerateSecretString: &awssecretsmanager.SecretStringGenerator{
			SecretStringTemplate: jsii.String(`{"username":"jwtappuser"}`),
			GenerateStringKey:    jsii.String("password"),
			ExcludeCharacters:    jsii.String(`"@/\`),
			PasswordLength:       jsii.Number(32),
		},
	})

	dbInstance := awsrds.NewDatabaseInstance(stack, jsii.String("JwtAppDatabase"), &awsrds.DatabaseInstanceProps{
		Engine: awsrds.DatabaseInstanceEngine_Postgres(&awsrds.PostgresInstanceEngineProps{
			Version: awsrds.PostgresEngineVersion_VER_15(),
		}),
		InstanceType:     awsec2.InstanceType_Of(awsec2.InstanceClass_BURSTABLE3, awsec2.InstanceSize_SMALL),
		Vpc:              vpc,
		VpcSubnets:       &awsec2.SubnetSelection{SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED},
		SecurityGroups:   &[]awsec2.ISecurityGroup{rdsSG},
		DatabaseName:     jsii.String("jwtappdb"),
		Credentials:      awsrds.Credentials_FromSecret(dbSecret, jsii.String("jwtappuser")),
		AllocatedStorage: jsii.Number(20),
		MaxAllocatedStorage: jsii.Number(100),
		DeletionProtection:  jsii.Bool(false),
		BackupRetention:     awscdk.Duration_Days(jsii.Number(7)),
		RemovalPolicy:       awscdk.RemovalPolicy_SNAPSHOT,
	})

	// Create ElastiCache Redis cluster
	redisSubnetGroup := awselasticache.NewCfnSubnetGroup(stack, jsii.String("RedisSubnetGroup"), &awselasticache.CfnSubnetGroupProps{
		Description: jsii.String("Subnet group for JWT App Redis cluster"),
		SubnetIds:   vpc.SelectSubnets(&awsec2.SubnetSelection{SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED}).SubnetIds,
		CacheSubnetGroupName: jsii.String("jwtapp-redis-subnet-group"),
	})

	redisCluster := awselasticache.NewCfnCacheCluster(stack, jsii.String("RedisCluster"), &awselasticache.CfnCacheClusterProps{
		CacheNodeType:          jsii.String("cache.t3.micro"),
		Engine:                 jsii.String("redis"),
		NumCacheNodes:          jsii.Number(1),
		VpcSecurityGroupIds:    &[]*string{redisSG.SecurityGroupId()},
		CacheSubnetGroupName:   redisSubnetGroup.CacheSubnetGroupName(),
		ClusterName:            jsii.String("jwtapp-redis"),
		EngineVersion:          jsii.String("7.0"),
		AutoMinorVersionUpgrade: jsii.Bool(true),
	})
	redisCluster.AddDependency(redisSubnetGroup)

	// Create JWT secret
	jwtSecret := awssecretsmanager.NewSecret(stack, jsii.String("JWTSecret"), &awssecretsmanager.SecretProps{
		SecretName: jsii.String("jwtapp-jwt-secret"),
		GenerateSecretString: &awssecretsmanager.SecretStringGenerator{
			PasswordLength:    jsii.Number(64),
			ExcludeCharacters: jsii.String(`"@/\`),
		},
	})

	// Create Lambda function from the compiled bootstrap binary
	lambdaCode := awslambda.Code_FromAsset(jsii.String("../"), &awss3assets.AssetOptions{
		Bundling: &awscdk.BundlingOptions{
			Image: awslambda.Runtime_PROVIDED_AL2023().BundlingImage(),
			Command: &[]*string{
				jsii.String("bash"),
				jsii.String("-c"),
				jsii.String("GOOS=linux GOARCH=amd64 go build -o /asset-output/bootstrap cmd/lambda/main.go"),
			},
			User: jsii.String("root"),
		},
	})

	lambdaFunction := awslambda.NewFunction(stack, jsii.String("JwtAppFunction"), &awslambda.FunctionProps{
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Handler:      jsii.String("bootstrap"),
		Code:         lambdaCode,
		Vpc:          vpc,
		VpcSubnets:   &awsec2.SubnetSelection{SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS},
		SecurityGroups: &[]awsec2.ISecurityGroup{lambdaSG},
		Timeout:      awscdk.Duration_Seconds(jsii.Number(30)),
		MemorySize:   jsii.Number(512),
		Environment: &map[string]*string{
			"APP_PORT":       jsii.String("8080"),
			"DB_HOST":        dbInstance.DbInstanceEndpointAddress(),
			"DB_PORT":        dbInstance.DbInstanceEndpointPort(),
			"DB_NAME":        jsii.String("jwtappdb"),
			"DB_USER":        jsii.String("jwtappuser"),
			"DB_DIALECT":     jsii.String("postgres"),
			"REDIS_HOST":     redisCluster.AttrRedisEndpointAddress(),
			"REDIS_PORT":     redisCluster.AttrRedisEndpointPort(),
			"REDIS_PASSWORD": jsii.String(""),
			"DB_PASSWORD_SECRET_ARN": dbSecret.SecretArn(),
			"JWT_SECRET_ARN": jwtSecret.SecretArn(),
		},
	})

	// Grant Lambda access to secrets
	dbSecret.GrantRead(lambdaFunction, nil)
	jwtSecret.GrantRead(lambdaFunction, nil)

	// Create Lambda Function URL
	functionUrl := lambdaFunction.AddFunctionUrl(&awslambda.FunctionUrlOptions{
		AuthType: awslambda.FunctionUrlAuthType_NONE,
		Cors: &awslambda.FunctionUrlCorsOptions{
			AllowedOrigins: &[]*string{jsii.String("*")},
			AllowedMethods: &[]awslambda.HttpMethod{
				awslambda.HttpMethod_GET,
				awslambda.HttpMethod_POST,
				awslambda.HttpMethod_DELETE,
				awslambda.HttpMethod_OPTIONS,
			},
			AllowedHeaders: &[]*string{jsii.String("*")},
		},
	})

	// Look up the hosted zone for juliops.com
	hostedZone := awsroute53.HostedZone_FromLookup(stack, jsii.String("HostedZone"), &awsroute53.HostedZoneProviderProps{
		DomainName: jsii.String("juliops.com"),
	})

	// Create certificate for the custom domain
	certificate := awscertificatemanager.NewCertificate(stack, jsii.String("Certificate"), &awscertificatemanager.CertificateProps{
		DomainName: jsii.String("jwtapp.juliops.com"),
		Validation: awscertificatemanager.CertificateValidation_FromDns(hostedZone),
	})

	// Note: Lambda Function URLs don't support custom domains directly via CDK
	// We'll create a CloudFront distribution or use a custom domain with API Gateway in the future
	// For now, create a CNAME record that points to the Function URL (after extracting hostname)

	// Output the Function URL
	awscdk.NewCfnOutput(stack, jsii.String("FunctionUrl"), &awscdk.CfnOutputProps{
		Value:       functionUrl.Url(),
		Description: jsii.String("Lambda Function URL"),
		ExportName:  jsii.String("JwtAppFunctionUrl"),
	})

	// Output the database endpoint
	awscdk.NewCfnOutput(stack, jsii.String("DatabaseEndpoint"), &awscdk.CfnOutputProps{
		Value:       dbInstance.DbInstanceEndpointAddress(),
		Description: jsii.String("RDS PostgreSQL database endpoint"),
		ExportName:  jsii.String("JwtAppDatabaseEndpoint"),
	})

	// Output the Redis endpoint
	awscdk.NewCfnOutput(stack, jsii.String("RedisEndpoint"), &awscdk.CfnOutputProps{
		Value:       redisCluster.AttrRedisEndpointAddress(),
		Description: jsii.String("ElastiCache Redis endpoint"),
		ExportName:  jsii.String("JwtAppRedisEndpoint"),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewJwtAppStack(app, "JwtAppStack", &JwtAppStackProps{
		awscdk.StackProps{
			Env: env(),
			Description: jsii.String("JWT App infrastructure with Lambda, RDS, and ElastiCache"),
		},
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
