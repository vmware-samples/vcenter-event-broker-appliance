_common_setup() {
    load '/root/test_helper/bats-support/load'
    load '/root/test_helper/bats-assert/load'
    load '/root/test_helper/bats-file/load'
    # get the containing directory of this file
    # use $BATS_TEST_FILENAME instead of ${BASH_SOURCE[0]} or $0,
    # as those will point to the bats executable's location or the preprocessed file respectively
    DIR="$( cd "$( dirname "$BATS_TEST_FILENAME" )" >/dev/null 2>&1 && pwd )"
    # make scripts in setup/ visible to PATH
    PATH="$DIR/../..:$PATH"
    source /root/test_env
}

_get_val_from_file() {
    file=$1
    val=$2

    echo $(sed -nr "s/^${val}=(['\"]?)(.*)(\1)/\2/p" < ${file})
}