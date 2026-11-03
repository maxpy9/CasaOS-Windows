"""Draws the pictures of the README: vector diagrams, the animated demo and the social preview.

Usage: readme_art.py <folder with real screenshots> <output folder>

All text comes from the STRINGS table, once per language. The demo and the social preview are
built only from real screenshots of the program.
"""
import html
import pathlib
import sys

from PIL import Image, ImageDraw, ImageFilter, ImageFont

PICS = pathlib.Path(sys.argv[1])
OUT = pathlib.Path(sys.argv[2])
OUT.mkdir(parents=True, exist_ok=True)

BG = "#0d1420"
PANEL = "#16202f"
PANEL2 = "#1d2a3d"
LINE = "#2d3d55"
TEXT = "#f2f6fb"
MUTED = "#9fb0c6"
BLUE = "#3b82f6"
CYAN = "#5ee3ff"
GREEN = "#34d399"
AMBER = "#f5b342"
RED = "#ef5a5f"
VIOLET = "#a78bfa"
FONT = "'Segoe UI', -apple-system, BlinkMacSystemFont, Roboto, 'Helvetica Neue', Arial, sans-serif"
MONO = "Consolas, 'SF Mono', Menlo, 'DejaVu Sans Mono', monospace"

STRINGS = {
    "en": {
        "hero_title": "CasaOS for Windows",
        "hero_sub": "Your personal cloud on your own PC. On when you press the button, off when you don't.",
        "hero_pill": "UNOFFICIAL PORT",
        "hero_a": "Normal installer",
        "hero_b": "No admin rights",
        "hero_c": "Nothing in the background",
        "arch_title": "Six services, one window in charge",
        "arch_launcher": "CasaOS.exe",
        "arch_launcher2": "starts, watches and stops everything",
        "arch_browser": "Your browser",
        "arch_browser2": "this PC or your network",
        "arch_gateway": "casaos-gateway",
        "arch_gateway2": "single entry point, login check, dashboard pages",
        "svc": [("message-bus", "live events"), ("user-service", "accounts, login"), ("casaos", "files, system info"), ("local-storage", "disks, volumes"), ("app-management", "App Store")],
        "arch_docker": "Docker Desktop",
        "arch_docker2": "runs the apps as containers",
        "arch_job": "Windows Job Object: if the window dies, every service dies with it",
        "life_title": "The biggest difference from CasaOS on Linux",
        "life_linux": "CasaOS on Linux",
        "life_linux2": "systemd services, always on",
        "life_win": "This port",
        "life_win2": "runs only between Start and Stop",
        "life_boot": "Boot",
        "life_now": "Time",
        "life_start": "Start",
        "life_stop": "Stop",
        "life_on": "server on",
        "life_all": "running from boot to shutdown",
        "life_off": "0 processes",
        "paths_title": "CasaOS still thinks it is on Linux",
        "paths_sub": "The dashboard and the APIs keep their Linux paths. A small package, vfs, translates them on the fly.",
        "paths_left": "What CasaOS asks for",
        "paths_right": "What Windows opens",
        "paths_mid": "vfs",
        "paths_rows": [("/DATA/Documents", "%USERPROFILE%\\CasaOS\\Documents"), ("/media/C/Users", "C:\\Users"), ("/media/D/Movies", "D:\\Movies"), ("/var/lib/casaos", "%LOCALAPPDATA%\\CasaOS\\root\\var\\lib\\casaos")],
        "map_title": "What had to be replaced",
        "map_left": "On Linux",
        "map_right": "On Windows",
        "map_rows": [("systemd", "CasaOS.exe supervisor + process snapshot"), ("lsblk", "DeviceIoControl, same JSON output"), ("/proc/net", "IP Helper API (TCP and UDP tables)"), ("bash network scripts", "gopsutil, virtual adapters filtered out"), ("/etc/localtime", "Windows Registry, converted to IANA"), ("Unix socket of the message bus", "TCP port"), ("/DATA volumes of the apps", "Windows paths, shared by Docker Desktop")],
        "time_title": "How it was made",
        "time": [("Clone", "7 Go repositories plus the prebuilt dashboard"), ("Paths", "vfs maps Linux paths to Windows"), ("Replace", "Every Linux-only call gets a Windows version"), ("Launcher", "Native Win32 window, no Electron"), ("Installer", "Per-user wizard with Inno Setup"), ("Real runs", "Tested on a real PC, bugs fixed"), ("Review", "Login hardened, embedded keys removed")],
        "demo": ["Install, open, press one button", "The server is on: open the link", "The full CasaOS dashboard", "Your files, and every disk of the PC", "The App Store catalog (needs Docker Desktop for apps)"],
        "social_title": "CasaOS for Windows",
        "social_sub": "Personal cloud on your PC, on and off with one button",
        "social_pill": "Unofficial port · Open source",
    },
    "it": {
        "hero_title": "CasaOS per Windows",
        "hero_sub": "Il tuo cloud personale sul tuo PC. Acceso quando premi il pulsante, spento quando non serve.",
        "hero_pill": "PORTING NON UFFICIALE",
        "hero_a": "Installer normale",
        "hero_b": "Niente admin",
        "hero_c": "Niente in background",
        "arch_title": "Sei servizi, una finestra che comanda",
        "arch_launcher": "CasaOS.exe",
        "arch_launcher2": "avvia, controlla e ferma tutto",
        "arch_browser": "Il tuo browser",
        "arch_browser2": "da questo PC o dalla rete",
        "arch_gateway": "casaos-gateway",
        "arch_gateway2": "ingresso unico, controllo login, pagine della dashboard",
        "svc": [("message-bus", "eventi in tempo reale"), ("user-service", "account, login"), ("casaos", "file, stato sistema"), ("local-storage", "dischi, volumi"), ("app-management", "App Store")],
        "arch_docker": "Docker Desktop",
        "arch_docker2": "esegue le app come container",
        "arch_job": "Job Object di Windows: se la finestra muore, tutti i servizi muoiono con lei",
        "life_title": "La differenza più grande rispetto a CasaOS su Linux",
        "life_linux": "CasaOS su Linux",
        "life_linux2": "servizi systemd, sempre attivi",
        "life_win": "Questo porting",
        "life_win2": "gira solo tra Avvia e Ferma",
        "life_boot": "Accensione",
        "life_now": "Tempo",
        "life_start": "Avvia",
        "life_stop": "Ferma",
        "life_on": "server acceso",
        "life_all": "attivo dall'accensione allo spegnimento",
        "life_off": "0 processi",
        "paths_title": "CasaOS crede ancora di essere su Linux",
        "paths_sub": "Dashboard e API tengono i loro percorsi Linux. Un piccolo pacchetto, vfs, li traduce al volo.",
        "paths_left": "Cosa chiede CasaOS",
        "paths_right": "Cosa apre Windows",
        "paths_mid": "vfs",
        "paths_rows": [("/DATA/Documents", "%USERPROFILE%\\CasaOS\\Documents"), ("/media/C/Users", "C:\\Users"), ("/media/D/Film", "D:\\Film"), ("/var/lib/casaos", "%LOCALAPPDATA%\\CasaOS\\root\\var\\lib\\casaos")],
        "map_title": "Cosa è stato sostituito",
        "map_left": "Su Linux",
        "map_right": "Su Windows",
        "map_rows": [("systemd", "Supervisore in CasaOS.exe + elenco processi"), ("lsblk", "DeviceIoControl, stesso JSON in uscita"), ("/proc/net", "API IP Helper (tabelle TCP e UDP)"), ("script bash di rete", "gopsutil, schede virtuali filtrate"), ("/etc/localtime", "Registro di Windows, convertito in IANA"), ("socket Unix del message bus", "Porta TCP"), ("volumi /DATA delle app", "Percorsi Windows, condivisi da Docker Desktop")],
        "time_title": "Come è stato creato",
        "time": [("Clone", "7 repository Go più la dashboard già compilata"), ("Percorsi", "vfs traduce i percorsi Linux in Windows"), ("Sostituzioni", "Ogni chiamata solo-Linux ha una versione Windows"), ("Launcher", "Finestra nativa Win32, niente Electron"), ("Installer", "Procedura guidata per utente con Inno Setup"), ("Prove reali", "Provato su un PC vero, bug corretti"), ("Revisione", "Login rafforzato, chiavi integrate rimosse")],
        "demo": ["Installa, apri, premi un pulsante", "Il server è acceso: apri il link", "La dashboard completa di CasaOS", "I tuoi file e tutti i dischi del PC", "Il catalogo dell'App Store (per le app serve Docker Desktop)"],
        "social_title": "CasaOS per Windows",
        "social_sub": "Cloud personale sul tuo PC, acceso e spento con un pulsante",
        "social_pill": "Porting non ufficiale · Open source",
    },
}


