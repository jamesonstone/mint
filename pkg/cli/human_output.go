package cli

import (
	"fmt"
	"io"

	"golang.org/x/term"
)

const (
	whiteBold = "\033[1;37m"
	reset     = "\033[0m"
)

var terminalWriterCheck = isTerminalWriter

type humanOutputStyle struct {
	enabled bool
}

func styleForWriter(w io.Writer) humanOutputStyle {
	return humanOutputStyle{enabled: terminalWriterCheck(w)}
}

func isTerminalWriter(w io.Writer) bool {
	fileLike, ok := w.(interface{ Fd() uintptr })
	if !ok {
		return false
	}

	return term.IsTerminal(int(fileLike.Fd()))
}

func (s humanOutputStyle) title(emoji, text string) string {
	if !s.enabled {
		return text
	}

	return whiteBold + emoji + " " + text + reset
}

func (s humanOutputStyle) label(text string) string {
	if !s.enabled {
		return text
	}

	return whiteBold + text + reset
}

func helpTemplate(enabled bool) string {
	usageHeader := "Usage:"
	aliasesHeader := "Aliases:"
	examplesHeader := "Examples:"
	commandsHeader := "Available Commands:"
	flagsHeader := "Flags:"
	globalFlagsHeader := "Global Flags:"
	additionalHelpHeader := "Additional Help Topics:"
	moreInfoLabel := "Use"

	if enabled {
		usageHeader = "🚀 Usage"
		aliasesHeader = "🏷️ Aliases"
		examplesHeader = "🧪 Examples"
		commandsHeader = "🧰 Available Commands"
		flagsHeader = "⚙️ Flags"
		globalFlagsHeader = "🌐 Global Flags"
		additionalHelpHeader = "📚 Additional Help Topics"
		moreInfoLabel = "🔎 Use"
	}

	return fmt.Sprintf(`{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}{{end}}

%s
  {{if .Runnable}}{{.UseLine}}{{end}}{{if and .Runnable .HasAvailableSubCommands}}
  {{end}}{{if .HasAvailableSubCommands}}{{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

%s
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

%s
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

%s
{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}  {{rpad .Name .NamePadding }} {{.Short}}
{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

%s
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

%s
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

%s
{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}  {{rpad .CommandPath .CommandPathPadding }} {{.Short}}
{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

%s "{{.CommandPath}} [command] --help" for more information about a command.
{{end}}`,
		usageHeader,
		aliasesHeader,
		examplesHeader,
		commandsHeader,
		flagsHeader,
		globalFlagsHeader,
		additionalHelpHeader,
		moreInfoLabel,
	)
}

func usageTemplate(enabled bool) string {
	header := "Usage:"
	commandsHeader := "Available Commands:"
	flagsHeader := "Flags:"
	globalFlagsHeader := "Global Flags:"

	if enabled {
		header = "🚀 Usage"
		commandsHeader = "🧰 Available Commands"
		flagsHeader = "⚙️ Flags"
		globalFlagsHeader = "🌐 Global Flags"
	}

	return fmt.Sprintf(`%s
  {{.UseLine}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasAvailableSubCommands}}

%s
{{range .Commands}}{{if .IsAvailableCommand}}  {{rpad .Name .NamePadding }} {{.Short}}
{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

%s
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

%s
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}
`, header, commandsHeader, flagsHeader, globalFlagsHeader)
}
