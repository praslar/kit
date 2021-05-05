package generator

import (
	"bytes"
	"fmt"
	"github.com/dave/jennifer/jen"
	"github.com/kujtimiihoxha/kit/fs"
	"github.com/kujtimiihoxha/kit/utils"
	"github.com/spf13/viper"
	"os/exec"
	"path"
	"strings"
)


// NewService implements Gen and is used to create a new service.
type NewService struct {
	BaseGenerator
	name          string
	interfaceName string
	destPath      string
	filePath      string
}

// NewNewService returns a initialized and ready generator.
//
// The name parameter is the name of the service that will be created
// this name should be without the `Service` suffix
func NewNewService(name string) Gen {
	gs := &NewService{
		name:          name,
		interfaceName: utils.ToCamelCase(name + "Service"),
		destPath:      fmt.Sprintf(viper.GetString("gk_service_path_format"), utils.ToLowerSnakeCase(name)),
	}
	gs.filePath = path.Join(gs.destPath, viper.GetString("gk_service_file_name"))
	gs.srcFile = jen.NewFilePath(strings.Replace(gs.destPath, "\\", "/", -1))
	gs.InitPg()
	gs.fs = fs.Get()
	return gs
}

// Generate will run the generator.
func (g *NewService) Generate() error {
	g.CreateFolderStructure(g.destPath)
	err := g.genModule()
	if err != nil {
		println(err.Error())
		return err
	}

	comments := []string{
		"Add your methods here",
		"e.x: Foo(ctx context.Context,s string)(rs string, err error)",
	}
	partial := NewPartialGenerator(nil)
	partial.appendMultilineComment(comments)
	g.code.Raw().Commentf("%s describes the service.", g.interfaceName).Line()
	g.code.appendInterface(
		g.interfaceName,
		[]jen.Code{partial.Raw()},
	)

	return g.fs.WriteFile(g.filePath, g.srcFile.GoString(), false)
}

func (g *NewService) genModule() error {
	prjName := utils.ToLowerSnakeCase(g.name)
	exist, _ := g.fs.Exists(prjName + "/go.mod")
	if exist {
		return nil
	}

	moduleName := prjName
	if viper.GetString("n_s_module") != "" {
		moduleName = viper.GetString("n_s_module")
		moduleNameSlice := strings.Split(moduleName, "/")
		moduleNameSlice[len(moduleNameSlice)-1] = utils.ToLowerSnakeCase(moduleNameSlice[len(moduleNameSlice)-1])
		moduleName = strings.Join(moduleNameSlice, "/")
	}
	cmdStr := "cd " + prjName + " && go mod init " + moduleName
	cmd := exec.Command("sh", "-c", cmdStr)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	_, err := cmd.Output()
	// return cmd.Stderr to debug (err here provides nothing useful, only `exit status 1`)
	if err != nil {
		return fmt.Errorf("genModule: sh -c %s => err:%v", cmdStr, err.Error()+" , "+stderr.String())
	}
	return nil
}

// thang.pham: new code
type NewModel struct {
	BaseGenerator
	name      string
	modelName string
	destPath  string
	filePath  string
	postgreFilePath  string
}

func NewNewModel(name string) Gen {
	gs := &NewModel{
		name:          name,
		modelName: utils.ToCamelCase(name),
		destPath:      fmt.Sprintf(viper.GetString("gk_model_path_format"), utils.ToLowerSnakeCase(name)),
	}
	gs.postgreFilePath =  path.Join(gs.destPath, viper.GetString("gk_db_postgres_file_name"))
	gs.filePath = path.Join(gs.destPath, viper.GetString("gk_model_file_name"))
	gs.srcFile = jen.NewFilePath(strings.Replace(gs.destPath, "\\", "/", -1))

	gs.InitPg()
	gs.fs = fs.Get()
	return gs
}