def esc(s):
    return html.escape(s, quote=True)


def text(x, y, s, size=16, fill=TEXT, weight=400, anchor="start", family=FONT):
    return f'<text x="{x}" y="{y}" font-family="{family}" font-size="{size}" font-weight="{weight}" fill="{fill}" text-anchor="{anchor}">{esc(s)}</text>'


def box(x, y, w, h, fill=PANEL, stroke=LINE, r=14, extra=""):
    return f'<rect x="{x}" y="{y}" width="{w}" height="{h}" rx="{r}" fill="{fill}" stroke="{stroke}" stroke-width="1.5" {extra}/>'


def svg(w, h, body, style=""):
    motion = "@media (prefers-reduced-motion: reduce){*{animation:none!important}}"
    return (
        f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {w} {h}" width="{w}" height="{h}" role="img">'
        f"<style>{style}{motion}</style>"
        f'<defs><linearGradient id="bg" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#0d1420"/><stop offset="1" stop-color="#101d33"/></linearGradient>'
        f'<linearGradient id="accent" x1="0" y1="0" x2="1" y2="0"><stop offset="0" stop-color="{BLUE}"/><stop offset="1" stop-color="{CYAN}"/></linearGradient>'
        f'<marker id="arrow" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse"><path d="M0 0L10 5L0 10z" fill="{MUTED}"/></marker></defs>'
        f'<rect width="{w}" height="{h}" rx="20" fill="url(#bg)"/>{body}</svg>'
    )


