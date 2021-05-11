package env

import "fmt"

type awsCredentials struct {
	AccessKeyID     string `envconfig:"AWS_ACCESS_KEY_ID"`
	SecretAccessKey string `envconfig:"AWS_SECRET_ACCESS_KEY"`
}

const (
	AccessKeyID     = "ACCESS_KEY_ID"
	SecretAccessKey = "SECRET_ACCESS_KEY"
)

var (
	awsEnvVars awsCredentials
)

// GetAWSVars devuelve el valor de la variable de ambiente especificada en envar
func GetAWSVars(envar string) (interface{}, error) {
	switch envar {
	case AccessKeyID:
		return awsEnvVars.AccessKeyID, nil
	case SecretAccessKey:
		return awsEnvVars.SecretAccessKey, nil
	default:
		return "", fmt.Errorf("No existe la variable %s", envar)
	}
}
