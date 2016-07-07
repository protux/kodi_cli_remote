package kodicommunicator

import (
    "administration"

    "encoding/json"
    "errors"
    "io/ioutil"
    "net/http"
    "strings"
    "strconv"
)

// ErrorResponse is the foundation for the JSONRPC
// Error returned by Kodi if something happened.
type ErrorResponse struct {
    Error Error `json:"error"`
    ID int `json:"id"`
    JsonRPC string `json:"jsonrpc"`
}

// Error is the part of the JSONRPC-Errorresponse 
// which contains the error-data.
type Error struct {
    Code int `json:"code"`
    Data Data `json:"data"`
    Message string `json:"message"`
}

// Data contains the parts of the request which
// caused the error.
type Data struct {
    Message string `json:"message"`
    Method string `json:"method"`
    Stack Stack `json:"stack"`
}

// Stack contains detailed information about the 
// problem causing elements of the request.
type Stack struct {
    Name string `json:"name"`
    Type string `json:"type"`
    Message string `json:"message"`
}

// Command represents a command which can be sent to Kodi. 
// It also represents a documentation and a translation from CLI-command
// to a command Kodi understands.
type Command struct {
    CliName string
    KodiName string
    Description string
    ParametersDescription map[string]string
    CreateParameterMap func(params []string) (map[string]interface{}, error)
}

// CommandRequest represents all parameters of a JSONRPC call.
type CommandRequest struct {
    JSONrpc string `json:"jsonrpc"`
    Method string `json:"method"`
    Params map[string]interface{} `json:"params,omitempty"`
    ID int `json:"id"`
}

// SetValues sets the method and the parameters for the JSONRPC call.
func (self *CommandRequest) SetValues(method string, params map[string]interface{}) {
    self.JSONrpc = `2.0`
    self.ID = 1
    self.Method = method
    self.Params = params
}