def hero(t):
    w, h = 1280, 420
    # the power button: its ring lights up and goes dark, like the server
    cx, cy = 1040, 205
    style = (
        "@keyframes pulse{0%,100%{opacity:.18;transform:scale(.92)}50%{opacity:.55;transform:scale(1.08)}}"
        ".glow{transform-origin:1040px 205px;animation:pulse 4s ease-in-out infinite}"
        "@keyframes dash{to{stroke-dashoffset:0}}"
        ".ring{stroke-dasharray:540;stroke-dashoffset:540;animation:dash 2.4s ease-out forwards}"
    )
    b = [
        f'<circle class="glow" cx="{cx}" cy="{cy}" r="150" fill="{CYAN}" opacity=".25" filter="url(#blur)"/>',
        f'<circle cx="{cx}" cy="{cy}" r="86" fill="{PANEL}" stroke="{LINE}" stroke-width="2"/>',
        f'<circle class="ring" cx="{cx}" cy="{cy}" r="86" fill="none" stroke="url(#accent)" stroke-width="5" stroke-linecap="round" transform="rotate(-90 {cx} {cy})"/>',
        f'<path d="M{cx - 27} {cy - 24}a38 38 0 1 0 54 0" fill="none" stroke="{CYAN}" stroke-width="9" stroke-linecap="round"/>',
        f'<line x1="{cx}" y1="{cy - 44}" x2="{cx}" y2="{cy - 3}" stroke="{CYAN}" stroke-width="9" stroke-linecap="round"/>',
    ]
    pill_w = 19 + len(t["hero_pill"]) * 10.2
    b.append(box(70, 74, pill_w, 34, fill="#3a2a0f", stroke=AMBER, r=17))
    b.append(text(70 + pill_w / 2, 97, t["hero_pill"], 15, AMBER, 700, "middle"))
    b.append(text(70, 182, t["hero_title"], 66, TEXT, 700))
    words = t["hero_sub"].split()
    lines, cur = [], ""
    for word in words:
        if len(cur) + len(word) > 52:
            lines.append(cur)
            cur = word
        else:
            cur = (cur + " " + word).strip()
    lines.append(cur)
    for i, line in enumerate(lines):
        b.append(text(70, 228 + i * 32, line, 23, MUTED))
    x = 70
    for key, color in (("hero_a", GREEN), ("hero_b", CYAN), ("hero_c", VIOLET)):
        label = t[key]
        pw = 42 + len(label) * 9.1
        b.append(box(x, 320, pw, 42, fill=PANEL2, stroke=LINE, r=21))
        b.append(f'<circle cx="{x + 20}" cy="341" r="5" fill="{color}"/>')
        b.append(text(x + 33, 347, label, 16.5, TEXT, 500))
        x += pw + 12
    defs = '<defs><filter id="blur" x="-50%" y="-50%" width="200%" height="200%"><feGaussianBlur stdDeviation="38"/></filter></defs>'
    return svg(w, h, defs + "".join(b), style)


