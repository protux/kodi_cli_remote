# Kodi Remote
A CLI based remote for Kodi using the JSONRPC API.

## Dependencies
In order to compile this project you need to run `go get github.com/mitchellh/go-homedir`

## Usage
Usage: `krm command [paramters]`
Parameters are entered as follows: `"key1:value,key2:value"`
To get help type `krm help`
To get help for a specific command type `krm help <command>`

## Todo
* Display returned data so the user will be informed about JSONRPC errors etc.
* Write tests
* Implement more commands
