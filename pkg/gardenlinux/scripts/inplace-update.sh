#!/bin/bash

set -Eeuo pipefail

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <version>"
    exit 1
fi

VERSION=$1
flags=()

if [[ -f /etc/gardenlinux/usirepo.conf ]]; then
    REPO=$(< /etc/gardenlinux/usirepo.conf)
    flags=(--repo "$REPO")
fi

if gardenlinux-update "${flags[@]}" "$VERSION"; then
    echo "exit status 0: success"
    reboot
else
    EXIT_CODE=$?

    case "$EXIT_CODE" in
        1)
            echo "exit status 1: invalid arguments"
            ;;
        2)
            echo "exit status 2: system failure"
            ;;
        3)
            echo "exit status 3: network problems"
            ;;
        *)
            echo "exit status $EXIT_CODE: unknown error"
            ;;
    esac

    exit "$EXIT_CODE"
fi
