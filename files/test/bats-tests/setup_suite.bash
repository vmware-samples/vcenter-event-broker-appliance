setup_suite() {
    # persist the shell_env file for later use by using a hardlink
    touch /root/config/shell_env
    ln /root/config/shell_env /root/test_env
}

teardown_suite() {
    #rm /root/test_env
    rm /root/ran_customization || true
    rm /root/.kube/config || true
}