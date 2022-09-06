if !has('nvim')
	echoerr 'debug-console: This plugin requires Neovim!'
endif

if exists('g:loaded_debug_console')
	finish
endif
let g:loaded_debug_console = 1

let s:dir = expand('<sfile>:h:h:h')

function! s:Start(host) abort
	let l:bin = s:dir.'/bin/debug-console'
	let l:cache = stdpath('cache')
	return jobstart([l:bin, 'nvim'], {
				\ 'rpc': v:true,
				\ 'env': {
					\ 'LOG_FILE': l:cache.'/debug-console.log',
					\ 'DAP_DIR': l:cache.'/dap',
					\ },
				\ })
endfunction

call remote#host#Register('debug-console', 'x', function('s:Start'))

sign define debug-console-breakpoint text=B

" The end of this file will be updated when `make` is run with a new manifest.

call remote#host#RegisterPlugin('debug-console', '0', [
\ {'type': 'autocmd', 'name': 'VimLeave', 'sync': 0, 'opts': {'pattern': '*'}},
\ {'type': 'command', 'name': 'CurrentLocation', 'sync': 1, 'opts': {}},
\ {'type': 'command', 'name': 'DebugRun', 'sync': 1, 'opts': {'eval': '{''Path'': expand(''%:p''), ''Filetype'': getbufvar(bufnr(''%''), ''&filetype'')}'}},
\ {'type': 'command', 'name': 'ToggleBreakpoint', 'sync': 1, 'opts': {}},
\ ])
