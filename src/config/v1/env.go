package v1

type envKey = string

const (
	envKeyCollectionEnv envKey = "COLLECTION_ENV"

	envKeyStorage envKey = "STORAGE"

	envKeySessionSecret envKey = "SESSION_SECRET"

	envKeyClientID     envKey = "CLIENT_ID"
	envKeyClientSecret envKey = "CLIENT_SECRET"

	envKeyAdministrators envKey = "ADMINISTRATORS"

	envKeySwiftAuthURL    envKey = "OS_AUTH_URL"
	envKeySwiftUserName   envKey = "OS_USERNAME"
	envKeySwiftPassword   envKey = "OS_PASSWORD"
	envKeySwiftTenantID   envKey = "OS_TENANT_ID"
	envKeySwiftTenantName envKey = "OS_TENANT_NAME"
	envKeySwiftContainer  envKey = "OS_CONTAINER"
	envKeySwiftTmpURLKey  envKey = "OS_TMP_URL_KEY"

	envKeyS3AccessKeyID     envKey = "S3_ACCESS_KEY_ID"
	envKeyS3SecretAccessKey envKey = "S3_SECRET_ACCESS_KEY"
	envKeyS3Region          envKey = "S3_REGION"
	envKeyS3Bucket          envKey = "S3_BUCKET"
	envKeyS3Endpoint        envKey = "S3_ENDPOINT"

	envKeyFilePath envKey = "FILE_PATH"

	envKeyPort envKey = "PORT"

	envKeyDBUserName envKey = "DB_USERNAME"
	envKeyDBPassword envKey = "DB_PASSWORD"
	envKeyDBHostName envKey = "DB_HOSTNAME"
	envKeyDBPort     envKey = "DB_PORT"
	envKeyDBDatabase envKey = "DB_DATABASE"
)
