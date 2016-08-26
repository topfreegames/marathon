#!/bin/bash
set -e

ls -la /
/bin/mount -t tmpfs -o size=256M tmpfs /pg_data
/bin/mount