func (g *NewModel) Generate() error {
	g.CreateFolderStructure(g.destPath)
	g.code.Raw().Commentf("%s describes the structure.", g.modelName).Line()
	g.code.appendStruct("BaseModel",
		jen.Id("ID").Qual("github.com/google/uuid", "UUID").Tag(map[string]string{
			"json": "id",
			"gorm": "primary_key;type:uuid;default:uuid_generate_v4()",
		}),
		jen.Id("CreatorID").Qual("github.com/google/uuid", "UUID").Tag(map[string]string{
			"json": "creator_id",
		}),
		jen.Id("UpdaterID").Qual("github.com/google/uuid", "UUID").Tag(map[string]string{
			"json": "updater_id",
		}),
		jen.Id("CreatedAt").Qual(	"time", "Time").Tag(map[string]string{
			"gorm": "column:created_at;default:CURRENT_TIMESTAMP",
		}),
		jen.Id("UpdatedAt").Qual(	"time", "Time").Tag(map[string]string{
			"gorm": "column:updated_at;default:CURRENT_TIMESTAMP",
		}),
		jen.Id("DeletedAt").Id("*").Qual(	"time", "Time").Tag(map[string]string{
			"json": "deleted_at",
			"sql" :	"index",
		}),
	)

	// get posgrest import
	postgresImports, err := utils.GetPostgresImportPath(g.name)
	if err != nil {
		return err
	}
	g.code.appendFunction(
		"AutoMigration",
		nil,
		nil,
		[]jen.Code{jen.Id("err").Error()},
		"",
		jen.List(jen.Id("dbPublic"), jen.Err()).Op(":=").Qual(postgresImports, "GetDatabase").Call(
			jen.Lit("default"),
		),
		jen.If(
			jen.Err().Op("!=").Nil().Block(
				jen.Return(),
			),
		),
		jen.List(jen.Id("_"), jen.Err()).Op("=").Id("dbPublic").Dot("DB").Call().Dot("Exec").Call(
			jen.Lit("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\""),
		),
		jen.If(
			jen.Err().Op("!=").Nil().Block(
				jen.Return(
					jen.Qual("fmt", "Errorf").Call(
						jen.Lit("error while creating DB extension 'uuid-ossp': %s"),
						jen.Id("err"),
					),
				),
			),
		),
		jen.Id("t").Op(":=").Id("dbPublic").Dot("AutoMigrate").Call(),
		jen.Return(
			jen.Id("t").Dot("Error"),
		),
	)
		postgresDB := `package model

import (
	"context"
	"github.com/jinzhu/gorm"
	"github.com/google/uuid"
)

func NewBasicPostgresDatabase(db *gorm.DB) PostgresDatabase {
	return &basicPostgresDatabase{
		db: db,
	}
}

type basicPostgresDatabase struct {
	db *gorm.DB
}

func (b basicPostgresDatabase) Create(ctx context.Context, req interface{}, creatorID uuid.UUID) (interface{}, error) {
	panic("implement me")
}

func (b basicPostgresDatabase) Update(ctx context.Context, req interface{}, objectID uuid.UUID, updaterID uuid.UUID) (interface{}, error) {
	panic("implement me")
}

func (b basicPostgresDatabase) GetOneByID(ctx context.Context, id uuid.UUID) (interface{}, error) {
	panic("implement me")
}

func (b basicPostgresDatabase) GetAll(ctx context.Context) (interface{}, error) {
	panic("implement me")
}

func (b basicPostgresDatabase) Delete(ctx context.Context, id uuid.UUID) error {
	panic("implement me")
}

type PostgresDatabase interface {
// Add your db method here,
// e.x: Create(s User)(rs User, err error)",
	Create(ctx context.Context, req interface{}, creatorID uuid.UUID) (interface{}, error)
	Update(ctx context.Context, req interface{}, objectID uuid.UUID, updaterID uuid.UUID) (interface{}, error)
	GetOneByID(ctx context.Context, id uuid.UUID) (interface{}, error)
	GetAll(ctx context.Context) (interface{}, error)
	Delete(ctx context.Context, id uuid.UUID) error
}`
	g.fs.WriteFile(g.postgreFilePath, postgresDB, false)
	return g.fs.WriteFile(g.filePath, g.srcFile.GoString(), false)
}

