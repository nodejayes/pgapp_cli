package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path"
	"strings"
)

func createMetaAppFile(targetDir, projectName, description string) error {
	return os.WriteFile(path.Join(targetDir, "app.go"), []byte(fmt.Sprintf(`package meta

var AppName = "%[1]v"
var Description = "%[2]v"`, projectName, description)), 0755)
}

func createMetaVersionFile(targetDir string) error {
	return os.WriteFile(path.Join(targetDir, "version.go"), []byte(fmt.Sprintf(`package meta

var Version = "1.0.0"`)), 0755)
}

func createMetaFolder(targetDir string) (string, error) {
	metaDir := path.Join(targetDir, "meta")
	return metaDir, os.MkdirAll(metaDir, 0755)
}

func createActionBaseFile(targetDir, modulePath, projectName string) error {
	actionDir := path.Join(targetDir, "pgapi", "action")
	err := os.MkdirAll(actionDir, 0755)
	if err != nil {
		return err
	}
	return os.WriteFile(path.Join(actionDir, "base.go"), []byte(fmt.Sprintf(`package action

import (
	"encoding/json"
	"github.com/go-playground/validator"
	"%[1]v/%[2]v/pgapi/format"
)

type BaseAction struct {
	validator *validator.Validate
}

func NewBaseAction() BaseAction {
	return BaseAction{
		validator: validator.New(),
	}
}

func (ba BaseAction) DecodeInput(input string, target interface{}) error {
	err := json.Unmarshal([]byte(input), &target)
	if err != nil {
		return err
	}
	return ba.validator.Struct(target)
}

func (ba BaseAction) EncodeResult(res []any, def format.ResultMessage) format.ResultMessage {
	if len(res) < 1 {
		return def
	}
	return format.ResultMessage{
		Data:    res,
		Message: "",
		State:   format.Success,
	}
}
`, modulePath, projectName)), 0755)
}

func createFormatPluginConfigFile(targetDir string) error {
	return os.WriteFile(path.Join(targetDir, "plugin_config.go"), []byte(fmt.Sprintf(`package format

type PluginConfig struct {
}`)), 0755)
}

func createFormatResultFile(targetDir string) error {
	return os.WriteFile(path.Join(targetDir, "result_message.go"), []byte(fmt.Sprintf(`package format

type ResultMessage struct {
	Data    []any
	Message string
	State   ResultState
}`)), 0755)
}

func createFormatStateFile(targetDir string) error {
	return os.WriteFile(path.Join(targetDir, "state.go"), []byte(fmt.Sprintf(`package format

type ResultState = int

const (
	Success ResultState = 0
	Error               = -1
	Warning             = -2
	Info                = 1
)`)), 0755)
}

func createFormatFolder(targetDir string) (string, error) {
	formatDir := path.Join(targetDir, "pgapi", "format")
	return formatDir, os.MkdirAll(formatDir, 0755)
}

func createProjectActionFile(targetDir, modulePath, projectName string) error {
	return os.WriteFile(path.Join(targetDir, "hello.go"), []byte(fmt.Sprintf(`package %[2]v

import (
	"fmt"
	"%[1]v/%[2]v/pgapi/action"
	"%[1]v/%[2]v/pgapi/format"
)

type (
	HelloActionParameter struct {
		Name string `+"`json:\"name\"`"+`
	}
	HelloAction struct {
		action.BaseAction
	}
)

func (h *HelloAction) Default() format.ResultMessage {
	return format.ResultMessage{
		Data:    make([]any, 0),
		Message: "Unknown Error",
		State:   format.Error,
	}
}

func (h *HelloAction) Run(input string, config format.PluginConfig) format.ResultMessage {
	var parameter HelloActionParameter
	err := h.DecodeInput(input, &parameter)
	if err != nil {
		return format.ResultMessage{
			Data:    make([]any, 0),
			Message: err.Error(),
			State:   format.Error,
		}
	}
	return format.ResultMessage{
		Data:    []any{fmt.Sprintf("Hello, %%v", parameter.Name)},
		Message: "",
		State:   format.Success,
	}
}

func NewHelloAction() *HelloAction {
	return &HelloAction{action.NewBaseAction()}
}
`, modulePath, projectName)), 0755)
}

