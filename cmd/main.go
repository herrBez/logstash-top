package main

import (
	"fmt"
	"log"
	"time"

	logstash_flow "cmd/main.go/internal"

	"atomicgo.dev/cursor"
	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
)

var helpMessages = map[bool]string{
	false: `Help Message
- Up/Down to select the pipeline
- Enter to select the pipeline
- 'q' quit
- 'p' pause stop querying new data
`,
	true: `Help Message
- Up/Down to select the pipeline
- Enter to select the pipeline
- 'q' quit
- 'b' back
- 'p' pause stop querying new data
`,
}

var pausedMessage = map[bool]string{
	true:  "Active Query is stopped. Press 'p' to restart\n",
	false: "\n",
}

func main() {
	lfc := logstash_flow.NewLogstashFlowConfig("http://localhost:9600")
	pipelineInfo, err := lfc.GetPipelineInfo()

	for {
		if err == nil {
			break
		} else {
			log.Printf("%s", err)
			time.Sleep(2 * time.Second)
			pipelineInfo, err = lfc.GetPipelineInfo()
		}
	}

	pipelines := []string{}
	selectedIndex := 0

	for k, _ := range pipelineInfo.Pipelines {
		pipelines = append(pipelines, k)
	}

	// Whether to quit the program or not
	quit := false
	// Whether a pipeline is selected or not
	selected := false
	// Whether we need to display the help message or the data
	help := false
	// Whether we perform curl request or we stop
	paused := false

	go keyboard.Listen(func(key keys.Key) (stop bool, err error) {
		switch key.Code {
		case keys.CtrlC:
			quit = true
			return true, nil // Stop listener
		case keys.Down:
			// if !selected {
			selectedIndex = (selectedIndex + 1) % len(pipelines)
			// }
		case keys.Up:
			// if !selected {
			selectedIndex = (selectedIndex - 1 + len(pipelines)) % len(pipelines)
			// }

		case keys.Enter:
			selected = true

		case keys.RuneKey: // Check if key is a rune key (a, b, c, 1, 2, 3, ...)
			if key.String() == "q" { // Check if key is "q"
				quit = true
				return true, nil // Stop listener
			}
			if key.String() == "b" {
				if selected {
					selected = false
				}
			}
			if key.String() == "p" {
				paused = !paused
			}
			if key.String() == "h" {
				help = !help
			}
			if key.String() == "w" {
				lfc.NextMetric()
			}

			// fmt.Printf("\rYou pressed the rune key: %s\n", key)
			// default:
			// 	fmt.Printf("\rYou pressed: %s\n", key)
		}

		return false, nil // Return false to continue listening
	})

	fmt.Println("Logstash Top")
	fmt.Println("--------------------------")
	area := cursor.NewArea()
	previousSuccessFlowStatsAnswer := logstash_flow.PipelineAnswer{}
	pipelineFlowStatsAnswer := logstash_flow.PipelineAnswer{}
	lastMessage := ""
	message := ""
	sleepDuration := 1 * time.Second

	for !quit {
		lastMessage = message
		if help {
			message = helpMessages[selected]
			sleepDuration = 200 * time.Millisecond
		} else {
			sleepDuration = 1 * time.Second
			// If not paused download new information
			if !paused {
				previousSuccessFlowStatsAnswer = pipelineFlowStatsAnswer
				pipelineFlowStatsAnswer, err = lfc.GetPipelineFlowStats(pipelineInfo)
			}

			useAnswer := pipelineFlowStatsAnswer
			if err != nil {
				useAnswer = previousSuccessFlowStatsAnswer
			}
			content := ""
			if selected {
				content = lfc.RenderPipelineFlowStats(useAnswer, pipelineInfo, pipelines, selectedIndex)
			} else {
				content = lfc.RenderPipelineFlowStatsDetails(pipelineFlowStatsAnswer, pipelineInfo, pipelines, selectedIndex)
			}
			errorString := "\n"
			if err != nil {
				errorString = fmt.Sprintf("Warning: could not fetch data %s\n", err)
			}
			message = errorString + content + pausedMessage[paused]
		}
		if lastMessage != message {
			area.Update(message)

		}
		time.Sleep(sleepDuration)
	}

	area.Update("Quitting application\n\n")

}
