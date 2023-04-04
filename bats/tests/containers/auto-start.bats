setup() {
    load '../helpers/load'
}

@test 'Verify that initial Behavior is all set to false' {
    factory_reset
    start_kubernetes
    wait_for_apiserver
    run bash -c "rdctl list-settings | jq -r '.application.autoStart'"
    assert_output false
    run  bash -c "rdctl list-settings | jq -r '.applicaion.startInBackground'"
    assert_output false
    run bash -c "rdctl list-settings | jq -r '.applicaion.window.quitOnClose'"
    assert_output false
    run bash -c "rdctl list-settings | jq -r '.applicaion.hideNotificationIcon'"
    assert_output false
}

@test 'Verify auto start' {
    run rdctl set --application.auto-start=true
    assert_success
    run bash -c "rdctl list-settings | jq -r '.application.autoStart'"
    assert_output true

    #verify that the auto-start config is created

    if is_linux; then
        run bash -c "cat $XDG_CONFIG_HOME/autostart/rancher-desktop.desktop"
        assert_success
    fi

    if is_windows; then
        run powershell.exe -c "reg query HKCU\Software\Microsoft\Windows\CurrentVersion\Run /v RancherDesktop"
        assert_output --partial "Program Files\Rancher Desktop\Rancher Desktop.exe"
    fi

    if is_macos; then
        run bash -c "cat ~/Library/LaunchAgents/io.rancherdesktop.autostart.plist"
        assert_success
    fi

}

@test 'Verify quit on close' {


}
@test 'Verify start in the backgorund' {


}

@test 'Verify hide notification icon' {


}