var (
    CommandMap = map[string]*Command {
        // Player 
        `play`: &Command {
            CliName: `play`, 
            KodiName: `Player.PlayPause`, 
            Description: `Resumes the current playback from pause state.`,
            ParametersDescription: map[string]string {},
            CreateParameterMap: func(params []string) (map[string]interface{}, error) {
                return map[string]interface{} {
                    `playerid`:1,
                }, nil
            },
        },
        `pause`: &Command {
            CliName: `pause`, 
            KodiName: `Player.PlayPause`, 
            Description: `Pauses the current playback.`,
            ParametersDescription: map[string]string {},
            CreateParameterMap: func(params []string) (map[string]interface{}, error) {
                return map[string]interface{} {
                    `playerid`:1,
                }, nil
            },
        },
        `stop`: &Command {
            CliName: `stop`, 
            KodiName: `Player.Stop`, 
            Description: `Stops the current playback.`,
            ParametersDescription: map[string]string {},
            CreateParameterMap: func(params []string) (map[string]interface{}, error) {
                return map[string]interface{} {
                    `playerid`:1,
                }, nil
            },
        },
        `mute`: &Command {
            CliName: `mute`, 
            KodiName: `Application.SetMute`, 
            Description: `Mutes or unmutes the audio.`,
            ParametersDescription: map[string]string {},
            CreateParameterMap: func(params []string) (map[string]interface{}, error) {
                return map[string]interface{} {
                    `mute`:`toggle`,
                }, nil
            },
        },
        `seek`: &Command {
            CliName: `seek`, 
            KodiName: `Player.Seek`, 
            Description: `Jumps to the given time.`,
            ParametersDescription: map[string]string {
                `-/+`: `Jump back/forth n seconds.`,
                `--/++`: `Jump back/forth n seconds.`,
                `[hh:]mm:ss`: `Junp to hours:minutes:seconds (hours optional)`,
            },
            CreateParameterMap: func(params []string) (map[string]interface{}, error) {
                if len(params) < 1 {
                    return map[string]interface{}{}, errors.New(`Not enough parameters. See "help seek" for usage information.`)
                }
                var val string
                if params[0] == `+` {
                    val = `smallforward`
                } else if params[0] == `++` {
                    val = `bigforward`
                } else if params[0] == `-` {
                    val = `smallbackward`
                } else if params[0] == `--` {
                    val = `bigbackward`
                } 
                if len(val) > 0 {
                    return map[string]interface{} {
                        `playerid`:1,
                        `value`:val,
                    }, nil
                }
                
                timeMap := map[string]int {
                    `hours`: 0,
                    `minutes`: 0,
                    `seconds`: 0,
                    `milliseconds`: 0,
                }
                hms := strings.Split(params[len(params) - 1], `:`)
                if len(hms) == 3 {
                    hours, err := parseTimeNumber(hms[0])
                    if err != nil {
                        return nil, err
                    }
                    timeMap[`hours`] = hours
                    hms = hms[1:]
                }
                if len(hms) == 2 {
                    minutes, err := parseTimeNumber(hms[0])
                    if err != nil {
                        return nil, err
                    }
                    seconds, err := parseTimeNumber(hms[1])
                    if err != nil {
                        return nil, err
                    }
                    timeMap[`minutes`] = minutes
                    timeMap[`seconds`] = seconds
                    return map[string]interface{} {
                        `playerid`:1,
                        `value`:timeMap,
                    }, nil
                }
                return map[string]interface{}{}, errors.New(`Illegal parameter. See "help seek" for usage information.`)
            },
        },
        `speed`: &Command {
            CliName: `speed`, 
            KodiName: `Player.Speed`, 
            Description: `Set the playback speed.`,
            ParametersDescription: map[string]string {
                `speed`: `Speed as integer`,
            },
            CreateParameterMap: func(params []string) (map[string]interface{}, error) {
                return map[string]interface{}{
                    `playerid`:1,
                    `speed`:params[0],
                }, nil
            },
        },
        
        // Input
        `action`: &Command {
            CliName: `action`, 
            KodiName: `Input.Select`, 
            Description: `Selects the current selection.`,
            ParametersDescription: map[string]string {},
            CreateParameterMap: func(params []string) (map[string]interface{}, error) {
                return map[string]interface{}{}, nil
            },
        },
        `context`: &Command {
            CliName: `context`, 
            KodiName: `Input.ContextMenu`, 
            Description: `Opens the context menu.`,
            ParametersDescription: map[string]string {},
            CreateParameterMap: func(params []string) (map[string]interface{}, error) {
                return map[string]interface{}{}, nil
            },
        },
        `info`: &Command {
            CliName: `info`, 
            KodiName: `Input.Info`, 
            Description: `Opens the info view.`,
            ParametersDescription: map[string]string {},
            CreateParameterMap: func(params []string) (map[string]interface{}, error) {
                return map[string]interface{}{}, nil
            },
        },
        `home`: &Command {
            CliName: `home`, 
            KodiName: `Input.Home`, 
            Description: `Returns to the home screen.`,
            ParametersDescription: map[string]string {},
            CreateParameterMap: func(params []string) (map[string]interface{}, error) {
                return map[string]interface{}{}, nil
            },
        },
        `back`: &Command {
            CliName: `back`, 
            KodiName: `Input.Back`, 
            Description: `Returns to the previous view.`,
            ParametersDescription: map[string]string {},
            CreateParameterMap: func(params []string) (map[string]interface{}, error) {
                return map[string]interface{}{}, nil
            },
        },
        `left`: &Command {
            CliName: `left`, 
            KodiName: `Input.Left`, 
            Description: `Sends the cursor one item to the left`,
            ParametersDescription: map[string]string {},
            CreateParameterMap: func(params []string) (map[string]interface{}, error) {
                return map[string]interface{}{}, nil
            },
        },
        `right`: &Command {
            CliName: `right`, 
            KodiName: `Input.Right`, 
            Description: `Sends the cursor one item to the right.`,
            ParametersDescription: map[string]string {},
            CreateParameterMap: func(params []string) (map[string]interface{}, error) {
                return map[string]interface{}{}, nil
            },
        },
        `up`: &Command {
            CliName: `up`, 
            KodiName: `Input.Up`, 
            Description: `Sends the cursor one item up.`,
            ParametersDescription: map[string]string {},
            CreateParameterMap: func(params []string) (map[string]interface{}, error) {
                return map[string]interface{}{}, nil
            },
        },
        `down`: &Command {
            CliName: `down`, 
            KodiName: `Input.Down`, 
            Description: `Sends the cursor one item down.`,
            ParametersDescription: map[string]string {},
            CreateParameterMap: func(params []string) (map[string]interface{}, error) {
                return map[string]interface{}{}, nil
            },
        },
        
        // 
        `notify`: &Command {
            CliName: `notify`, 
            KodiName: `GUI.ShowNotification`, 
            Description: `Displays a notification on the screen.`,
            ParametersDescription: map[string]string {
                `title`: `The title of the notification.`,
                `message`: `The message of the notification.`,
                `displaytime`: `(optional) The time in milliseconds the notification is displayed.`,
            },
            CreateParameterMap: func(params []string) (map[string]interface{}, error) {
                return map[string]interface{}{}, nil
            },
        },
        `clean`: &Command {
            CliName: `clean`, 
            KodiName: `VideoLibrary.Clean`, 
            Description: `Cleans the video library from non-existent items.`,
            ParametersDescription: map[string]string {},
            CreateParameterMap: func(params []string) (map[string]interface{}, error) {
                return map[string]interface{}{}, nil
            },
        },
        `update`: &Command {
            CliName: `update`, 
            KodiName: `VideoLibrary.Scan`, 
            Description: `Scans the video sources for new library items.`,
            ParametersDescription: map[string]string {},
            CreateParameterMap: func(params []string) (map[string]interface{}, error) {
                return map[string]interface{}{}, nil
            },
        },
    }
)

