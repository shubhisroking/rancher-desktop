setup() {
    load '../helpers/load'
}

@test 'Verify that initial Behavior is all set to false' {
    factory_reset
    start_kubernetes
    wait_for_apiserver
    run bash -c "rdctl list-settings | jq -r '.application.autoStart'"
    assert_output false
    run bash -c "rdctl list-settings | jq -r '.applicaion.startInBackground'"
    assert_output false
    run bash -c "rdctl list-settings | jq -r '.applicaion.window.quitOnClose'"
    assert_output false
    run bash -c "rdctl list-settings | jq -r '.applicaion.hideNotificationIcon'"
    assert_output false
}

@test 'Enable auto start' {
    run rdctl set --application.auto-start=true
    assert_success
    run bash -c "rdctl list-settings | jq -r '.application.autoStart'"
    assert_output true
}

@test 'Verify that the auto-start config is created' {

    if is_linux; then
        run bash -c "cat $XDG_CONFIG_HOME/autostart/rancher-desktop.desktop &>/dev/null"
        assert_success
    fi

    if is_macos; then
        run bash -c "cat $HOME/Library/LaunchAgents/io.rancherdesktop.autostart.plist &>/dev/null"
        assert_success
    fi

    if is_windows; then
        run powershell.exe -c "reg query HKCU\Software\Microsoft\Windows\CurrentVersion\Run /v RancherDesktop"
        assert_output --partial "Program Files\Rancher Desktop\Rancher Desktop.exe"
    fi
}

@test 'Disable auto start' {
    run rdctl set --application.auto-start=false
    assert_success
    run bash -c "rdctl list-settings | jq -r '.application.autoStart'"
    assert_output false
}

@test 'Verify that the auto-start config is removed' {

    if is_linux; then
        run bash -c "cat $XDG_CONFIG_HOME/autostart/rancher-desktop.desktop &>/dev/null"
        assert_failure
    fi

    if is_macos; then
        run bash -c "cat $HOME/Library/LaunchAgents/io.rancherdesktop.autostart.plist &>/dev/null"
        assert_failure
    fi

    if is_windows; then
        run powershell.exe -c "reg query HKCU\Software\Microsoft\Windows\CurrentVersion\Run /v RancherDesktop"
        assert_output --partial "The system was unable to find the specified registry"
    fi
}

@test 'Enable quit-on-close' {
    run rdctl set --application.window.quit-on-close=true
    assert_success
    run bash -c "rdctl list-settings | jq -r '.application.window.quitOnClose'"
    assert_output true
}

@test 'Disable quit-on-close' {
    run rdctl set --application.window.quit-on-close=false
    assert_success
    run bash -c "rdctl list-settings | jq -r '.application.window.quitOnClose'"
    assert_output false
}

@test 'Enable start-in-background' {
    run rdctl set --application.start-in-background=true
    assert_success
    run bash -c "rdctl list-settings | jq -r '.application.startInBackground'"
    assert_output true
}

@test 'Disable start-in-background' {
    run rdctl set --application.start-in-background=false
    assert_success
    run bash -c "rdctl list-settings | jq -r '.application.startInBackground'"
    assert_output false
}

@test 'Enable hide-notification-icon' {
    run rdctl set --application.hide-notification-icon=true
    assert_success
    run bash -c "rdctl list-settings | jq -r '.application.hideNotificationIcon'"
    assert_output true
}

@test 'Disable hide-notification-icon' {
    run rdctl set --application.hide-notification-icon=false
    assert_success
    run bash -c "rdctl list-settings | jq -r '.application.hideNotificationIcon'"
    assert_output false
}
