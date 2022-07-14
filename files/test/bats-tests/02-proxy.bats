setup() {
   load 'common-setup'
   _common_setup
}

@test "proxy file exists" {
   assert [ -f /etc/sysconfig/proxy ]
}

@test "proxy file NO_PROXY" {
    # v=$(_get_val_from_file /etc/sysconfig/proxy NO_PROXY)
    # assert [ "$v" == "$NO_PROXY" ]
    if [ -z ${NO_PROXY} ]; then
        skip "NO_PROXY not set"
    fi
    assert_file_contains /etc/sysconfig/proxy "^NO_PROXY=\"${NO_PROXY}\"$"
    assert_file_contains /usr/lib/systemd/system/containerd.service "^Environment=NO_PROXY=${NO_PROXY}$"
}

@test "proxy file HTTP_PROXY" {
    if [ -z ${HTTP_PROXY} ]; then
        skip "HTTP_PROXY not set"
    fi
    assert_file_contains /etc/sysconfig/proxy "^HTTP_PROXY=\"${HTTP_PROXY}\"$"
    assert_file_contains /usr/lib/systemd/system/containerd.service "^Environment=HTTP_PROXY=${HTTP_PROXY}$"
}