#!/bin/bash

_goreg_completions() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    opts="-h --help -v --version -w --write -l --local -o --order -n --organization -m --minimize-group -a --sort-include-alias -r --remove-import-comment"

    # Suggest options
    if [[ ${cur} == -* ]]; then
        COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
        return 0
    fi

    # Complete filenames for .go files
    COMPREPLY=( $(compgen -f -X '!*.go' -- "$cur") )
}

complete -F _goreg_completions goreg
