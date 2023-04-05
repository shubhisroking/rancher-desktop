# Test case 30

setup() {
    load '../helpers/load'
    PATH=$(ruby -e 'print ENV["PATH"].split(":").reject{|x| x[".rd/bin"]}.join(":")')
    export PATH="$PATH"
}

@test 'factory reset' {
    factory_reset
}

@test 'start rancher desktop' {
    run rdctl start --application.path-management-strategy=rcfiles --kubernetes.enabled=false
    assert_success
}

@test 'ensure the app is running' {
    try --max 5 --delay 5 rdctl shell echo abc
    assert_success
}

@test 'ensure path-management is managed' {
    # Looks like starting doesn't necessarily guarantee the files have the managed path spec
    run rdctl set --application.path-management-strategy=manual
    assert_success
    sleep 5
    run rdctl set --application.path-management-strategy=rcfiles
    assert_success
    sleep 5
}

# Running `bash -l -c` causes bats to hang
#@test 'bash managed' {
#    run bash -l -c "which rdctl"
#    assert_output --partial '.rd/bin/rdctl'
#}

@test 'ksh managed' {
    if [ -f "$HOME/.kshrc" ] ; then
        run ksh -c "which rdctl"
        assert_output --partial '.rd/bin/rdctl'
    fi
}

@test 'zsh managed' {
    if [ -f "$HOME/.zshrc" ] ; then
        run zsh -i -c "which rdctl"
        assert_output --partial '.rd/bin/rdctl'
    fi
}

@test 'fish managed' {
    run fish -c "which rdctl"
    assert_output --partial '.rd/bin/rdctl'
}

@test 'move to manual path-management' {
    rdctl set --application.path-management-strategy=manual
    sleep 5
}

#@test 'bash unmanaged' {
#    run bash -l -c "which -s rdctl"
#    assert_failure
#}

@test 'ksh unmanaged' {
    if [ -f "$HOME/.kshrc" ] ; then
        run ksh -c "which -s rdctl"
        assert_failure
    fi
}

@test 'zsh unmanaged' {
    if [ -f "$HOME/.zshrc" ] ; then
        run zsh -i -c "which -s rdctl"
        assert_failure
    fi
}

@test 'fish unmanaged' {
    run fish -c "which -s rdctl"
    assert_failure
}

#@test 'say goodnight dick' {
#    ps auxww | grep bats.*verify-paths.bats | awk '{ print $2 }' | xargs kill
#}
