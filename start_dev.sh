#!/bin/bash

./tailwindcss -i static/index.css -o static/output.css --watch --minify &> /dev/stdout &

air

# Kill background processes when this script is stopped
trap "exit" INT TERM ERR
trap "kill 0" EXIT
