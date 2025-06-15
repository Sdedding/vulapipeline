package cmd

type Globals struct {
	Info    bool `short:"i" help:"Set log level INFO"`
	Quiet   bool `short:"q" help:"Set log level WARN"`
	Verbose bool `short:"v" help:"Set log level DEBUG"`
}
