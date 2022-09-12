A rewrite of https://github.com/dradtke/vim-dap, exclusively in Go using Neovim's RPC API, though
potentially extensible to other editors.

## Plugin Configuration

Using Packer:

```lua
-- Debug Console
use {
	'dradtke/debug-console',
	rtp = 'misc/nvim',
	config = function(name, plugin)
		require('debug-console').setup(plugin.path, {
			-- Filetype DAP configurations
		})
	end,
}
```

- The Neovim-specific plugin files live in `misc/nvim`, so that needs to be added to the runtime
  path.
- Once the plugin is installed, `setup()` needs to be called with the plugin's installation path (so
  that it can build the necessary binary), and a per-filetype DAP configuration.

## DAP Configuration

The `run` object tells the plugin how to run the debug adapter. Many are simply subprocesses with
communication over standard streams, but other behaviors exist too.

The `launch` object tells the plugin how to send the `launch` request to the adapter once it's
running. This takes a function (TODO: multiple functions?) that should return the arguments to be
passed to the request.

## Running

```
:DebugRun <launch configuration> [args...]
```

As an example, to run a Go test file with verbose output (equivalent to running `go test -v`):

```
:DebugRun test -test.v
```

<!-- vim: set tw=100: -->
