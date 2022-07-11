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
	return jobstart([l:bin, 'nvim'], {
				\ 'rpc': v:true,
				\ 'env': {
					\ 'LOG_FILE': stdpath('cache').'/debug-console.log'
					\ },
				\ })
endfunction

call remote#host#Register('debug-console', 'x', function('s:Start'))

" The end of this file will be updated when `make` is run with a new manifest.

call remote#host#RegisterPlugin('debug-console', '0', [
\ {'type': 'command', 'name': 'DebugRun', 'sync': 1, 'opts': {}},
\ ])
