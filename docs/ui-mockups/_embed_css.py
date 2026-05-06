import pathlib

base = pathlib.Path(__file__).resolve().parent
css = (base / "shared.css").read_text(encoding="utf-8")
indented = "\n".join(("    " + line) if line.strip() else "" for line in css.splitlines())
block = "  <style>\n" + indented + "\n  </style>"
old = '  <link rel="stylesheet" href="shared.css" />'
for f in sorted(base.glob("*.html")):
    if f.name.startswith("_"):
        continue
    t = f.read_text(encoding="utf-8")
    if old not in t:
        print("skip", f.name)
        continue
    f.write_text(t.replace(old, block, 1), encoding="utf-8")
    print("ok", f.name)
