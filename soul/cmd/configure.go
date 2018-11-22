package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var file string
var composeFile string

var configureCmd = &cobra.Command{
	Use:   "configure tool path",
	Short: "Generate the configuration files required to run a tool",
	Long: `Generate the configuration files required to run a tool.
          Each tool requires different configuration.`,
}

var configureCodeshipCmd = &cobra.Command{
	Use:   "codeship path",
	Short: "Continuous Integration Platform in the cloud",
	Long:  `Codeship runs your automated tests and configured deployment when you push to your repository. It takes care of managing and scaling the infrastructure so that you are able to test and release more frequently .`,
	Run:   configureCodeship,
	Args:  cobra.ExactArgs(1),
}

type File struct {
	EncryptedEnvFile string   `yaml:"encrypted_env_file"`
	Services         []string `yaml:"services"`
}

type codeshipServices map[string]*codeshipService

type codeshipService struct {
	*Service
	DockercfgService string `yaml:"dockercfg_service"`
	DockerfilePath   string `yaml:"dockerfile_path"`
	AddDocker        bool   `yaml:"add_docker"`
	EncryptedEnvFile string `yaml:"encrypted_env_file"`
}

type Compose struct {
	Version  uint8               `yaml:"version"`
	Services map[string]*Service `yaml:"services"`
	Networks interface{}
}

type Service struct {
	Image       string                 `yaml:"image"`
	Command     string                 `yaml:"command"`
	DependsOn   []string               `yaml:"depends_on"`
	Links       []string               `yaml:"links"`
	Environment map[string]interface{} `yaml:"environment"`
	Networks    interface{}            `yaml:"networks"`
}

func configureCodeship(cmd *cobra.Command, args []string) {
	compose := &Compose{}
	f := &File{}
	cs := codeshipServices{}

	err := readFile(composeFile, compose)
	if err != nil {
		panic(err)
	}

	err = readFile(file, f)
	if err != nil {
		panic(err)
	}

	for _, service := range f.Services {
		if s, ok := compose.Services[service]; ok {
			cs[service] = &codeshipService{Service: s}
		}
	}
	fmt.Println(args[0])
	// if err := writeFile(args[0], cs); err != nil {
	// 	panic(err)
	// }

	fmt.Println(compose)
}

func readFile(name string, out interface{}) error {
	b, err := ioutil.ReadFile(composeFile)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(b, out)
	if err != nil {
		return err
	}
	return nil
}

func writeFile(name string, in interface{}) error {
	b, err := yaml.Marshal(in)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(name, b, 664)
}

func init() {
	configureCodeshipCmd.Flags().StringVarP(&file, "file", "f", "", "configuration file")
	configureCodeshipCmd.Flags().StringVar(&composeFile, "compose-file", "", "docker compose file")
	configureCodeshipCmd.MarkFlagRequired("file")
	configureCodeshipCmd.MarkFlagRequired("compose-file")
}
