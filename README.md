# explain

## the main thing

Explain things with openai models in a really convenient manner for people that just live in the terminal and don't like context switching


## how to install

Right now it needs the go toolchain for installing.

`$ go install github.com/alarbada/explain`


## how to use

`explain` accepts the following command-line flags:

- `-clear`: Clears the conversation history. Use this flag without any arguments.

- `-model [model_name]`: Changes the model used for the conversation to the one specified by `[model_name]`. If missed a list of models will be printed

- `-init`: Initializes the configuration file.

- `-config`: Shows the current configuration.

- `-conversation`: Displays the current conversation. Use this flag without any arguments.

- `-help`: Show this help


# TODO 

- [ ] Maybe make this more fancy with these:
  - https://github.com/charmbracelet/bubbletea 
  - https://github.com/charmbracelet/lipgloss
- [ ] After the previous, allow for selecting multiple chats. Maybe the goal here is to replace my previous tool, [sira](https://github.com/alarbada/sira)
- [ ] Add a pricing page with a monthly and yearly subscription, of course
