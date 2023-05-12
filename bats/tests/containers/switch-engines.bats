# Test case 20

load '../helpers/load'
RD_CONTAINER_ENGINE=moby

switch_container_engine() {
    local name=$1
    RD_CONTAINER_ENGINE="${name}"
    rdctl set --container-engine.name="${name}"
    wait_for_container_engine
}

pull_containers() {
    ctrctl run -d -p 8085:80 --restart=always nginx 1>&3-
    ctrctl run -d --restart=no busybox /bin/sh -c "sleep inf" 1>&3-
    run ctrctl ps --format '{{json .Image}}'
    assert_success
    assert_output --partial nginx
    assert_output --partial busybox
}

@test 'factory reset' {
    factory_reset
}

@test 'start application' {
    start_container_engine
    wait_for_container_engine
}

@test 'start moby and pull nginx' {
    pull_containers
}

run_curl() {
    if is_unix; then
        run curl localhost:8085
        assert_success
        assert_output --partial "Welcome to nginx"
    elif is_windows; then
        run powershell.exe -c "wsl -d rancher-desktop curl localhost:8085"
        assert_success
        assert_output --partial "Welcome to nginx"
    fi
}

@test 'verify that container UI is accessible moby' {
    try run_curl
    assert_success
}

@test "switch to containerd" {
    switch_container_engine containerd
    pull_containers
}

@test 'verify that container UI is accessible containerd' {
    try run_curl
    assert_success
}

verify_post_switch_containers() {
    run ctrctl ps --format '{{json .Image}}'
    assert_output --partial "nginx"
    refute_output --partial "busybox"
}

switch_back_verify_post_switch_containers() {
    local name=$1
    switch_container_engine "${name}"
    try --max 12 --delay 5 verify_post_switch_containers
    assert_success
}

@test 'switch back to moby and verify containers' {
    switch_back_verify_post_switch_containers moby
}

@test 'verify that container UI is accessible switching back to moby' {
    try run_curl
    assert_success
}

@test 'switch back to containerd and verify containers' {
    switch_back_verify_post_switch_containers containerd
}

@test 'verify that container UI is accessible switching back to containerd' {
    try run_curl
    assert_success
}
