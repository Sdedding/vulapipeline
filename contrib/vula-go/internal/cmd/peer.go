package cmd

import "github.com/alecthomas/kong"

type PeerCmd struct {
	Show   PeerShowCmd   `cmd:"" help:"Show peer information"`
	Import PeerImportCmd `cmd:"" help:"Import peer descriptors"`
	Addr   PeerAddrCmd   `cmd:"" help:"Add and remove peer addresses"`
	Set    PeerSetCmd    `cmd:"" help:"Set arbitrary peer properties"`
	Remove PeerRemoveCmd `cmd:"" help:"Remove a peer"`
}

func (c *PeerCmd) Help() string {
	return `Commands to view and modify peer information
	
When "peer" is the top-level command, its subcommands communicate with
organize via dbus. This is the normal way to use these commands (eg, you can
run "vula peer show").

At this point in development, "peer" may also be run as a subcommand of the
"organize" command, in which case it will instantiate and operate on the
organize object directly. In the case of the "set" and "remove" peer
subcommands this only makes sense when there isn't an organize daemon
process running (as the daemon process will overwrite the state file written
by the "organize peer ..." command process without ever reading its
contents). Note that we do not currently check to ensure that there is not
an organize daemon running when "organize peer ..." commands are run.`
}

func (c *PeerCmd) Run(ctx *kong.Context) error {
	// TODO: Implement peer command
	panic("`peer` command not yet implemented")
	return nil
}

type PeerShowCmd struct {
	Descriptor bool `short:"D" help:"Print peer descriptor(s) instead of status"`
	All        bool `short:"a" help:"Show both enabled and disabled peers"`
	Disabled   bool `short:"d" help:"Show only disabled peers"`
	Enabled    bool `short:"e" help:"Show only enabled peers"`

	Peers []string `arg:"" optional:"" name:"peers" help:"Peers to show"`
}

func (c *PeerShowCmd) Help() string {
	return `Show peer information

With no arguments, all enabled peers are shown.

Peer arguments can be specified as ID, name, or IP, unless options are also
specified, in which case arguments must (currently) be IDs.`
}

func (c *PeerShowCmd) Run() error {
	// TODO: Implement peer show command
	panic("`peer show` command not yet implemented")
	return nil
}

type PeerImportCmd struct {
	File string `arg:"" optional:"" name:"file" help:"File containing peer descriptors"`
}

func (c *PeerImportCmd) Help() string {
	return `Import peer descriptors

Reads from standard input if a file is not specified.

Prints the result of processing each descriptor.

Can consume the output of "vula peer show --descriptor" from another system,
or the output of "vula discover --no-dbus --interface eth0".`
}

type PeerAddrCmd struct {
	Add PeerAddrAddCmd `cmd:"" help:"Add an address to a peer"`
	Rm  PeerAddrRmCmd  `cmd:"" help:"Delete an address from a peer"`
}

func (c *PeerAddrCmd) Help() string {
	return `Modify peer addresses

In the future, this might also show addresses, but for now that can be done
with the "peer show" command.`
}

type PeerAddrAddCmd struct {
	Vk string `arg:"" name:"vk" help:"Peer vula key"`
	Ip string `arg:"" name:"ip" help:"Peer IP"`
}

func (c *PeerAddrAddCmd) Help() string {
	return `Add an address to a peer`
}

type PeerAddrRmCmd struct {
	// TODO: probably routed to another command!
}

func (c *PeerAddrRmCmd) Help() string {
	return `Read 'rm' property of addr object`
}

type PeerSetCmd struct {
	Vk string `arg:"" name:"vk" help:"Peer vula key"`
	// TODO: this could be confusing for users switching
	// Note: unlike in the original vula, here the value must come before the paths because:
	// `required "value" cannot come after optional "path"`
	Value string   `arg:"" name:"value" help:"Peer value"`
	Path  []string `arg:"" optional:"" name:"path" help:"Peer path"`
}

func (c *PeerSetCmd) Help() string {
	// TODO: This text mentions `peer addr del` but there is only `peer addr rm`
	return `Modify arbitrary peer properties

This is currently the only way to verify peers, enable/disable them, and
enable or disable IP addresses.

In the future, this command should perhaps only be available for debugging,
and the normal user tasks which it currently performs should be handled by
other commands.

This command is *not* able to remove keys from dictionaries; for removing IP
addresses (instead of disabling them with this command) you can use the
"vula peer addr del" command. This command is able to add new IPs, but that
is better done with "vula peer addr add".`
}

type PeerRemoveCmd struct {
	Vk string `arg:"" name:"vk" help:"Peer vula key"`
}

func (c *PeerRemoveCmd) Help() string {
	return `Remove a peer`
}
