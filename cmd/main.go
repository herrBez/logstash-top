package main

import (
	"fmt"
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
	true:  "Active Query is stopped\n",
	false: "\n",
}

func main() {
	lfc := logstash_flow.NewLogstashFlowConfig("http://localhost:9600")
	pipelineInfo := lfc.GetPipelineInfo()

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

			// fmt.Printf("\rYou pressed the rune key: %s\n", key)
			// default:
			// 	fmt.Printf("\rYou pressed: %s\n", key)
		}

		return false, nil // Return false to continue listening
	})

	fmt.Println("Logstash Top")
	fmt.Println("--------------------------")
	area := cursor.NewArea()
	pipelineFlowStatsAnswer := lfc.GetPipelineFlowStats(pipelineInfo)

	for !quit {
		if help {
			area.Update(helpMessages[selected])
			time.Sleep(200 * time.Millisecond)
		} else {
			if !paused {
				pipelineFlowStatsAnswer = lfc.GetPipelineFlowStats(pipelineInfo)
			}
			if !selected {
				content := logstash_flow.RenderPipelineFlowStats(pipelineFlowStatsAnswer, pipelineInfo, pipelines, selectedIndex)

				area.Update(content + pausedMessage[paused])
				time.Sleep(1 * time.Second)

			} else {

				content := logstash_flow.RenderPipelineFlowStatsDetails(pipelineFlowStatsAnswer, pipelineInfo, pipelines, selectedIndex)

				area.Update(content + pausedMessage[paused])
				time.Sleep(1 * time.Second)
			}
		}
	}

	area.Update("Quitting application\n\n")

}
