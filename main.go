package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var log = logrus.New()

// Rootcmd is the root command

var RootCmd = &cobra.Command{
	Use:   "match-versions",
	Short: "match-versions matches the go mod versions in one module to another",
	Long: `match-versions takes a path to a go module (a package that has a 
go.mod file) and changes the current module's go.mod file to import 
the versions specified in the given module.`,
	Run: Match,
}

func init() {
	RootCmd.Flags().String("get", "", "path to module from which we are getting the versions")
	RootCmd.Flags().String("set", "", "path of module to which we are setting the versions")
}
func main() {
	if err := RootCmd.Execute(); err != nil {
		log.Error(err)
	}
}

// Match is the command for matching and updating go.mod files
func Match(cmd *cobra.Command, args []string) {
	getPath, err := cmd.Flags().GetString("get")
	if err != nil {
		log.Error(err)
		return
	}
	setPath, err := cmd.Flags().GetString("set")
	if err != nil {
		log.Error(err)
		return
	}
	if getPath == "" || setPath == "" {
		log.Error("the 'get' flag and 'set' flag must both be set")
		return
	}

	if _, err := os.Stat(getPath); os.IsNotExist(err) {
		log.Errorf("directory '%s' does not exist", getPath)
		return
	}
	getGoModPath := filepath.Join(getPath, "go.mod")
	if _, err := os.Stat(getGoModPath); os.IsNotExist(err) {
		log.Errorf("go.mod file does not exist at '%s'", getPath)
		return
	}

	if _, err := os.Stat(setPath); os.IsNotExist(err) {
		log.Errorf("directory '%s' does not exist", setPath)
		return
	}
	setGoModPath := filepath.Join(setPath, "go.mod")
	if _, err := os.Stat(setGoModPath); os.IsNotExist(err) {
		log.Errorf("go.mod file does not exist at '%s'", setPath)
		return
	}

	// parse gomod file and print all requireds
	getMap, _, err := parseFileToRequireMap(getGoModPath)
	if err != nil {
		log.Errorf("error parsing get go.mod file: %s", err)
		return
	}

	setMap, setKeys, err := parseFileToRequireMap(setGoModPath)
	if err != nil {
		log.Errorf("error parsing set go.mod file: %s", err)
	}
	for _, key := range setKeys {
		val, ok := getMap[key]
		if !ok {
			continue
		}
		setMap[key] = val
	}
	encodeModFile(setGoModPath, setMap, setKeys)
}

func parseFileToRequireMap(path string) (map[string]string, []string, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	modString := string(bytes)
	openIndex := strings.Index(modString, "require (")
	if openIndex < 0 {
		return nil, nil, fmt.Errorf("could not find the list of required imports")
	}
	modString = modString[openIndex+9:]
	closeIndex := strings.Index(modString, ")")
	if closeIndex < 0 {
		return nil, nil, fmt.Errorf("go.mod file contains formatting errors")
	}
	modString = modString[:closeIndex]
	replaceArray := strings.Split(strings.TrimSpace(modString), "\n")
	replaceMap := map[string]string{}
	for i, replace := range replaceArray {
		split := strings.SplitN(strings.TrimSpace(replace), " ", 2)
		replaceArray[i] = split[0]
		replaceMap[split[0]] = split[1]
	}
	return replaceMap, replaceArray, nil
}

func encodeModFile(path string, requireMap map[string]string, requireKeys []string) error {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	modString := string(bytes)
	openIndex := strings.Index(modString, "require (")
	if openIndex < 0 {
		return fmt.Errorf("could not find the list of required imports")
	}
	start := modString[:openIndex+9]
	closeIndex := strings.Index(modString, ")")
	if closeIndex < 0 {
		return fmt.Errorf("go.mod file contains formatting errors")
	}
	end := modString[closeIndex:]
	require := ""
	for _, key := range requireKeys {
		require += fmt.Sprintf("\n\t%s %s", key, requireMap[key])
	}
	require += "\n"
	start += require
	start += end
	return ioutil.WriteFile(path, []byte(start), 0644)
}
