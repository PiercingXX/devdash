# Screenshots

`generate.sh` generates the screenshots of the documentation in `doc/img` from the example dashboards:

```shell
./scripts/screenshots/generate.sh
```

It needs tmux, python3 with Pillow, openssl, and podman or docker.

* The dashboards run in tmux. `ansi2png.py` renders the content of tmux, with its colors, as a PNG. Change the font with the environment variables `FONT` and `BOLD_FONT`.
* The localhost and Github dashboards run in a container, not to display the disks and the processes of your computer.
* The Google Analytics and Search Console dashboards display sample data: `fakegoogle.py` is a fake of their APIs, used by DevDash with the environment variable `DEVDASH_GOOGLE_API_ENDPOINT`.
