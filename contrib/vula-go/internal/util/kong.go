package util

import (
	"fmt"
	"strings"

	"github.com/alecthomas/kong"
)

// VulaHelpPrinter prints help texts like the [python version of vula]
// which is based on [click].
//
// It is fairly hacky but is here just in case one wants a "true" drop-in replacement
// down to the formatting of the help text.
//
// [click]: https://palletsprojects.com/projects/click/.
// [python version of vula]: https://codeberg.org/vula/vula
func VulaHelpPrinter(_options kong.HelpOptions, ctx *kong.Context) error {
	app := ctx.Model
	commandString := ctx.Command()
	selected := ctx.Selected()

	head := generateHead(ctx)
	fmt.Println(head)

	description := generateDescription(app, selected)
	if description != "" {
		fmt.Println(description)
	}

	options := ""
	if selected != nil && selected.Flags != nil {
		options = generateOptions(selected.Flags)
	} else if app != nil && commandString == "" && app.Flags != nil {
		options = generateOptions(app.Flags)
	} else {
		options = generateOptions(nil)
	}
	// Note: there is always a "--help" flag/option
	fmt.Println(options)

	commands := ""
	if selected != nil && selected.Children != nil {
		commands = generateCommands(selected.Children)
	} else if app != nil && commandString == "" && app.Children != nil {
		commands = generateCommands(app.Children)
	}
	if commands != "" {
		fmt.Println(commands)
	}

	return nil
}

func generateHead(ctx *kong.Context) string {
	app := ctx.Model
	appName := app.Name
	commandString := ctx.Command()
	selected := ctx.Selected()

	usageBase := fmt.Sprintf("Usage: %s", appName)

	// note: "[OPTIONS]" is always there as there is always a "--help" flag/option
	if commandString != "" {
		usageBase = fmt.Sprintf("%s %s [OPTIONS]", usageBase, commandString)
	} else {
		usageBase = fmt.Sprintf("%s [OPTIONS]", usageBase)
	}

	if selected != nil && selected.Positional != nil {
		for _, child := range selected.Positional {
			name := strings.ToUpper(child.Name)
			if child.Required {
				usageBase = fmt.Sprintf("%s %s", usageBase, name)
			} else {
				usageBase = fmt.Sprintf("%s [%s]", usageBase, name)
			}
			// this is truly hacky and perhaps incompatible with generics?
			if strings.HasPrefix(child.Target.String(), "<[]") {
				usageBase = fmt.Sprintf("%s...", usageBase)
			}
		}
	} else if (app.Node != nil && app.Node.Children != nil) || (selected != nil && selected.Children != nil) {
		usageBase = fmt.Sprintf("%s COMMAND [ARGS]...", usageBase)
	}

	return usageBase
}

func generateDescription(app *kong.Application, selected *kong.Node) string {
	description := ""
	if selected != nil {
		description = selected.Detail
	} else if app != nil {
		description = app.Help
	}

	if description != "" {
		indentedDescription := "\n"
		lines := strings.Split(description, "\n")
		for i, line := range lines {
			indentedDescription += "  " + line
			if i < len(lines)-1 {
				indentedDescription += "\n"
			}
		}
		return indentedDescription
	}

	return description
}

func generateOptions(flags []*kong.Flag) string {
	var optionsText string
	var maxLength int

	for _, flag := range flags {
		var names string
		if flag.Short != 0 {
			if flag.Value.Tag.Negatable != "" {
				names = fmt.Sprintf("-%c, --%s / --no-%s", flag.Short, flag.Name, flag.Name)
			} else {
				names = fmt.Sprintf("-%c, --%s", flag.Short, flag.Name)
			}
		} else {
			if flag.Value.Tag.Negatable != "" {
				names = fmt.Sprintf("--%s / --no-%s", flag.Name, flag.Name)
			} else {
				names = fmt.Sprintf("--%s", flag.Name)
			}
		}
		if len(names) > maxLength {
			maxLength = len(names)
		}
	}

	optionsText = "\nOptions:"

	for _, flag := range flags {
		if flag.Name != "help" {
			var names string
			if flag.Short != 0 {
				if flag.Value.Tag.Negatable != "" {
					names = fmt.Sprintf("-%c, --%s / --no-%s", flag.Short, flag.Name, flag.Name)
				} else {
					names = fmt.Sprintf("-%c, --%s", flag.Short, flag.Name)
				}
			} else {
				if flag.Value.Tag.Negatable != "" {
					names = fmt.Sprintf("--%s / --no-%s", flag.Name, flag.Name)
				} else {
					names = fmt.Sprintf("--%s", flag.Name)
				}
			}
			optionsText = fmt.Sprintf("%s\n  %-*s  %s", optionsText, maxLength, names, flag.Help)
		}
	}

	optionsText = fmt.Sprintf("%s\n  %-*s  %s", optionsText, maxLength, "--help", "Show this message and exit.")

	return optionsText
}

func generateCommands(nodes []*kong.Node) string {
	if len(nodes) == 0 {
		return ""
	}

	var commandsText string
	var maxLength int

	for _, command := range nodes {
		if len(command.Name) > maxLength {
			maxLength = len(command.Name)
		}
	}

	if len(nodes) > 0 {
		commandsText = "\nCommands:"
	}

	for _, command := range nodes {
		commandsText = fmt.Sprintf("%s\n  %-*s  %s", commandsText, maxLength, command.Name, command.Help)
	}

	return commandsText
}