func createProjectActionFolder(targetDir, projectName string) (string, error) {
	projDir := path.Join(targetDir, "pgapi", projectName)
	return projDir, os.MkdirAll(projDir, 0755)
}

func createApiFile(targetDir string) error {
	return os.WriteFile(path.Join(targetDir, "api.go"), []byte(fmt.Sprintf(`package pgapi

import "C"

var Name = "pgapi"

//export GoDispatch
func GoDispatch(name, input, cfg *C.char) *C.char {
	in := C.GoString(input)
	config := C.GoString(cfg)
	n := C.GoString(name)
	out := Handle(n, in, config)
	return C.CString(out)
}`)), 0755)
}

func createWrapperFile(targetDir, modulePath, projectName string) error {
	return os.WriteFile(path.Join(targetDir, "wrapper.go"), []byte(fmt.Sprintf(`package pgapi

import (
	"encoding/json"
	"fmt"
	"%[1]v/%[2]v/pgapi/format"
	"%[1]v/%[2]v/pgapi/%[2]v"
	"runtime/debug"
)

type (
	IAction interface {
		Default() format.ResultMessage
		Run(input string, config format.PluginConfig) format.ResultMessage
	}
)

var actions = map[string]IAction{
	"hello": %[2]v.NewHelloAction(),
}

func Handle(name, input, cfg string) string {
	var str []byte
	var err error

	// catch any Error
	defer func() {
		if r := recover(); r != nil {
			errorText := fmt.Sprintf("[recovered]: %%v stack: %%v", r, string(debug.Stack()))
			fmt.Println(errorText)
		}
	}()

	// parse config for run parameter
	var pluginConfig format.PluginConfig
	err = json.Unmarshal([]byte(cfg), &pluginConfig)
	if err != nil {
		str, err = json.Marshal(format.ResultMessage{
			Data:    make([]any, 0),
			Message: fmt.Sprintf("invalid plugin config cannot parse %%v: %%v", cfg, err.Error()),
			State:   format.Error,
		})
		if err != nil {
			return fmt.Sprintf("{\"data\":[],\"message\":\"%%v\",\"state\":-1}", err.Error())
		}
		return string(str)
	}

	// search the action
	action := actions[name]
	if action == nil {
		str, err = json.Marshal(format.ResultMessage{
			Data:    make([]any, 0),
			Message: fmt.Sprintf("action %%v not found", name),
			State:   format.Error,
		})
		if err != nil {
			return fmt.Sprintf("{\"data\":[],\"message\":\"%%v\",\"state\":-1}", err.Error())
		}
		return string(str)
	}

	// execute the action
	var res format.ResultMessage
	res = action.Run(input, pluginConfig)
	str, err = json.Marshal(res)
	if err != nil {
		return fmt.Sprintf("{\"data\":[],\"message\":\"%%v\",\"state\":-1}", err.Error())
	}
	return string(str)
}
`, modulePath, projectName)), 0755)
}

func createPgapiFolder(targetDir string) (string, error) {
	pgapiDir := path.Join(targetDir, "pgapi")
	return pgapiDir, os.MkdirAll(pgapiDir, 0755)
}

func createScriptFile(targetDir, projectName string) error {
	scriptDir := path.Join(targetDir, "scripts")
	err := os.MkdirAll(scriptDir, 0755)
	if err != nil {
		return err
	}
	return os.WriteFile(path.Join(scriptDir, "main.sql"), []byte(fmt.Sprintf(`CREATE FUNCTION %[1]v.dispatch(text, text, text) RETURNS text
AS
'$libdir/%[1]v',
'dispatch'
LANGUAGE C IMMUTABLE STRICT;`, projectName)), 0755)
}