// thang.pham: new code
type NewConfig struct {
	BaseGenerator
	name      string
	modelName string
	destPath  string

	statusFilePath  string
	configFilePath  string
}

func NewNewConfig(name string) Gen {
	gs := &NewConfig{
		name:          name,
		modelName: utils.ToCamelCase(name),
		destPath:      fmt.Sprintf(viper.GetString("gk_config_path_format"), utils.ToLowerSnakeCase(name)),
	}

	gs.statusFilePath = path.Join(gs.destPath, viper.GetString("gk_status_file_name"))
	gs.configFilePath = path.Join(gs.destPath, viper.GetString("gk_config_file_name"))
	gs.srcFile = jen.NewFilePath(strings.Replace(gs.destPath, "\\", "/", -1))

	gs.InitPg()
	gs.fs = fs.Get()
	return gs
}

func (g *NewConfig) Generate() error {
	g.CreateFolderStructure(g.destPath)

	g.code.appendStruct("AppConfig",
		jen.Id("DBHost").Id("string").Tag(map[string]string{
			"env": "DB_HOST",
			"envDefault": "localhost",
		}),
		jen.Id("DBPort").Id("string").Tag(map[string]string{
			"env": "DB_PORT",
			"envDefault": "5432",
		}),
		jen.Id("DBUser").Id("string").Tag(map[string]string{
			"env": "DB_USER",
			"envDefault": "postgres",
		}),
		jen.Id("DBPass").Id("string").Tag(map[string]string{
			"env": "DB_PASS",
			"envDefault": "123456",
		}),
		jen.Id("DBName").Id("string").Tag(map[string]string{
			"env": "DB_NAME",
			"envDefault": "postgres",
		}),
		jen.Id("DBSchema").Id("string").Tag(map[string]string{
			"env": "DB_SCHEMA",
			"envDefault": "public",
		}),
		jen.Id("LogFormat").Id("string").Tag(map[string]string{
			"env": "LOG_FORMAT",
			"envDefault": "text",
		}),
		jen.Id("LogLevel").Id("string").Tag(map[string]string{
			"env": "LOG_LEVEL",
			"envDefault": "debug",
		}),
		jen.Id("LogOutput").Id("string").Tag(map[string]string{
			"env": "LOG_OUTPUT",
			"envDefault": "file://logs/metadata.log",
		}),
	)

	statusYml := `gen:
  success:
    code: 1001
    status: 200
    message: "Success"

  not_found:
    code: 1002
    status: 404
    message: "Not found"

  timeout:
    code: 1003
    status: 500
    message: "Timeout"

  bad_request:
    code: 1004
    status: 400
    message: "Invalid input"

  internal:
    code: 1005
    status: 500
    message: "Internal server error"

  unauthorized:
    code: 1006
    status: 401
    message: "Authorization required"

  database:
    code: 1007
    status: 500
    message: "Database error"
`
	// write status.yml
	g.fs.WriteFile(g.statusFilePath, statusYml, true)

	// write config.go
	return g.fs.WriteFile(g.configFilePath, g.srcFile.GoString(), false)
}


type NewUtils struct {
	BaseGenerator
	name      string
	modelName string
	destPath  string

	constantFilePath  string
	statusFilePath  string
	utilFilePath  string
}

