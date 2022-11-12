local M = {}

M.run = function()
	local mason_dir = vim.fn.stdpath('data')..'/mason'
	vim.fn.DebugConsoleRun({
		type = 'subprocess',
		command = {mason_dir..'/bin/dlv', 'dap', '--client-addr', '${CLIENT_ADDR}'},
		dialClient = true,
	})
end

M.launch = function(filepath, args)
	-- See: https://pkg.go.dev/github.com/go-delve/delve/service/dap#LaunchConfig
	vim.fn.DebugConsoleLaunch({
		mode = 'test',
		program = filepath,
		args = args,
	})
end

return M