// parseTimeNumber parses a number and makes sure that
// the number is >= 0 and <= 59
func parseTimeNumber(number string) (int, error) {
    num, err := strconv.Atoi(number)
    if err != nil {
        return 0, err
    } else if num > 59 || num < 0 {
        return 0, errors.New(`A time-number needs to be between 0 and 59, but was ` + number)
    } else {
        return num, nil
    }
}

// GetCommandForName returns a copy of the Command related to the CliName passed
// if it exists. 
func GetCommandForName(cmd string) (Command, bool) {
    command, success := CommandMap[cmd]
    return *command, success 
}

// getRepeatCount returns for some allowed actions the number how often this action
// should be executed.
func getRepeatCount(action string, params *[]string) int {
    if len(*params) > 0 && (action == `down` || action == `up` || action == `left` || action == `right`) {
        num, err := strconv.Atoi((*params)[len(*params) - 1])
        if err != nil || num < 1 {
            return 1
        }
        *params = (*params)[:len(*params) - 1]
        return num
    }
    return 1
}

// ExecuteCommand takes the action, looks up the appropriate JSON-RPC command
// and sends the request to the configured address.
func ExecuteCommand(config administration.Configuration, action string, params []string) error {
    repeatCount := getRepeatCount(action, &params)
    cmd, err := createJsonCommand(action, params)
    if err == nil {
        for i := 0; i < repeatCount; i++ {
            err = sendRequest(config.Host, config.Port, cmd)
        }
        return err
    } else {
        return err
    }
}

// sendRequest actually sends the request to Kodi.
func sendRequest(host, port, js string) error {

    requestURL := `http://` + host + `:` + port + `/jsonrpc`
    if request, err := http.NewRequest(`POST`, requestURL, strings.NewReader(js)); err == nil {
        var header http.Header = map[string][]string{}
        header.Add(`Content-Type`, `application/json`)
        request.Header = header
        var client http.Client

        if response, err := client.Do(request); err == nil {
            defer response.Body.Close()

            if resp, err := ioutil.ReadAll(response.Body); err == nil {
                
                var errorResponse ErrorResponse
                if err = json.Unmarshal(resp, &errorResponse); err == nil {
                    if errorResponse.Error.Code != 0 {
                        return createJsonError(errorResponse)
                    }
                } else {
                    return err
                }
            } else {
                return err
            }
        } else {
            return err
        }
    } else {
        return err
    }
    return nil
}

// createJsonError creates a more readable message from an ErrorResponse
func createJsonError(errorResponse ErrorResponse) error {
    var message string = ``
    if errorResponse.Error.Data.Message != `` {
        message += errorResponse.Error.Data.Message + ` `
    }
    if errorResponse.Error.Data.Stack.Message != `` {
        message += errorResponse.Error.Data.Stack.Message + ` regarding ` 
    }
    if errorResponse.Error.Data.Stack.Name != `` {
        message += `parameter "` + errorResponse.Error.Data.Stack.Name + `" `
    }
    if errorResponse.Error.Data.Stack.Type != `` {
        message += `of type "` + errorResponse.Error.Data.Stack.Type + `"`
    }
    return errors.New(message)
}

// createJsonCommand takes the action and the params and creates a Command.
// If the Command was created successfully the first return value will be the
// JSON and the second nil, otherwise the first one will be nil and the second
// one will be an error message.
func createJsonCommand(action string, params []string) (string, error) {
    var command CommandRequest
    cmd, success := CommandMap[action]
    
    if success {
        paramMap, err := cmd.CreateParameterMap(params)
        if err != nil {
            return ``, err
        }
        command.SetValues(cmd.KodiName, paramMap)
        output, err := json.Marshal(command)
        
        if err == nil {
            return string(output), nil
        } else {
            return ``, err
        }
    } else {
        return ``, errors.New("The Command " + action + " is unknown.")
    }
}

