match-versions will match the versions of one go.mod file to another.

**make sure you are checked out at the correct commit of the "get" module**

`go run main.go --get absolute/path/to/module --set absolute/path/to/module/you/want/to/change`

this will iterate through the 'get' and 'set' go.mod files, create a map of packages to versions for each, then replace any of the 'set' package versions with any 'get' package versions that overlap
