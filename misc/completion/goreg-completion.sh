#!/bin/bash

# Check if running in Zsh or Bash
if [ -n "$BASH_VERSION" ]; then
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
elif [ -n "$ZSH_VERSION" ]; then
    # Zsh completion
    _goreg() {
        local arguments
        arguments=(
            '-h[Show help message and exit]'
            '--help[Show help message and exit]'
            '-v[Show version]'
            '--version[Show version]'
            '-w[Write formatted imports directly to the file]'
            '--write[Write formatted imports directly to the file]'
            '-l[Specify the local module path]:local module path:_files'
            '--local[Specify the local module path]:local module path:_files'
            '-o[Specify the order of import groups]:group order:(std thirdparty organization local)'
            '--order[Specify the order of import groups]:group order:(std thirdparty organization local)'
            '-n[Specify the module path of your organization]:organization path:_files'
            '--organization[Specify the module path of your organization]:organization path:_files'
            '-m[Do not separate import groups when an alias is present]'
            '--minimize-group[Do not separate import groups when an alias is present]'
            '-a[Sort imports with aliases within their respective groups]'
            '--sort-include-alias[Sort imports with aliases within their respective groups]'
            '-r[Remove the comments in the import]'
            '--remove-import-comment[Remove the comments in the import]'
            ':Go file:_files -g "*.go"'
        )
        _arguments -s $arguments
    }

    compdef _goreg goreg
fi