def architecture(t):
    w, h = 1280, 720
    style = (
        "@keyframes flow{to{stroke-dashoffset:-28}}"
        ".flow{stroke-dasharray:6 8;animation:flow 1.1s linear infinite}"
    )
    b = [text(640, 56, t["arch_title"], 30, TEXT, 700, "middle")]
    # top row
    b.append(box(60, 90, 420, 104, fill="#14263f", stroke=BLUE))
    b.append(text(270, 132, t["arch_launcher"], 24, TEXT, 700, "middle"))
    b.append(text(270, 164, t["arch_launcher2"], 16.5, MUTED, 400, "middle"))
    b.append(box(800, 90, 420, 104))
    b.append(text(1010, 132, t["arch_browser"], 24, TEXT, 700, "middle"))
    b.append(text(1010, 164, t["arch_browser2"], 16.5, MUTED, 400, "middle"))
    # gateway
    b.append(box(190, 256, 900, 92, fill="#102c3a", stroke=CYAN))
    b.append(text(640, 294, t["arch_gateway"], 24, TEXT, 700, "middle"))
    b.append(text(640, 325, t["arch_gateway2"], 16.5, MUTED, 400, "middle"))
    b.append(f'<path class="flow" d="M270 194V256" stroke="{BLUE}" stroke-width="2.5" fill="none"/>')
    b.append(f'<path class="flow" d="M1010 194V256" stroke="{CYAN}" stroke-width="2.5" fill="none"/>')
    # services
    n = len(t["svc"])
    cw, gap, x0 = 216, 20, 60
    for i, (name, what) in enumerate(t["svc"]):
        x = x0 + i * (cw + gap)
        b.append(f'<path class="flow" d="M{x + cw / 2} 348V416" stroke="{MUTED}" stroke-width="2" fill="none"/>')
        b.append(box(x, 416, cw, 96))
        b.append(text(x + cw / 2, 456, name, 18.5, TEXT, 600, "middle"))
        b.append(text(x + cw / 2, 486, what, 15, MUTED, 400, "middle"))
    # docker, below app-management
    x = x0 + (n - 1) * (cw + gap)
    b.append(f'<path class="flow" d="M{x + cw / 2} 512V566" stroke="{VIOLET}" stroke-width="2.5" fill="none"/>')
    b.append(box(x - 110, 566, cw + 110, 96, fill="#1f1a38", stroke=VIOLET))
    b.append(text(x - 110 + (cw + 110) / 2, 606, t["arch_docker"], 22, TEXT, 700, "middle"))
    b.append(text(x - 110 + (cw + 110) / 2, 636, t["arch_docker2"], 15.5, MUTED, 400, "middle"))
    # the job object
    b.append(f'<rect x="60" y="566" width="760" height="96" rx="14" fill="none" stroke="{GREEN}" stroke-width="1.5" stroke-dasharray="6 6"/>')
    b.append(f'<circle cx="92" cy="614" r="7" fill="{GREEN}"/>')
    b.append(text(112, 620, t["arch_job"], 17.5, TEXT, 500))
    return svg(w, h, "".join(b), style)


