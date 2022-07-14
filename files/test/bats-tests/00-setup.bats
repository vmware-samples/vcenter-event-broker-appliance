setup() {
   load 'common-setup'
   _common_setup
}

@test "can run setup script" {
   # run the setup scripts
   bats_require_minimum_version 1.5.0
   run -0 setup.sh
}

validate_yamls() {
   while read yaml; do
      kubeval --ignore-missing-schemas --quiet -v 1.21.5 -s file:///root/kubernetes-json-schema/ "$yaml" | grep -vE -e '^(PASS)' -e 'given: null' -e 'not validated against a schema'
   done </root/kubeval-yamls.txt
   rm /root/kubeval-yamls.txt
}

@test "validate kubernetes yaml manifests" {
   run validate_yamls
   refute_output
}