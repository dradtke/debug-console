local M = {}

function init(config)
	vim.cmd 'runtime debug-console.vim'
	-- Register DAP configurations
	for _,v in pairs(config) do
		v.launch = string.dump(v.launch)
	end
	vim.fn.DebugConsoleSetConfig(config)
end

M.setup = function(plugin_path, config)
	local repo_root = vim.fn.fnamemodify(plugin_path, ':h:h')
	vim.fn.mkdir(repo_root..'/bin', 'p')
	--local build_command = 'cd '..vim.fn.shellescape(repo_root)..' && go build -o ./bin ./cmd/debug-console'

	local build_command = {'go', 'build', '-tags', 'nvim', '-o', './bin', './cmd/debug-console'}
	vim.fn.jobstart(build_command, {
		cwd = repo_root,
		on_stderr = function(job_id, data, event_type)
			for _,line in ipairs(data) do
				if line ~= '' then
					print('debug-console: '..line)
				end
			end
		end,
		on_exit = function(job_id, exit_code, event_type)
			if exit_code == 0 then
				init(config)
			else
				print('debug-console: build exited with code: '..exit_code)
			end
		end,
	})
end

return M