func createSqlMergerFile(targetDir, modulePath, projectName string) error {
	sqlMergerDir := path.Join(targetDir, "sqlmerger")
	err := os.MkdirAll(sqlMergerDir, 0755)
	if err != nil {
		return err
	}
	return os.WriteFile(path.Join(sqlMergerDir, "main.go"), []byte(`package main

import (
	"fmt"
	"`+modulePath+"/"+projectName+`/meta"
	"os"
	"path"
	"strings"
)

var header = fmt.Sprintf(`+"`"+`-- complain if script is sourced in psql, rather than via CREATE EXTENSION"+
	\echo Use "CREATE EXTENSION %v" to load this file. \quit
	`+"`"+`, meta.AppName)
var aprasterFile = fmt.Sprintf("%v.control", meta.AppName)

var targetFile = fmt.Sprintf("%v--%v.sql", meta.AppName, meta.Version)
var aprasterControl = fmt.Sprintf(`+"`"+`
	# %[3]v extension
	comment = '%[1]v'
	default_version = '%[2]v'
	relocatable = false
	schema = %[3]v
	`+"`"+`, meta.Description, meta.Version, meta.AppName)

func writeControlFile(filePath string) error {
	return os.WriteFile(filePath, []byte(aprasterControl), 0755)
}

func collectSql(folder, content string) (string, error) {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return content, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			tmp, _ := collectSql(path.Join(folder, entry.Name()), content)
			content += tmp
			continue
		}
		if strings.HasSuffix(entry.Name(), ".sql") {
			var str []byte
			str, err = os.ReadFile(path.Join(folder, entry.Name()))
			if err == nil {
				content += fmt.Sprintf("\n-- %v\n\n%v", entry.Name(), string(str))
			}
		}
	}
	return content, nil
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	targetDir := path.Join(wd, "bin")
	err = os.RemoveAll(targetDir)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		panic(err)
	}
	combined, err := collectSql(path.Join(wd, "scripts"), header)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(path.Join(targetDir, targetFile), []byte(combined), 0755)
	if err != nil {
		panic(err)
	}
	err = writeControlFile(path.Join(targetDir, aprasterFile))
	if err != nil {
		panic(err)
	}
}
`), 0755)
}

func createMainFile(targetDir, projectName, modulePath, pgVersion string) error {
	return os.WriteFile(path.Join(targetDir, "main.go"), []byte(fmt.Sprintf(`package main

//go:generate go run ./sqlmerger/main.go
//go:generate go build -o ./bin/%[1]v.so -buildmode=c-shared ./main.go

// #cgo CFLAGS: -I /usr/include/postgresql/%[3]v/server/
// #cgo LDFLAGS: -Wl,-unresolved-symbols=ignore-all
/*
extern char* GoDispatch(char*, char*, char*);

#include "postgres.h"
#include "fmgr.h"
#include "utils/builtins.h"

PG_MODULE_MAGIC;

PG_FUNCTION_INFO_V1(dispatch);

Datum
dispatch(PG_FUNCTION_ARGS)
{
	// read the first argument as postgres text and convert it to a C String
    char* name = text_to_cstring(PG_GETARG_TEXT_PP(0));
    char* input = text_to_cstring(PG_GETARG_TEXT_PP(1));
    char* cfg = text_to_cstring(PG_GETARG_TEXT_PP(2));
    // give the C String into the go function and get the result as C String
    char* res = GoDispatch(name, input, cfg);
    // convert the result C String into Postgres Text and mark it as return value
    // it returns the value at the end of the function automatically
    PG_RETURN_TEXT_P(cstring_to_text(res));
    // take care to delete the 2 variables at the end of the transaction
    // pfree is from Postgres
	pfree(name);
	pfree(input);
	pfree(cfg);
	pfree(res);
}
*/
import "C"
import "%[2]v/%[1]v/pgapi"

var _ = pgapi.Name

func main() {}
`, projectName, modulePath, pgVersion)), 0755)
}

func createGitIgnoreFile(targetDir string) error {
	return os.WriteFile(path.Join(targetDir, ".gitignore"), []byte("bin"), 0755)
}

func createModFile(targetDir, projectName, modulePath, goVersion string) error {
	return os.WriteFile(path.Join(targetDir, "go.mod"), []byte(fmt.Sprintf(`module %[1]v/%[2]v

go %[3]v

require github.com/go-playground/validator v9.31.0+incompatible

require (
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
)`, modulePath, projectName, goVersion)), 0755)
}

