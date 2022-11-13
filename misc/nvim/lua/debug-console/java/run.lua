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
	vim.lsp.buf_request(0, 'workspace/executeCommand', {
		command = 'java.project.getClasspaths',
		arguments = { 'file://'..filepath, vim.fn.json_encode({ scope = 'test' }) },
	}, function(err, result, ctx, config)
		if err then error(err) end
		vim.fn.DebugConsoleLaunch({
			mainClass = 'Hello', -- TODO: this should be better
			args = '',
			classPaths = result['classpaths'],
			modulePaths = result['modulepaths'],
			cwd = result['projectRoot'],
			projectName = vim.fn.fnamemodify(result['projectRoot'], ':t'),
			shortenCommandLine = 'jarmanifest',
		})
	end)
end

return M
