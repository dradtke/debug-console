local M = {}

M.run = function()
	vim.lsp.buf_request(0, 'workspace/executeCommand', {
		command = 'vscode.java.startDebugSession',
	}, function(err, result, ctx, config)
		if err then error(err) end
		vim.fn.DebugConsoleRun({
			type = 'remote',
			address = 'localhost:'..result,
		})
	end)
end

M.launch = function(filepath, args)
	-- TODO: support passing in additional arguments beyond the main class name
	vim.lsp.buf_request(0, 'workspace/executeCommand', {
		command = 'java.project.getClasspaths',
		arguments = { 'file://'..filepath, vim.fn.json_encode({ scope = 'test' }) },
	}, function(err, result, ctx, config)
		if err then error(err) end
		-- See: https://github.com/microsoft/java-debug/blob/5bac075002154f2c8441a39026db53bdc67aef56/com.microsoft.java.debug.core/src/main/java/com/microsoft/java/debug/core/protocol/Requests.java#L104
		vim.fn.DebugConsoleLaunch({
			mainClass = args[1],
			args = '',
			classPaths = result['classpaths'],
			modulePaths = result['modulepaths'],
			cwd = result['projectRoot'],
			projectName = vim.fn.fnamemodify(result['projectRoot'], ':t'),
			shortenCommandLine = 'jarmanifest',
			console = 'integratedTerminal',
		})
	end)
end

return M