def lifecycle(t):
    w, h = 1280, 450
    style = (
        "@keyframes grow{from{transform:scaleX(0)}to{transform:scaleX(1)}}"
        ".bar{transform-box:fill-box;transform-origin:left;animation:grow 1.6s ease-out both}"
    )
    b = [text(640, 56, t["life_title"], 30, TEXT, 700, "middle")]
    x0, x1 = 300, 1210
    span = x1 - x0
    # linux
    b.append(text(60, 138, t["life_linux"], 22, TEXT, 700))
    b.append(text(60, 166, t["life_linux2"], 15.5, MUTED))
    b.append(box(x0, 116, span, 46, fill=PANEL, stroke=LINE, r=10))
    b.append(f'<rect class="bar" x="{x0}" y="116" width="{span}" height="46" rx="10" fill="{BLUE}"/>')
    b.append(text(x0 + span / 2, 146, t["life_all"], 17, "#fff", 600, "middle"))
    # windows
    y = 240
    b.append(text(60, y + 22, t["life_win"], 22, TEXT, 700))
    b.append(text(60, y + 50, t["life_win2"], 15.5, MUTED))
    b.append(box(x0, y, span, 46, fill=PANEL, stroke=LINE, r=10))
    for a, c in ((0.12, 0.40), (0.68, 0.86)):
        bx, bw = x0 + a * span, (c - a) * span
        b.append(f'<rect class="bar" x="{bx:.0f}" y="{y}" width="{bw:.0f}" height="46" rx="10" fill="{GREEN}"/>')
        b.append(text(bx + bw / 2, y + 30, t["life_on"], 16, "#06281c", 700, "middle"))
        b.append(f'<line x1="{bx:.0f}" y1="{y + 46}" x2="{bx:.0f}" y2="{y + 72}" stroke="{MUTED}" stroke-width="1.5"/>')
        b.append(text(bx, y + 92, t["life_start"], 15, MUTED, 500, "middle"))
        ex = bx + bw
        b.append(f'<line x1="{ex:.0f}" y1="{y + 46}" x2="{ex:.0f}" y2="{y + 72}" stroke="{MUTED}" stroke-width="1.5"/>')
        b.append(text(ex, y + 92, t["life_stop"], 15, MUTED, 500, "middle"))
    for a, c in ((0, 0.12), (0.40, 0.68), (0.86, 1)):
        b.append(text(x0 + (a + c) / 2 * span, y + 29, t["life_off"], 14.5, MUTED, 500, "middle"))
    # axis
    ay = 392
    b.append(f'<line x1="{x0}" y1="{ay}" x2="{x1}" y2="{ay}" stroke="{LINE}" stroke-width="2" marker-end="url(#arrow)"/>')
    b.append(text(x0, ay + 26, t["life_boot"], 15, MUTED))
    b.append(text(x1, ay + 26, t["life_now"], 15, MUTED, 400, "end"))
    return svg(w, h, "".join(b), style)


def paths(t):
    rows = t["paths_rows"]
    h = 196 + len(rows) * 74
    w = 1280
    style = "@keyframes flow{to{stroke-dashoffset:-28}}.flow{stroke-dasharray:6 8;animation:flow 1.1s linear infinite}"
    b = [text(640, 56, t["paths_title"], 30, TEXT, 700, "middle"), text(640, 92, t["paths_sub"], 17.5, MUTED, 400, "middle")]
    b.append(text(60, 146, t["paths_left"], 16, CYAN, 600))
    b.append(text(560, 146, t["paths_right"], 16, GREEN, 600))
    for i, (a, c) in enumerate(rows):
        y = 164 + i * 74
        b.append(box(60, y, 370, 58))
        b.append(text(82, y + 37, a, 19, TEXT, 500, family=MONO))
        b.append(f'<path class="flow" d="M430 {y + 29}H560" stroke="{MUTED}" stroke-width="2" fill="none" marker-end="url(#arrow)"/>')
        b.append(box(462, y + 13, 66, 32, fill="#14263f", stroke=BLUE, r=16))
        b.append(text(495, y + 35, t["paths_mid"], 15.5, TEXT, 700, "middle", MONO))
        b.append(box(560, y, 660, 58))
        b.append(text(582, y + 37, c, 19, TEXT, 500, family=MONO))
    return svg(w, h, "".join(b), style)


def replaced(t):
    rows = t["map_rows"]
    w, h = 1280, 150 + len(rows) * 66
    b = [text(640, 56, t["map_title"], 30, TEXT, 700, "middle")]
    b.append(text(60, 106, t["map_left"], 16, RED, 600))
    b.append(text(520, 106, t["map_right"], 16, GREEN, 600))
    for i, (a, c) in enumerate(rows):
        y = 122 + i * 66
        b.append(box(60, y, 400, 54, fill="#2a1a22", stroke="#5b2d3a"))
        b.append(text(80, y + 34, a, 17.5, TEXT, 500))
        b.append(f'<line x1="464" y1="{y + 27}" x2="514" y2="{y + 27}" stroke="{MUTED}" stroke-width="2" marker-end="url(#arrow)"/>')
        b.append(box(520, y, 700, 54, fill="#10291f", stroke="#1f5a43"))
        b.append(text(540, y + 34, c, 17.5, TEXT, 500))
    return svg(w, h, "".join(b))


