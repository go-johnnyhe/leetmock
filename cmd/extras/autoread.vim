if exists('g:loaded_mycli_autoread') | finish | endif
let g:loaded_mycli_autoread = 1

set autoread
set updatetime=100

function! CheckForChanges(...)
    silent! checktime
endfunction

" Timer-based checking for Vim 8+
if has('timers')
    let s:autoread_timer = -1
    
    function! StartAutoreadTimer()
        if s:autoread_timer != -1
            call timer_stop(s:autoread_timer)
        endif
        let s:autoread_timer = timer_start(200, function('CheckForChanges'), {'repeat': -1})
    endfunction
    
    call StartAutoreadTimer()
endif

augroup mycli_autoread
    autocmd!
    autocmd FocusGained,BufEnter,CursorHold,CursorHoldI * call CheckForChanges()
    autocmd CursorMoved,CursorMovedI * call CheckForChanges()
    autocmd InsertEnter,InsertLeave * call CheckForChanges()
    autocmd TextChanged,TextChangedI * call CheckForChanges()
augroup END