func createProjectFolder(projectName string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	targetDir := path.Join(wd, projectName)
	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		return err
	}
	return nil
}

var Create = &cobra.Command{
	Use:   "create [go module path] [go version] [postgres version] [description]",
	Short: "create a new postgres application",
	Long:  `create a new postgres application and creates the project structure`,
	Args:  cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		goModulePath := args[0]
		goVersion := args[1]
		pgVersion := args[2]
		description := ""
		if len(args) > 3 {
			description = args[3]
		}
		tmp := strings.Split(goModulePath, "/")
		projectName := tmp[len(tmp)-1]
		modulePath := strings.Join(tmp[:len(tmp)-1], "/")
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		targetDir := path.Join(wd, projectName)
		fmt.Println("ProjectName: " + projectName)
		fmt.Println("ModulPath: " + modulePath)
		fmt.Println("PG Version: " + pgVersion)
		fmt.Println("Go Version: " + goVersion)
		fmt.Println("TargetDir: " + targetDir)
		fmt.Println("Description: " + description)

		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			fmt.Println("create Project Folder...")
			err = createProjectFolder(projectName)
			if err != nil {
				panic(err)
			}
			fmt.Println("init Go Module...")
			err = createModFile(targetDir, projectName, modulePath, goVersion)
			if err != nil {
				panic(err)
			}
			fmt.Println("create main.go...")
			err = createMainFile(targetDir, projectName, modulePath, pgVersion)
			if err != nil {
				panic(err)
			}
			fmt.Println("create .gitignore...")
			err = createGitIgnoreFile(targetDir)
			if err != nil {
				panic(err)
			}

			fmt.Println("create sqlmerger...")
			err = createSqlMergerFile(targetDir, modulePath, projectName)
			if err != nil {
				panic(err)
			}
			fmt.Println("create scripts...")
			err = createScriptFile(targetDir, projectName)
			if err != nil {
				panic(err)
			}

			var pgapiDir string
			fmt.Println("create pgapi...")
			pgapiDir, err = createPgapiFolder(targetDir)
			if err != nil {
				panic(err)
			}
			fmt.Println("create wrapper.go...")
			err = createWrapperFile(pgapiDir, modulePath, projectName)
			if err != nil {
				panic(err)
			}
			fmt.Println("create api.go...")
			err = createApiFile(pgapiDir)
			if err != nil {
				panic(err)
			}

			var projectDir string
			fmt.Println("create action folder...")
			projectDir, err = createProjectActionFolder(targetDir, projectName)
			if err != nil {
				panic(err)
			}
			fmt.Println("create hello.go...")
			err = createProjectActionFile(projectDir, modulePath, projectName)
			if err != nil {
				panic(err)
			}

			var formatFolder string
			fmt.Println("create format folder...")
			formatFolder, err = createFormatFolder(targetDir)
			if err != nil {
				panic(err)
			}
			fmt.Println("create plugin_config.go...")
			err = createFormatPluginConfigFile(formatFolder)
			if err != nil {
				panic(err)
			}
			fmt.Println("create state.go...")
			err = createFormatStateFile(formatFolder)
			if err != nil {
				panic(err)
			}
			fmt.Println("create result_message.go...")
			err = createFormatResultFile(formatFolder)
			if err != nil {
				panic(err)
			}

			fmt.Println("create app folder...")
			err = createActionBaseFile(targetDir, modulePath, projectName)
			if err != nil {
				panic(err)
			}

			var metaDir string
			fmt.Println("create meta folder...")
			metaDir, err = createMetaFolder(targetDir)
			if err != nil {
				panic(err)
			}
			fmt.Println("create app.go...")
			err = createMetaAppFile(metaDir, projectName, description)
			if err != nil {
				panic(err)
			}
			fmt.Println("create version.go...")
			err = createMetaVersionFile(metaDir)
			if err != nil {
				panic(err)
			}

			tidy := exec.Command("go", "mod", "tidy")
			tidy.Dir = targetDir
			err = tidy.Run()
			if err != nil {
				panic(err)
			}
		} else {
			panic(fmt.Errorf("folder already exists %v", targetDir))
		}
	},
}