def timeline(t):
    steps = t["time"]
    cols = 4
    rows = (len(steps) + cols - 1) // cols
    cw, ch, gx, gy = 275, 150, 26, 44
    w = 1280
    x0 = (w - (cols * cw + (cols - 1) * gx)) / 2
    h = 110 + rows * ch + (rows - 1) * gy + 40
    style = (
        "@keyframes rise{from{opacity:0;transform:translateY(10px)}to{opacity:1;transform:none}}"
        ".step{animation:rise .6s ease-out both}"
    )
    b = [text(640, 56, t["time_title"], 30, TEXT, 700, "middle")]
    colors = [BLUE, CYAN, GREEN, VIOLET, AMBER, RED, "#f472b6"]
    for i, (name, what) in enumerate(steps):
        r, c = i // cols, i % cols
        x, y = x0 + c * (cw + gx), 92 + r * (ch + gy)
        delay = i * 0.18
        words, lines, cur = what.split(), [], ""
        for word in words:
            if len(cur) + len(word) > 29:
                lines.append(cur)
                cur = word
            else:
                cur = (cur + " " + word).strip()
        lines.append(cur)
        g = [box(x, y, cw, ch)]
        g.append(f'<circle cx="{x + 34}" cy="{y + 36}" r="18" fill="{colors[i]}"/>')
        g.append(text(x + 34, y + 42, str(i + 1), 17, "#0d1420", 800, "middle"))
        g.append(text(x + 62, y + 43, name, 20.5, TEXT, 700))
        for j, line in enumerate(lines[:3]):
            g.append(text(x + 20, y + 84 + j * 24, line, 15.5, MUTED))
        b.append(f'<g class="step" style="animation-delay:{delay:.2f}s">{"".join(g)}</g>')
        if c < cols - 1 and i < len(steps) - 1:
            b.append(f'<line x1="{x + cw + 3}" y1="{y + ch / 2}" x2="{x + cw + gx - 3}" y2="{y + ch / 2}" stroke="{MUTED}" stroke-width="2" marker-end="url(#arrow)"/>')
    return svg(w, int(h), "".join(b), style)


# ---------------------------------------------- pictures from real screenshots


def font(name, size):
    return ImageFont.truetype(f"C:/Windows/Fonts/{name}", size)


def rounded(img, radius):
    mask = Image.new("L", img.size, 0)
    ImageDraw.Draw(mask).rounded_rectangle((0, 0, img.size[0] - 1, img.size[1] - 1), radius, fill=255)
    out = img.convert("RGBA")
    out.putalpha(mask)
    return out


def shadow_paste(canvas, img, xy, radius=14):
    x, y = xy
    layer = Image.new("RGBA", canvas.size, (0, 0, 0, 0))
    ImageDraw.Draw(layer).rounded_rectangle((x + 4, y + 10, x + img.size[0] + 4, y + img.size[1] + 10), radius, fill=(0, 0, 0, 170))
    layer = layer.filter(ImageFilter.GaussianBlur(16))
    canvas.alpha_composite(layer)
    canvas.alpha_composite(rounded(img, radius), (x, y))


def hide_address(img):
    """Blurs the private network address shown in the window."""
    img = img.convert("RGB")
    region = (40, 388, 240, 412)
    img.paste(img.crop(region).filter(ImageFilter.GaussianBlur(7)), region[:2])
    return img


def load():
    return {
        "stopped": Image.open(PICS / "launcher-stopped.png").convert("RGB"),
        "running": hide_address(Image.open(PICS / "launcher-running.png")),
        "home": Image.open(PICS / "dash-home.png").convert("RGB"),
        "files": Image.open(PICS / "dash-files.png").convert("RGB"),
        "store": Image.open(PICS / "dash-store.png").convert("RGB"),
    }


