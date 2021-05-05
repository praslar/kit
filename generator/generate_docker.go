package generator

import (
	"fmt"
	"path"

	yaml "gopkg.in/yaml.v2"

	"strings"

	"github.com/kujtimiihoxha/kit/fs"
	"github.com/kujtimiihoxha/kit/utils"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

// GenerateDocker implements Gen and is used to generate
// docker files for services.
type GenerateDocker struct {
	BaseGenerator
	dockerCompose *DockerCompose
	glide         bool
}

// DockerCompose represents the docker-compose.yml
type DockerCompose struct {
	Version  string                 `yaml:"version"`
	Services map[string]interface{} `yaml:"services"`
}

// BuildService represents one docker service build.
type BuildService struct {
	Context    string `yaml:"context"`
	DockerFile string `yaml:"dockerfile"`
}

// DockerService represents one docker service.
type DockerService struct {
	Image         string `yaml:"image"`
	Command       string `yaml:"command"`
	WorkingDir    string `yaml:"working_dir"`
	Volumes       []string
	Environment   []string `yaml:"environment"`
	Ports         []string `yaml:"ports"`
	Networks      []string `yaml:"networks"`
}

// NewGenerateDocker returns a new docker generator.
func NewGenerateDocker(glide bool) Gen {
	i := &GenerateDocker{
		glide: glide,
	}
	i.dockerCompose = &DockerCompose{}
	i.dockerCompose.Version = "3"
	i.dockerCompose.Services = map[string]interface{}{}
	i.fs = fs.Get()
	return i
}

// Generate generates the docker configurations.
func (g *GenerateDocker) Generate() (err error) {
	f, err := g.fs.Fs.Open(".")
	if err != nil {
		return err
	}
	names, err := f.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, v := range names {
		if b, err := afero.IsDir(g.fs.Fs, v); err != nil {
			return err
		} else if !b {
			continue
		}
		svcFilePath := path.Join(
			fmt.Sprintf(viper.GetString("gk_service_path_format"), utils.ToLowerSnakeCase(v)),
			viper.GetString("gk_service_file_name"),
		)
		httpFilePath := path.Join(
			fmt.Sprintf(viper.GetString("gk_http_path_format"), utils.ToLowerSnakeCase(v)),
			viper.GetString("gk_http_file_name"),
		)
		grpcFilePath := path.Join(
			fmt.Sprintf(viper.GetString("gk_grpc_path_format"), utils.ToLowerSnakeCase(v)),
			viper.GetString("gk_grpc_file_name"),
		)

		err = g.generateDockerFile(v, svcFilePath, httpFilePath, grpcFilePath)
		if err != nil {
			return err
		}
	}
	d, err := yaml.Marshal(g.dockerCompose)
	if err != nil {
		return err
	}
	return g.fs.WriteFile("docker-compose.yml", string(d), true)
}
func (g *GenerateDocker) generateDockerFile(name, svcFilePath, httpFilePath, grpcFilePath string) (err error) {
	pth, err := utils.GetDockerFileProjectPath()
	if err != nil {
		return err
	}
	if b, err := g.fs.Exists(path.Join(name, "Dockerfile")); err != nil {
		return err
	} else if b {
		pth = "/go/src/" + pth
		return g.addToDockerCompose(name, pth, httpFilePath, grpcFilePath)
	}
	if b, err := g.fs.Exists("docker-compose.yml"); err != nil {
		return err
	} else if b {
		r, err := g.fs.ReadFile("docker-compose.yml")
		if err != nil {
			return err
		}
		err = yaml.Unmarshal([]byte(r), g.dockerCompose)
		if err != nil {
			return err
		}
	}
	isService := false
	if b, err := g.fs.Exists(svcFilePath); err != nil {
		return err
	} else if b {
		isService = true
	}

	if !isService {
		return
	}
	dockerFile := `FROM golang:alpine as build-env

ARG VERSION=0.0.0
ARG SERVICE="svc-general"

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# cache dependencies first
WORKDIR /svc
COPY go.mod /svc
COPY go.sum /svc
RUN go mod download

# lastly copy source, any change in source will not break above cache
COPY . /svc

# Build the binary
RUN go build -a -ldflags="-s -w -X 'main.version=${VERSION}' -X 'main.name=${SERVICE}'" -o /app ./main.go

# # <- Second step to build minimal image
FROM alpine:3.11

# RUN apk add --no-cache git ca-certificates tzdata

# we have no self-sign certificate, don't need to update
# && update-ca-certificates
WORKDIR /svc
COPY ./conf/app.conf /svc/conf/app.conf
COPY --from=build-env /app /svc/app

ENTRYPOINT ["/svc/app"]
`
	fpath := "/go/src/" + pth
	err = g.addToDockerCompose(name, fpath, httpFilePath, grpcFilePath)
	if err != nil {
		return err
	}
	return g.fs.WriteFile(path.Join(name, "Dockerfile"), dockerFile, true)
}

func (g *GenerateDocker) addToDockerCompose(name, pth, httpFilePath, grpcFilePath string) (err error) {
	hasHTTP := false
	hasGRPC := false
	if b, err := g.fs.Exists(httpFilePath); err != nil {
		return err
	} else if b {
		hasHTTP = true
	}
	if b, err := g.fs.Exists(grpcFilePath); err != nil {
		return err
	} else if b {
		hasGRPC = true
	}
	usedPorts := []string{}
	for _, v := range g.dockerCompose.Services {
		k, ok := v.(map[interface{}]interface{})
		if ok {
			for _, p := range k["ports"].([]interface{}) {
				pt := strings.Split(p.(string), ":")
				usedPorts = append(usedPorts, pt[0])
			}
		} else {
			for _, p := range v.(*DockerService).Ports {
				pt := strings.Split(p, ":")
				usedPorts = append(usedPorts, pt[0])
			}
		}

	}
	if g.dockerCompose.Services[name] == nil {
		g.dockerCompose.Services[name] = &DockerService{
			Image:       "golang:1.14",
			Command:     "[\"go\", \"run\", \"main.go\"]",
			WorkingDir:  "/"+name,
			Volumes:     []string{
				".:" + " /" + name,
				"shared_gopath:/gopath",
			},
			Environment: []string{
				" HTTP_PORT: 80",
				"GOPATH: /gopath",
				" ELASTIC_APM_SERVER_URL: 3.1.24.78:8200",
			},
			Ports:       []string{"8109:80"},
			Networks:    []string{"finan_network"},
		}
		if hasHTTP {
			httpExpose := 8800
			for {
				ex := false
				for _, v := range usedPorts {
					if v == fmt.Sprintf("%d", httpExpose) {
						ex = true
						break
					}
				}
				if ex {
					httpExpose++
				} else {
					break
				}
			}
			g.dockerCompose.Services[name].(*DockerService).Ports = []string{
				fmt.Sprintf("%d", httpExpose) + ":8081",
			}
			usedPorts = append(usedPorts, fmt.Sprintf("%d", httpExpose))
		}
		if hasGRPC {
			grpcExpose := 8800
			for {
				ex := false
				for _, v := range usedPorts {
					if v == fmt.Sprintf("%d", grpcExpose) {
						ex = true
						break
					}
				}
				if ex {
					grpcExpose++
				} else {
					break
				}
			}
			if g.dockerCompose.Services[name].(*DockerService).Ports == nil {
				g.dockerCompose.Services[name].(*DockerService).Ports = []string{}
			}
			g.dockerCompose.Services[name].(*DockerService).Ports = append(
				g.dockerCompose.Services[name].(*DockerService).Ports,
				fmt.Sprintf("%d", grpcExpose)+":8082",
			)
		}
	}
	return
}
