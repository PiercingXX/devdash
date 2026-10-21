#!/bin/bash
# Generate the screenshots of the documentation in doc/img.
# Needs tmux, python3 with Pillow, and podman or docker. The Google services use sample data from a fake API.
set -euo pipefail

dir=$(cd "$(dirname "$0")" && pwd)
root=$(cd "$dir/../.." && pwd)
work=$(mktemp -d)
trap 'kill "${fake:-}" 2>/dev/null || true; tmux -L devdash-shots kill-server 2>/dev/null || true; rm -rf "$work"' EXIT

engine=$(command -v podman || command -v docker)
socket=devdash-shots

(cd "$root" && CGO_ENABLED=0 go build -o "$work/devdash" .)
"$engine" build -q -t devdash-screenshots "$dir" > /dev/null
cp "$root"/example/*.yml "$work"

# shot <name> <columns> <rows> <seconds> <command>: capture the command in tmux, and render it in doc/img/<name>.png.
shot() {
    local name=$1 cols=$2 rows=$3 wait=$4
    shift 4
    tmux -L $socket kill-server 2> /dev/null || true
    tmux -L $socket new-session -d -s shot -x "$cols" -y "$rows" "TERM=tmux-256color $*; sleep 600"
    sleep "$wait"
    tmux -L $socket capture-pane -t shot -p -e > "$work/$name.ansi"
    if tmux -L $socket capture-pane -t shot -p | grep -q ERROR; then
        echo "error in the screenshot $name:" >&2
        tmux -L $socket capture-pane -t shot -p >&2
        exit 1
    fi
    tmux -L $socket kill-server
    python3 "$dir/ansi2png.py" "$work/$name.ansi" "$root/doc/img/$name.png" "$cols" "$rows"
    echo "doc/img/$name.png"
}

# In a container, with a generic home.
container="$engine run --rm -it --hostname devbox --network host -e XDG_CONFIG_HOME=/home/me/.config -v $work:/work -w /work devdash-screenshots /work/devdash"
shot default 130 26 6 "$container"
shot localhost 150 45 6 "$container -c ./localhost.yml"
shot github 150 40 12 "$container -c ./github.yml"

# The Google services, with the fake API and a keyfile of a fake service account.
port=$(python3 -c "import socket; s = socket.socket(); s.bind(('127.0.0.1', 0)); print(s.getsockname()[1])")
python3 "$dir/fakegoogle.py" "$port" "$work/requests.log" &
fake=$!
openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out "$work/key.pem" 2> /dev/null
python3 - "$work" "$port" << 'PY'
import json, sys
work, port = sys.argv[1], sys.argv[2]
json.dump({"type": "service_account", "private_key": open(work + "/key.pem").read(),
           "client_email": "devdash@example.iam.gserviceaccount.com", "token_uri": f"http://127.0.0.1:{port}/token"},
          open(work + "/my-keyfile.json", "w"))
PY
sleep 1

google="cd $work && DEVDASH_GOOGLE_API_ENDPOINT=http://127.0.0.1:$port/ ./devdash"
shot google-analytics 150 45 6 "$google -c ./google-analytics.yml"
shot search-console 150 40 6 "$google -c ./search-console.yml"
shot blog 150 45 8 "$google -c ./blog.yml"
