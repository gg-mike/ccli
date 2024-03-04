package engine

import (
	"encoding/base64"
	"fmt"
	"maps"
	"regexp"
	"strings"

	"github.com/gg-mike/ccli/pkg/model"
)

type envInstance struct {
	value string
	path  string
}

func createEnvSteps(ctx *model.QueueContext) error {
	workdirSteps, workdirCleanup := createWorkdirStep(ctx)
	secretsSteps, secretsCleanup, err := createSecretsStep(ctx)
	if err != nil {
		return err
	}
	variablesSteps, variablesCleanup, err := createVariablesStep(ctx)
	if err != nil {
		return err
	}

	ctx.Config.Steps = append([]model.PipelineConfigStep{
		workdirSteps, secretsSteps, variablesSteps,
	}, ctx.Config.Steps...)

	ctx.Config.Cleanup = append(ctx.Config.Cleanup, workdirCleanup...)
	ctx.Config.Cleanup = append(ctx.Config.Cleanup, secretsCleanup...)
	ctx.Config.Cleanup = append(ctx.Config.Cleanup, variablesCleanup...)

	return nil
}

func createWorkdirStep(ctx *model.QueueContext) (model.PipelineConfigStep, []string) {
	workdir := strings.ReplaceAll(ctx.Build.ID(), "/", "_")
	return model.PipelineConfigStep{
		Name:     "Work dir setup",
		Commands: []string{"cd ~", "mkdir -p " + workdir, "cd " + workdir},
	}, []string{"cd ~", "rm -rf " + workdir}
}

func createSecretsStep(ctx *model.QueueContext) (model.PipelineConfigStep, []string, error) {
	secrets := map[string]envInstance{}
	for _, secret := range ctx.Secrets {
		value, err := secret.Value()
		if err != nil {
			return model.PipelineConfigStep{}, []string{}, err
		}
		secrets[secret.Key] = envInstance{value, secret.Path}
	}

	commands, cleanUpCommands, err := prepareStepCommands(ctx.Config.System, secrets, "_")
	if err != nil {
		return model.PipelineConfigStep{}, []string{}, err
	}

	return model.PipelineConfigStep{Name: "Secret exports", Commands: commands}, cleanUpCommands, nil
}

func createVariablesStep(ctx *model.QueueContext) (model.PipelineConfigStep, []string, error) {
	variables := map[string]envInstance{}

	variables["__PROJECT_NAME"] = envInstance{ctx.Build.ProjectName, ""}
	variables["__REPO"] = envInstance{ctx.Repo, ""}

	variables["__PIPELINE_NAME"] = envInstance{ctx.Build.PipelineName, ""}
	variables["__BRANCH"] = envInstance{ctx.Branch, ""}
	maps.Copy(variables, setRepoVariables(ctx.Repo))

	for _, variable := range ctx.Variables {
		variables[variable.Key] = envInstance{variable.Value, variable.Path}
	}

	commands, cleanUpCommands, err := prepareStepCommands(ctx.Config.System, variables, "")
	if err != nil {
		return model.PipelineConfigStep{}, []string{}, err
	}

	return model.PipelineConfigStep{Name: "Variable exports", Commands: commands}, cleanUpCommands, nil
}

func prepareStepCommands(system string, env map[string]envInstance, prefix string) ([]string, []string, error) {
	var templateEnv, templateFile, templateFileDelete string
	// TODO: support for over OS
	if strings.ToLower(system) == "linux" {
		templateEnv = "export %s%s=\"%s\""
		templateFile = "export %s%s=\"%s\" && echo '%s' > %s"
		templateFileDelete = "rm -f %s"
	} else {
		panic(system + " is not supported")
	}
	commands := []string{}
	cleanUpCommands := []string{}
	for k, v := range env {
		if v.path != "" {
			value, err := base64.StdEncoding.DecodeString(v.value)
			if err != nil {
				return []string{}, []string{}, err
			}
			commands = append(commands, fmt.Sprintf(templateFile, prefix, k, v.path, value, v.path))
			cleanUpCommands = append(cleanUpCommands, fmt.Sprintf(templateFileDelete, v.path))
		} else {
			commands = append(commands, fmt.Sprintf(templateEnv, prefix, k, v.value))
		}
	}
	return commands, cleanUpCommands, nil
}

func setRepoVariables(repo string) map[string]envInstance {
	// TODO: support for over providers
	if repo == "" {
		return map[string]envInstance{}
	} else if strings.Contains(repo, "git@github.com") {
		re := regexp.MustCompile(`git@github\.com:(?P<owner>[\w\d\.\-_]+)\/(?P<name>[\w\d\.\-_]+)\.git`)
		matches := re.FindStringSubmatch(repo)
		return map[string]envInstance{
			"__GITHUB_OWNER": {matches[re.SubexpIndex("owner")], ""},
			"__GITHUB_NAME":  {matches[re.SubexpIndex("name")], ""},
		}
	} else if strings.Contains(repo, "https://github.com") {
		re := regexp.MustCompile(`https:\/\/github\.com\/(?P<owner>[\w\d\.\-_]+)\/(?P<name>[\w\d\.\-_]+)`)
		matches := re.FindStringSubmatch(repo)
		return map[string]envInstance{
			"__GITHUB_OWNER": {matches[re.SubexpIndex("owner")], ""},
			"__GITHUB_NAME":  {matches[re.SubexpIndex("name")], ""},
		}
	}
	return map[string]envInstance{}
}
