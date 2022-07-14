setup() {
   load 'common-setup'
   _common_setup
}

@test "validate root password" {
   salt=$(getent shadow root | cut -d$ -f3)
   epassword=$(getent shadow root | cut -d: -f2)
   match=$(echo ${ROOT_PASSWORD} | openssl passwd -6 -salt ${salt} -stdin)
   assert [ ${match} == ${epassword} ]
}

@test "validate iptables" {
   assert_file_contains /etc/systemd/scripts/ip4save '^-A INPUT -i gw0 -j ACCEPT$'
}