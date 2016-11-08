#!/usr/bin/env bash

# Redirect output to stderr.
exec 1>&2
# enable user input
exec < /dev/tty

forbiddenregexp='^\+.*(describe\.only|it\.only|console\.log|[.]then\(|[.]catch\()\s*[(]'
# CHECK
if test $(git diff --cached | egrep $forbiddenregexp | wc -l) != 0
then
    echo "Proposed diff:"
    exec git diff --cached | egrep -ne $forbiddenregexp
    echo
    echo "In the above diff, there's at least one occurrence of:"
    echo "    * it.only;"
    echo "    * describe.only;"
    echo "    * .then( (a promise return);"
    echo "    * .catch( (a promise catch-all);"
    echo "    * console.log."
    echo
    echo "Please remove it before continuing!"
    exit 1;
fi
