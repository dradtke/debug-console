local M = {}

M.setup = function(plugin_path)
	local repo_root = vim.fn.fnamemodify(plugin_path, ':h:h')
	vim.fn.mkdir(repo_root..'/bin', 'p')
	--local build_command = 'cd '..vim.fn.shellescape(repo_root)..' && go build -o ./bin ./cmd/debug-console'

	local build_command = {'go', 'build', '-o', './bin', './cmd/debug-console'}
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
			if exit_code ~= 0 then
				print('debug-console: build exited with code: '..exit_code)
			end
		end,
	})
end

return M
