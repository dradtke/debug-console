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
		local mason_dir = '/home/damien/.local/share/nvim/mason'  -- TODO: get mason dir
		require('debug-console').setup(plugin.path, {
			go = {
				run = {
					type = 'subprocess',
					command = {mason_dir..'/bin/go-debug-adapter'},
				},
				-- TODO: support multiple launch configurations
				launch = function(filepath)
					local launch_args = {
						request = 'launch',
						program = filepath,
						dlvToolPath = '/home/damien/.asdf/installs/golang/1.18.3/packages/bin/dlv',
						args = {}
					}
					local test_suffix = '_test.go'
					if filepath:sub(-#test_suffix) == test_suffix then
						launch_args['mode'] = 'test'
					else
						launch_args['mode'] = 'debug'
					end
					return launch_args
				end,
			},
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

<!-- vim: set tw=100: -->
