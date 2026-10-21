"""Render the output of `tmux capture-pane -p -e` as a PNG, with the Catppuccin Mocha palette.

Usage: ansi2png.py <capture> <out.png> <columns> <rows>. The fonts can be changed with FONT and BOLD_FONT.
"""
import os
import re
import sys

from PIL import Image, ImageDraw, ImageFont

FONT = os.environ.get("FONT", "/usr/share/fonts/TTF/JetBrainsMonoNerdFontMono-Regular.ttf")
BOLD = os.environ.get("BOLD_FONT", "/usr/share/fonts/TTF/JetBrainsMonoNerdFontMono-Bold.ttf")
SIZE = 15

BG, FG = (30, 30, 46), (205, 214, 244)
PALETTE = [
    (69, 71, 90), (243, 139, 168), (166, 227, 161), (249, 226, 175),
    (137, 180, 250), (245, 194, 231), (148, 226, 213), (186, 194, 222),
    (88, 91, 112), (243, 139, 168), (166, 227, 161), (249, 226, 175),
    (137, 180, 250), (245, 194, 231), (148, 226, 213), (166, 173, 200),
]


def xterm(n):
    if n < 16:
        return PALETTE[n]
    if n < 232:
        n -= 16
        steps = [0, 95, 135, 175, 215, 255]
        return steps[n // 36], steps[n // 6 % 6], steps[n % 6]
    v = 8 + (n - 232) * 10
    return v, v, v


def parse(lines):
    """Cells of each line: (char, fg, bg, bold)."""
    grid = []
    for line in lines:
        fg, bg, bold, rev = None, None, False, False
        cells = []
        for token in re.split(r"(\x1b\[[0-9;]*m)", line):
            if token.startswith("\x1b["):
                codes = [int(c) if c else 0 for c in token[2:-1].split(";")]
                i = 0
                while i < len(codes):
                    c = codes[i]
                    if c == 0:
                        fg, bg, bold, rev = None, None, False, False
                    elif c == 1:
                        bold = True
                    elif c == 22:
                        bold = False
                    elif c == 7:
                        rev = True
                    elif c == 27:
                        rev = False
                    elif 30 <= c <= 37:
                        fg = PALETTE[c - 30]
                    elif 90 <= c <= 97:
                        fg = PALETTE[c - 90 + 8]
                    elif c == 39:
                        fg = None
                    elif 40 <= c <= 47:
                        bg = PALETTE[c - 40]
                    elif 100 <= c <= 107:
                        bg = PALETTE[c - 100 + 8]
                    elif c == 49:
                        bg = None
                    elif c in (38, 48) and i + 2 < len(codes) and codes[i + 1] == 5:
                        col = xterm(codes[i + 2])
                        fg, bg = (col, bg) if c == 38 else (fg, col)
                        i += 2
                    elif c in (38, 48) and i + 4 < len(codes) and codes[i + 1] == 2:
                        col = tuple(codes[i + 2:i + 5])
                        fg, bg = (col, bg) if c == 38 else (fg, col)
                        i += 4
                    i += 1
                continue
            for ch in token:
                f, b = fg or FG, bg or BG
                if rev:
                    f, b = b, f
                cells.append((ch, f, b, bold))
        grid.append(cells)
    return grid


def render(grid, out, cols, rows):
    font = ImageFont.truetype(FONT, SIZE)
    bold = ImageFont.truetype(BOLD, SIZE)
    cw = round(font.getlength("M"))
    ch = round(SIZE * 1.3)
    pad = 18
    img = Image.new("RGB", (cols * cw + 2 * pad, rows * ch + 2 * pad), BG)
    d = ImageDraw.Draw(img)
    for y, cells in enumerate(grid[:rows]):
        for x, (c, f, b, bo) in enumerate(cells[:cols]):
            px, py = pad + x * cw, pad + y * ch
            if b != BG:
                d.rectangle([px, py, px + cw - 1, py + ch - 1], fill=b)
            if c in "─│┌┐└┘":
                # Draw the box lines to join them without gaps between the cells.
                mx, my = px + cw // 2, py + ch // 2
                if c in "─┌└":
                    d.line([mx if c != "─" else px, my, px + cw, my], fill=f)
                if c in "─┐┘":
                    d.line([px, my, mx if c != "─" else px + cw, my], fill=f)
                if c in "│┌┐":
                    d.line([mx, my if c != "│" else py, mx, py + ch], fill=f)
                if c in "│└┘":
                    d.line([mx, py, mx, my if c != "│" else py + ch], fill=f)
            elif c != " ":
                d.text((px, py + (ch - SIZE) // 2 - 1), c, font=bold if bo else font, fill=f)
    img.save(out, optimize=True)


if __name__ == "__main__":
    src, out, cols, rows = sys.argv[1], sys.argv[2], int(sys.argv[3]), int(sys.argv[4])
    grid = parse(open(src, encoding="utf-8").read().split("\n"))
    # Crop the empty rows at the bottom.
    while grid and all(c == " " and b == BG for c, f, b, bo in grid[-1]):
        grid.pop()
    render(grid, out, cols, min(rows, len(grid)))
