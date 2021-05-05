package main

import (
	"path"
	"runtime"

	"github.com/kujtimiihoxha/kit/cmd"
	"github.com/spf13/viper"
)

func main() {
	setDefaults()
	viper.AutomaticEnv()
	cmd.Execute()
}

func setDefaults() {
	// thang.pham: new code
	viper.SetDefault("gk_model_path_format", path.Join("%s", "pkg", "model"))
	viper.SetDefault("gk_postgres_path_format", path.Join("%s", "pkg", "db", "postgres"))
	viper.SetDefault("gk_config_path_format", path.Join("%s", "config"))
	viper.SetDefault("gk_utils_path_format", path.Join("%s", "pkg", "utils"))
	viper.SetDefault("gk_grpc_path_format", path.Join("%s", "pkg", "grpc"))

	viper.SetDefault("gk_model_file_name", "base.go")
	viper.SetDefault("gk_db_postgres_file_name", "postgres.go")
	viper.SetDefault("gk_config_file_name", "config.go")
	viper.SetDefault("gk_status_file_name", "status.yml")
	viper.SetDefault("gk_utils_utils_file_name", "utils.go")
	viper.SetDefault("gk_utils_constant_file_name", "constant.go")
	viper.SetDefault("gk_utils_status_file_name", "status.go")
	viper.SetDefault("gk_db_postgre_file_name", "db.go")
	viper.SetDefault("gk_config_postgre_file_name", "config.go")
	viper.SetDefault("gk_docker_compose_file_name", "docker-compose.yml")
	// end new code

	viper.SetDefault("gk_service_path_format", path.Join("%s", "pkg", "service"))
	viper.SetDefault("gk_cmd_service_path_format", path.Join("%s", "cmd", "service"))
	viper.SetDefault("gk_cmd_path_format", path.Join("%s", "cmd"))
	viper.SetDefault("gk_endpoint_path_format", path.Join("%s", "pkg", "endpoint"))
	viper.SetDefault("gk_http_path_format", path.Join("%s", "pkg", "http"))
	viper.SetDefault("gk_http_client_path_format", path.Join("%s", "client", "http"))
	viper.SetDefault("gk_grpc_client_path_format", path.Join("%s", "client", "grpc"))
	viper.SetDefault("gk_client_cmd_path_format", path.Join("%s", "cmd", "client"))
	viper.SetDefault("gk_grpc_path_format", path.Join("%s", "pkg", "grpc"))
	viper.SetDefault("gk_grpc_pb_path_format", path.Join("%s", "pkg", "grpc", "pb"))

	viper.SetDefault("gk_service_file_name", "service.go")
	viper.SetDefault("gk_service_middleware_file_name", "middleware.go")
	viper.SetDefault("gk_endpoint_base_file_name", "endpoint_gen.go")
	viper.SetDefault("gk_endpoint_file_name", "endpoint.go")
	viper.SetDefault("gk_endpoint_middleware_file_name", "middleware.go")
	viper.SetDefault("gk_http_file_name", "handler.go")
	viper.SetDefault("gk_http_base_file_name", "handler_gen.go")
	viper.SetDefault("gk_cmd_base_file_name", "service_gen.go")
	viper.SetDefault("gk_cmd_svc_file_name", "service.go")
	viper.SetDefault("gk_http_client_file_name", "http.go")
	viper.SetDefault("gk_grpc_client_file_name", "grpc.go")
	viper.SetDefault("gk_grpc_pb_file_name", "%s.proto")
	viper.SetDefault("gk_grpc_base_file_name", "handler_gen.go")
	viper.SetDefault("gk_grpc_file_name", "handler.go")
	if runtime.GOOS == "windows" {
		viper.SetDefault("gk_grpc_compile_file_name", "compile.bat")
	} else {
		viper.SetDefault("gk_grpc_compile_file_name", "compile.sh")
	}
	viper.SetDefault("gk_service_struct_prefix", "basic")

}
