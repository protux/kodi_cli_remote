package kodicommunicator

import (
    "administration"

    "encoding/json"
    "errors"
    "io/ioutil"
    "net/http"
    "strings"
)

// Command represents a command which can be sent to Kodi. 
// It also represents a documentation and a translation from CLI-command
// to a command Kodi understands.
type Command struct {
    CliName string
    KodiName string
    Description string
    Parameters map[string]string
}

// CommandRequest represents all parameters of a JSONRPC call.
type CommandRequest struct {
    JSONrpc string `json:"jsonrpc"`
    Method string `json:"method"`
    Params map[string]string `json:"params,omitempty"`
    ID int `json:"ID"`
}

// SetValues sets the method and the parameters for the JSONRPC call.
func (self *CommandRequest) SetValues(method string, params map[string]string) {
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
            KodiName: `Input.PlayPause`, 
            Description: `Resumes the current playback from pause state.`,
            Parameters: map[string]string {}, 
        },
        `pause`: &Command {
            CliName: `pause`, 
            KodiName: `Input.PlayPause`, 
            Description: `Pauses the current playback.`,
            Parameters: map[string]string {}, 
        },
        `stop`: &Command {
            CliName: `stop`, 
            KodiName: `Input.Stop`, 
            Description: `Stops the current playback.`,
            Parameters: map[string]string {}, 
        },
        `mute`: &Command {
            CliName: `mute`, 
            KodiName: `Application.SetMute`, 
            Description: `Mutes or unmutes the audio.`,
            Parameters: map[string]string {}, 
        },
        `seek`: &Command {
            CliName: `seek`, 
            KodiName: `Player.Seek`, 
            Description: `Jumps to the given time.`,
            Parameters: map[string]string {
                `percentage`: `(optional) the time in percent of the total time.`,
                `time`: `(optional) the relative time to jump`,
                `totaltime`: `(optional) the total time to jump to.`,
            }, 
        },
        `speed`: &Command {
            CliName: `speed`, 
            KodiName: `Player.Speed`, 
            Description: `Set the playback speed.`,
            Parameters: map[string]string {
                `speed`: `Speed as integer`,
            }, 
        },
        
        // Input
        `action`: &Command {
            CliName: `action`, 
            KodiName: `Input.Select`, 
            Description: `Selects the current selection.`,
            Parameters: map[string]string {}, 
        },
        `context`: &Command {
            CliName: `context`, 
            KodiName: `Input.ContextMenu`, 
            Description: `Opens the context menu.`,
            Parameters: map[string]string {}, 
        },
        `info`: &Command {
            CliName: `info`, 
            KodiName: `Input.Info`, 
            Description: `Opens the info view.`,
            Parameters: map[string]string {}, 
        },
        `home`: &Command {
            CliName: `home`, 
            KodiName: `Input.Home`, 
            Description: `Returns to the home screen.`,
            Parameters: map[string]string {}, 
        },
        `back`: &Command {
            CliName: `back`, 
            KodiName: `Input.Back`, 
            Description: `Returns to the previous view.`,
            Parameters: map[string]string {}, 
        },
        `left`: &Command {
            CliName: `left`, 
            KodiName: `Input.Left`, 
            Description: `Sends the cursor one item to the left`,
            Parameters: map[string]string {}, 
        },
        `right`: &Command {
            CliName: `right`, 
            KodiName: `Input.Right`, 
            Description: `Sends the cursor one item to the right.`,
            Parameters: map[string]string {}, 
        },
        `up`: &Command {
            CliName: `up`, 
            KodiName: `Input.Up`, 
            Description: `Sends the cursor one item up.`,
            Parameters: map[string]string {}, 
        },
        `down`: &Command {
            CliName: `down`, 
            KodiName: `Input.Down`, 
            Description: `Sends the cursor one item down.`,
            Parameters: map[string]string {}, 
        },
        
        // 
        `notify`: &Command {
            CliName: `notify`, 
            KodiName: `GUI.ShowNotification`, 
            Description: `Displays a notification on the screen.`,
            Parameters: map[string]string {
                `title`: `The title of the notification.`,
                `message`: `The message of the notification.`,
                `displaytime`: `(optional) The time in milliseconds the notification is displayed.`,
            }, 
        },
        `clean`: &Command {
            CliName: `clean`, 
            KodiName: `VideoLibrary.Clean`, 
            Description: `Cleans the video library from non-existent items.`,
            Parameters: map[string]string {}, 
        },
        `update`: &Command {
            CliName: `update`, 
            KodiName: `VideoLibrary.Scan`, 
            Description: `Scans the video sources for new library items.`,
            Parameters: map[string]string {}, 
        },
    }
)

// GetCommandForName returns a copy of the Command related to the CliName passed
// if it exists. 
func GetCommandForName(cmd string) (Command, bool) {
    command, success := CommandMap[cmd]
    return *command, success 
}

// ExecuteCommand takes the action, looks up the appropriate JSON-RPC command
// and sends the request to the configured address.
func ExecuteCommand(config administration.Configuration, action string, params map[string]string) error {
    cmd, err := createJsonCommand(action, params)
    if err == nil {
        return sendRequest(config.Host, config.Port, cmd)
    } else {
        return err
    }
}

func sendRequest(host, port, js string) error {
    if request, err := http.NewRequest(`POST`, `http://` + host + `:` + port + `/jsonrpc`, strings.NewReader(js)); err == nil {
        var header http.Header = map[string][]string{}
        header.Add(`Content-Type`, `application/json`)
        request.Header = header
        
        var client http.Client
        if response, err := client.Do(request); err == nil {
            defer response.Body.Close()
            _, err = ioutil.ReadAll(response.Body)
            return nil
        }
        return err // TODO handle response
    } else {
        return err
    }
}

// createJsonCommand takes the action and the params and creates a Command.
// If the Command was created successfully the first return value will be the
// JSON and the second nil, otherwise the first one will be nil and the second
// one will be an error message.
func createJsonCommand(action string, params map[string]string) (string, error) {
    var command CommandRequest
    cmd, success := CommandMap[action]
    
    if success {
        command.SetValues(cmd.KodiName, params)    
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