def demo(t, shots, path):
    W, H = 1280, 720
    base = Image.new("RGBA", (W, H), (13, 20, 32, 255))
    grad = Image.new("RGBA", (W, H), (0, 0, 0, 0))
    gd = ImageDraw.Draw(grad)
    gd.ellipse((-200, -300, 700, 500), fill=(59, 130, 246, 45))
    gd.ellipse((800, 300, 1600, 1100), fill=(94, 227, 255, 28))
    base.alpha_composite(grad.filter(ImageFilter.GaussianBlur(120)))
    cap_font = font("seguisb.ttf", 30)
    sw, sh = 800, 500

    def frame(launcher, screen, caption):
        f = base.copy()
        lh = 560
        lw = int(launcher.size[0] * lh / launcher.size[1])
        if screen is None:
            shadow_paste(f, launcher.resize((lw, lh), Image.LANCZOS), ((W - lw) // 2, 34))
        else:
            shadow_paste(f, launcher.resize((lw, lh), Image.LANCZOS), (34, 34))
            shadow_paste(f, screen.resize((sw, sh), Image.LANCZOS), (W - sw - 34, 64))
        d = ImageDraw.Draw(f)
        d.text((W // 2, 650), caption, font=cap_font, fill=(242, 246, 251, 255), anchor="mm")
        return f.convert("RGB")

    c = t["demo"]
    scenes = [
        frame(shots["stopped"], None, c[0]),
        frame(shots["running"], None, c[1]),
        frame(shots["running"], shots["home"], c[2]),
        frame(shots["running"], shots["files"], c[3]),
        frame(shots["running"], shots["store"], c[4]),
    ]
    frames, times = [], []
    for i, scene in enumerate(scenes):
        frames.append(scene)
        times.append(2600)
        nxt = scenes[(i + 1) % len(scenes)]
        for k in (1, 2, 3, 4):
            frames.append(Image.blend(scene, nxt, k / 5))
            times.append(70)
    frames[0].save(path, save_all=True, append_images=frames[1:], duration=times, loop=0, quality=80, method=4)
    scenes[2].save(path.with_name(path.stem + "-still.jpg"), quality=88)


def social(t, shots, path):
    W, H = 1280, 640
    c = Image.new("RGBA", (W, H), (13, 20, 32, 255))
    grad = Image.new("RGBA", (W, H), (0, 0, 0, 0))
    gd = ImageDraw.Draw(grad)
    gd.ellipse((-200, -300, 700, 500), fill=(59, 130, 246, 60))
    gd.ellipse((800, 200, 1600, 1000), fill=(94, 227, 255, 40))
    c.alpha_composite(grad.filter(ImageFilter.GaussianBlur(120)))
    d = ImageDraw.Draw(c)
    d.text((70, 150), t["social_pill"], font=font("seguisb.ttf", 26), fill=(245, 179, 66, 255))
    d.text((70, 200), t["social_title"].replace(" for ", "\nfor ").replace(" per ", "\nper "), font=font("segoeuib.ttf", 68), fill=(242, 246, 251, 255), spacing=8)
    sub, lines, cur = t["social_sub"].split(), [], ""
    for word in sub:
        if len(cur) + len(word) > 30:
            lines.append(cur)
            cur = word
        else:
            cur = (cur + " " + word).strip()
    lines.append(cur)
    d.multiline_text((70, 400), "\n".join(lines), font=font("segoeui.ttf", 32), fill=(159, 176, 198, 255), spacing=10)
    lw = 300
    lh = int(shots["running"].size[1] * lw / shots["running"].size[0])
    shadow_paste(c, shots["home"].resize((610, 381), Image.LANCZOS), (W - 640, 150))
    shadow_paste(c, shots["running"].resize((lw, lh), Image.LANCZOS), (W - 700, 130))
    c.convert("RGB").save(path, quality=92)


def main():
    shots = load()
    (OUT / "dashboard.jpg").write_bytes(b"")
    shots["home"].resize((1600, 1000), Image.LANCZOS).save(OUT / "dashboard.jpg", quality=86)
    shots["files"].resize((1600, 1000), Image.LANCZOS).save(OUT / "files.jpg", quality=86)
    shots["store"].resize((1600, 1000), Image.LANCZOS).save(OUT / "app-store.jpg", quality=86)
    shots["stopped"].save(OUT / "launcher-stopped.png", optimize=True)
    shots["running"].save(OUT / "launcher-running.png", optimize=True)
    for lang, t in STRINGS.items():
        for name, fn in (("hero", hero), ("architecture", architecture), ("lifecycle", lifecycle), ("paths", paths), ("replaced", replaced), ("timeline", timeline)):
            (OUT / f"{name}-{lang}.svg").write_text(fn(t), encoding="utf-8")
        demo(t, shots, OUT / f"demo-{lang}.webp")
    social(STRINGS["en"], shots, OUT / "social-preview.png")
    for f in sorted(OUT.iterdir()):
        print(f"{f.name:28} {f.stat().st_size / 1024:8.0f} KB")


main()
