# debug echos output to stderr rather than stdout
function debug () {
    echo "DEBUG: $@" > /dev/stderr
}

