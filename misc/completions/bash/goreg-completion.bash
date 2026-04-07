#!/bin/bash

_goreg_completions() {
    local cur prev opts subcommands
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    opts="-h --help -v --version -w --write -l --local -o --order -n --organization -m --minimize-group -a --sort-include-alias -r --remove-import-comment"
    subcommands="init"

    # If we're at the first argument position, suggest subcommands and options
    if [[ ${COMP_CWORD} -eq 1 ]]; then
        if [[ ${cur} == -* ]]; then
            COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
        else
            COMPREPLY=( $(compgen -W "${subcommands}" -- ${cur}) $(compgen -f -X '!*.go' -- "$cur") )
        fi
        return 0
    fi

    # Suggest options
    if [[ ${cur} == -* ]]; then
        COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
        return 0
    fi

    # Complete filenames for .go files
    COMPREPLY=( $(compgen -f -X '!*.go' -- "$cur") )
}

complete -F _goreg_completions goreg
