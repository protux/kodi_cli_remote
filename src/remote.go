package main

import (
    "errors"
    "fmt"
    "os"
    "strings"
    
    "administration"
    "kodicommunicator"
)

func checkAndHandleArgumentsConfig(configuration *administration.Configuration, args []string) bool {
    changed := false
    
    for _, arg := range args {
        if strings.HasPrefix(arg, "--host=") {
            configuration.Host = strings.Split(arg, `=`)[1]
            changed = true
        } else if strings.HasPrefix(arg, "--port=") {
            configuration.Port = strings.Split(arg, `=`)[1]
            changed = true
        }
    }
    return changed
}

func splitParameterIntoMap(args []string) map[string]interface{} {
    params := map[string]interface{}{}
    
    if len(args) > 1 {
        paramPairs := strings.Split(args[1], ",")
        for _, paramPair := range paramPairs {
            pair := strings.Split(paramPair, ":")
            params[pair[0]] = pair[1]
        }
    }
    return params
}

func printHelp(args []string) {
    fmt.Println(`If you run the tool the first time you need to configure it. Therefore you need to call it with the parameters --host=<kodi-address> and --port=<kodi-port>.`)
    fmt.Println()
    fmt.Println(`If the tool is properly configured you can just run it by passing the name of the command as the first parameter and as the second parameter the parameter for the command.`)
    fmt.Println(`The command-params need to be passed like "title:test123,message:I'm here!" so a complete call would look like 'krm notify "title:test123,I'm here!'`)
    printUsage(args)
}

func checkAndPrintHelp(args []string) bool {
    for idx, arg := range args {
        if arg == `help` {
            if idx < len(args) - 1 {
                command, success := kodicommunicator.GetCommandForName(args[idx + 1])
                if success {
                    fmt.Println(`Help for command`, command.CliName)
                    fmt.Println(`Description:`, command.Description)
                    if len(command.ParametersDescription) > 0 {
                        fmt.Println(`Parameters`)
                        for param, desc := range command.ParametersDescription {
                            fmt.Println(param, `-`, desc)
                        }
                    }
                } else {
                    fmt.Println("The command", arg, "is not supported.")
                }
            } else {
                printHelp(args);
            }
            return true
        }
    }
    return false
}

func printUsage(args []string) {
    fmt.Println(`Usage:`, args[0], `command [paramter]`)
    fmt.Println(`Parameters are entered as follows: "key1:value,key2:value"`)
    fmt.Println(`To get help type`, args[0], `help`)
    fmt.Println(`To get help for a specific command type`, args[0], `help <command>`)
    fmt.Println()
    fmt.Println(`List of all available commands:`)
    for key, value := range kodicommunicator.CommandMap {
        fmt.Println(key, `-`, value.Description)
    }
}

func main() {
    if len(os.Args) < 2 {
        printUsage(os.Args)
    } else if !checkAndPrintHelp(os.Args) {
        args := os.Args[1:]
        config, err := administration.CreateConfiguration()
        
        
        if err == nil {
            if checkAndHandleArgumentsConfig(&config, args) {
                if err := administration.WriteConfiguration(config); err != nil {
                    fmt.Println(err.Error())
                }
            } else {
                if len(config.Host) == 0 {
                    err = errors.New(`No host configured. Please see "help" to learn about how to configure the remote.`)
                }
        
                err := kodicommunicator.ExecuteCommand(config, args[0], args[1:])
                if err != nil {   
                    fmt.Println(err.Error())
                }
            }
        } else {
            fmt.Println(err.Error())
        }
    }
}