func (n NewUtils) Generate() error {
	n.CreateFolderStructure(n.destPath)
	configImport, err := utils.GetConfigImportPath(n.name)
	if err != nil {
		return err
	}
	n.code.raw.Var().Id("conf").Qual(configImport,"AppConfig")
	n.code.NewLine()
	n.code.appendFunction("IsErrNotFound",
	nil,
	[]jen.Code{jen.Id("err").Error()},
	nil,
	"bool",
		jen.Return(
			jen.Qual("errors", "Is").Call(
				jen.Id("err"),
				jen.Qual("github.com/jinzhu/gorm","ErrRecordNotFound"),
			),
		),
	)
	n.code.NewLine()
	n.code.appendFunction("LoadEnv",
		nil,
		nil,
		nil,
		"",
		jen.Id("_").Op("=").Qual("github.com/caarlos0/env/v6","Parse").Call(
			jen.Id("&conf"),
		),
	)
	n.code.NewLine()
	n.code.appendFunction("GetEnv",
		nil,
		nil,
		[]jen.Code{jen.Qual(configImport,"AppConfig")},
		"",
		jen.Return(
			jen.Id("conf"),
		),
	)
	// write status.yml
	constant := `package utils`
	n.fs.WriteFile(n.constantFilePath, constant, true)

	status := `
package utils

import (

	"github.com/praslar/common/response"
	"github.com/sirupsen/logrus"

	"gopkg.in/yaml.v3"
	"os"
	"sync"
)

type (
	//Status format from status pkg
	Status = response.ResponseStatus

	GenStatus struct {
		Success      Status
		BadRequest   Status 
		Unauthorized Status
		Internal     Status
		Database     Status
	}

	statuses struct {
		Gen              GenStatus
	}
)

var (
	all  *statuses
	once sync.Once
)

// Init load statuses from the given config file.
// Init panics if cannot access or error while parsing the config file.
func Init(conf string) {
	once.Do(func() {
		f, err := os.Open(conf)
		if err != nil {
			logrus.Errorf("Fail to open status file, %v", err)
			panic(err)
		}
		all = &statuses{}
		if err := yaml.NewDecoder(f).Decode(all); err != nil {
			logrus.Errorf("Fail to parse status file data to statuses struct, %v", err)
			panic(err)
		}
	})
}

// all return all registered statuses.
// all will load statuses from configs/Status.yml if the statuses has not initialized yet.
func load(err string) *statuses {
	conf := os.Getenv("STATUS_PATH")
	if conf == "" {
		conf = "conf/status.yml"
	}
	Init(conf)
	if err != "" {
		all.Gen.BadRequest.XMessage = err
	}
	return all
}

func Gen(err string) GenStatus {
	return load(err).Gen
}
`
	n.fs.WriteFile(n.statusFilePath, status, true)
	// write config.go
	return n.fs.WriteFile(n.utilFilePath, n.srcFile.GoString(), false)
}

func NewNewUtils(name string) Gen {
	gs := &NewUtils{
		name:          name,
		modelName: 	   utils.ToCamelCase(name),
		destPath:      fmt.Sprintf(viper.GetString("gk_utils_path_format"), utils.ToLowerSnakeCase(name)),
	}

	gs.constantFilePath = path.Join(gs.destPath, viper.GetString("gk_utils_constant_file_name"))
	gs.utilFilePath = path.Join(gs.destPath, viper.GetString("gk_utils_utils_file_name"))
	gs.statusFilePath = path.Join(gs.destPath, viper.GetString("gk_utils_status_file_name"))
	gs.srcFile = jen.NewFilePath(strings.Replace(gs.destPath, "\\", "/", -1))

	gs.InitPg()
	gs.fs = fs.Get()
	return gs
}

type NewPostgreDatabase struct {
	BaseGenerator

	name      string
	modelName string
	destPath  string

	postgreFilePath  string
	configFilePath  string
}

