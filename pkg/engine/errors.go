package engine

import "errors"

var (
	ErrInvalidBuild     = errors.New("invalid build")
	ErrInvalidPipeline  = errors.New("invalid project")
	ErrInvalidProject   = errors.New("invalid pipeline")
	ErrInvalidSecrets   = errors.New("invalid secrets")
	ErrInvalidVariables = errors.New("invalid variables")

	ErrBuildSave       = errors.New("unable to save build to database")
	ErrBuildInitFailed = errors.New("build init ended with error")
)