func (n NewPostgreDatabase) Generate() error {
	n.CreateFolderStructure(n.destPath)
	configImport, err := utils.GetConfigImportPath(n.name)
	if err != nil {
		return err
	}
	n.code.raw.Var().Id("Muxtex").Id("*").Qual("sync","RWMutex")
	n.code.NewLine()
	n.code.raw.Const().Id("DefaultConnName").Op("=").Lit("default")
	n.code.NewLine()
	n.code.appendFunction("GetDBInfo",
		nil,
		nil,
		[]jen.Code{jen.Id("dbInfo").Id("DBInfo") ,jen.Id("err").Error()},
		"",
		jen.Id("conf").Op(":=").Qual(configImport,"AppConfig").Block(),
		jen.Id("_").Op("=").Qual("github.com/caarlos0/env/v6","Parse").Call(jen.Id("&conf")),
		jen.Id("dbInfo").Op("=").Id("DBInfo").Values(
				jen.Id("Host").Op(":").Id("conf").Dot("DBHost"),
				jen.Id("Port").Op(":").Id("conf").Dot("DBPort"),
				jen.Id("Name").Op(":").Id("conf").Dot("DBName"),
				jen.Id("User").Op(":").Id("conf").Dot("DBUser"),
				jen.Id("Pass").Op(":").Id("conf").Dot("DBPass"),
				jen.Id("SearchPath").Op(":").Id("conf").Dot("DBSchema"),
		),
		jen.Return(),
	)
	n.code.NewLine()
	n.code.appendFunction("GetDatabase",
		nil,
		[]jen.Code{jen.Id("aliasName").String()},
		[]jen.Code{jen.Id("db").Id("*").Qual("github.com/jinzhu/gorm","DB") ,jen.Id("err").Error()},
		"",
		jen.Var().Id("customerSchema").String(),
		jen.Id("err").Op("=").Nil(),
		jen.List(jen.Id("_"),jen.Id("errConv")).Op(":=").Qual("strconv","Atoi").Call(jen.Id("aliasName")),
		jen.If(jen.Id("errConv").Op("==").Nil()).Block(
			jen.Id("customerSchema").Op("=").Id("aliasName"),
		).Else().Block(
			jen.Id("customerSchema").Op("=").Lit("default"),
		),
		jen.List(jen.Id("db"),jen.Id("err")).Op("=").Id("GetDB").Call(jen.Id("customerSchema")),
		jen.If(jen.Id("err").Op("!=").Nil()).Block(
			jen.Var().Id("dbInfo").Id("DBInfo"),
			jen.List(jen.Id("dbInfo"),jen.Id("err")).Op("=").Id("GetDBInfo").Call(),
			jen.If(jen.Id("err").Op("==").Nil()).Block(
				jen.Id("err").Op("=").Id("RegisterDataBase").Call(jen.Id("customerSchema"),jen.Lit("postgres"),jen.Id("CreateDBConnectionString").Call(jen.Id("dbInfo"))),
				jen.If(jen.Id("err").Op("==").Nil()).Block(
					jen.List(jen.Id("db"),jen.Id("err")).Op("=").Id("GetDB").Call(jen.Id("customerSchema")),
				),
			),
		),
		jen.Return(),
	)

	db := `package postgres

import (
	"fmt"
	"sync"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type DBInfo struct {
	Host       string
	Port       string
	Name       string
	User       string
	Pass       string
	SearchPath string
}

var (
	dataBaseCache = &_dbCache{cache: make(map[string]*alias)}
)

type DB struct {
	*sync.RWMutex
	DB *gorm.DB
}

type alias struct {
	Name         string
	MaxIdleConns int
	MaxOpenConns int
	DB           *DB
}

type _dbCache struct {
	mux   sync.RWMutex
	cache map[string]*alias
}

func (ac *_dbCache) add(name string, al *alias) (added bool) {
	ac.mux.Lock()
	defer ac.mux.Unlock()
	if _, ok := ac.cache[name]; !ok {
		ac.cache[name] = al
		added = true
	}
	return
}

// get db alias if cached.
func (ac *_dbCache) get(name string) (al *alias, ok bool) {
	ac.mux.RLock()
	defer ac.mux.RUnlock()
	al, ok = ac.cache[name]
	return
}

// get default alias.
func (ac *_dbCache) getDefault() (al *alias) {
	al, _ = ac.get("default")
	return
}

func GetDB(aliasNames ...string) (*gorm.DB, error) {
	var name string
	if len(aliasNames) > 0 {
		name = aliasNames[0]
	} else {
		name = "default"
	}
	al, ok := dataBaseCache.get(name)
	if ok {
		return al.DB.DB.Debug(), nil
	}
	return &gorm.DB{}, fmt.Errorf("DataBase of alias name %s not found", name)
}

func RegisterDataBase(aliasName, driverName, dataSource string, params ...int) error {
	var (
		err error
		db  *gorm.DB
		al  *alias
	)

	db, err = gorm.Open(driverName, dataSource)
	if err != nil {
		err = fmt.Errorf("register db %s, %s", aliasName, err.Error())
		goto end
	}

	al, err = addAliasWthDB(aliasName, driverName, db)
	if err != nil {
		goto end
	}

	for i, v := range params {
		switch i {
		case 0:
			SetMaxIdleConns(al.Name, v)
		case 1:
			SetMaxOpenConns(al.Name, v)
		}
	}

end:
	if err != nil {
		if db != nil {
			//_ = db.Close()
		}
	}

	return err
}

func addAliasWthDB(aliasName, driverName string, db *gorm.DB) (*alias, error) {
	al := new(alias)
	al.Name = aliasName
	al.DB = &DB{
		RWMutex: new(sync.RWMutex),
		DB:      db,
	}

	err := db.DB().Ping()
	if err != nil {
		return nil, fmt.Errorf("register db Ping %s, %s", aliasName, err.Error())
	}

	if !dataBaseCache.add(aliasName, al) {
		return nil, fmt.Errorf("DataBase alias name %s already registered, cannot reuse", aliasName)
	}

	return al, nil
}

// get table alias.
func getDbAlias(name string) *alias {
	if al, ok := dataBaseCache.get(name); ok {
		return al
	}
	panic(fmt.Errorf("unknown DataBase alias name %s", name))
}

// SetMaxIdleConns ChangeNumber the max idle conns for *sql.DB, use specify db alias name
func SetMaxIdleConns(aliasName string, maxIdleConns int) {
	al := getDbAlias(aliasName)
	al.MaxIdleConns = maxIdleConns
	al.DB.DB.DB().SetMaxIdleConns(maxIdleConns)
}

// SetMaxOpenConns ChangeNumber the max open conns for *sql.DB, use specify db alias name
func SetMaxOpenConns(aliasName string, maxOpenConns int) {
	al := getDbAlias(aliasName)
	al.MaxOpenConns = maxOpenConns
	al.DB.DB.DB().SetMaxOpenConns(maxOpenConns)
}

func CreateDBConnectionString(info DBInfo) (dbConnString string) {
	host := info.Host
	port := info.Port
	database := info.Name
	user := info.User
	pass := info.Pass
	searchPath := info.SearchPath
	dbConnString = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s search_path=%s", host, port, user, database, pass, searchPath)
	return
}
`
	n.fs.WriteFile(n.postgreFilePath, db, true)
	// write config.go
	return n.fs.WriteFile(n.configFilePath, n.srcFile.GoString(), false)
}

func NewNewPostgreDatabase(name string) Gen {
	gs := &NewPostgreDatabase{
		name:          name,
		modelName: 	   utils.ToCamelCase(name),
		destPath:      fmt.Sprintf(viper.GetString("gk_postgres_path_format"), utils.ToLowerSnakeCase(name)),
	}
	gs.postgreFilePath = path.Join(gs.destPath, viper.GetString("gk_db_postgre_file_name"))
	gs.configFilePath = path.Join(gs.destPath, viper.GetString("gk_config_postgre_file_name"))
	gs.srcFile = jen.NewFilePath(strings.Replace(gs.destPath, "\\", "/", -1))

	gs.InitPg()
	gs.fs = fs.Get()
	return gs
}