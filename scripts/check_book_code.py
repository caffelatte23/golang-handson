#!/usr/bin/env python3
"""本文 (book/*.md) 中の Go サンプルプログラムを実際に実行して検証するスクリプト。

- ```go ブロックのうち `package main` と `func main()` を含み、標準ライブラリだけを
  import しているものを `gofmt -l` / `go vet` / `go run` で検査する
- ブロックの直後 (空行のみを挟む) に ```text ブロックがあれば、それを期待出力として比較する
- 期待出力がある場合、プログラムが異常終了しても (panic の例など) 出力が一致すれば成功とする
- ブロック直前の行に <!-- nocheck --> があるものはスキップする
  (意図的にコンパイルエラーや panic を起こす例など)

使い方: python3 scripts/check_book_code.py [--fix] [book/02-basics.md ...]
  --fix: gofmt されていないブロックを gofmt の結果で書き換える
"""
import pathlib
import re
import subprocess
import sys
import tempfile

FENCE = re.compile(r"^```(\w*)\s*$")


def blocks(text):
    lines = text.split("\n")
    i, out = 0, []
    while i < len(lines):
        m = FENCE.match(lines[i])
        if m and m.group(1):
            lang, start = m.group(1), i
            j = i + 1
            while j < len(lines) and lines[j].strip() != "```":
                j += 1
            out.append((lang, start, "\n".join(lines[i + 1 : j]), j))
            i = j + 1
        else:
            i += 1
    return lines, out


def stdlib_only(src):
    imports = re.findall(r'^\s*(?:import\s+)?(?:\w+\s+)?"([^"]+)"', src, re.M)
    return all("." not in imp.split("/")[0] for imp in imports)


def gofmt_blocks(md):
    """検査対象と同じ ```go ブロック (package main と func main を含むもの) を gofmt で整形して書き戻す。"""
    lines, bs = blocks(md.read_text())
    for lang, start, src, end in reversed(bs):
        if lang != "go" or "package main" not in src or "func main()" not in src:
            continue
        r = subprocess.run(["gofmt"], input=src + "\n", capture_output=True, text=True)
        if r.returncode == 0 and r.stdout.rstrip("\n") != src:
            lines[start + 1 : end] = r.stdout.rstrip("\n").split("\n")
    md.write_text("\n".join(lines))


def main(paths):
    fix = "--fix" in paths
    paths = [p for p in paths if p != "--fix"]
    if fix:
        for md in [pathlib.Path(p) for p in paths] or sorted(pathlib.Path("book").glob("*.md")):
            gofmt_blocks(md)
    files = [pathlib.Path(p) for p in paths] or sorted(pathlib.Path("book").glob("*.md"))
    failed = checked = skipped = 0
    for md in files:
        lines, bs = blocks(md.read_text())
        for idx, (lang, start, src, end) in enumerate(bs):
            if lang != "go" or "package main" not in src or "func main()" not in src:
                continue
            if start > 0 and "nocheck" in lines[start - 1]:
                skipped += 1
                continue
            if not stdlib_only(src):
                skipped += 1
                continue
            expected = None
            if idx + 1 < len(bs):
                nlang, nstart, nsrc, _ = bs[idx + 1]
                if nlang == "text" and all(not l.strip() for l in lines[end + 1 : nstart]):
                    expected = nsrc
            checked += 1
            loc = f"{md}:{start + 1}"
            with tempfile.TemporaryDirectory() as d:
                pathlib.Path(d, "go.mod").write_text("module example\n\ngo 1.24\n")
                pathlib.Path(d, "main.go").write_text(src + "\n")
                fmt = subprocess.run(["gofmt", "-l", "main.go"], cwd=d, capture_output=True, text=True)
                if fmt.stdout.strip():
                    failed += 1
                    diff = subprocess.run(["gofmt", "-d", "main.go"], cwd=d, capture_output=True, text=True)
                    print(f"✗ {loc}: not gofmt-ed\n{diff.stdout}")
                    continue
                vet = subprocess.run(["go", "vet", "."], cwd=d, capture_output=True, text=True)
                if vet.returncode != 0:
                    failed += 1
                    print(f"✗ {loc}: go vet failed\n{vet.stderr}")
                    continue
                try:
                    run = subprocess.run(["go", "run", "."], cwd=d, capture_output=True, text=True, timeout=60)
                except subprocess.TimeoutExpired:
                    failed += 1
                    print(f"✗ {loc}: timeout")
                    continue
                if run.returncode != 0 and expected is None:
                    failed += 1
                    print(f"✗ {loc}: exit {run.returncode}\n{run.stderr}")
                    continue
                if expected is not None:
                    got = (run.stdout + (run.stderr if run.returncode != 0 else "")).rstrip()
                    if got != expected.rstrip():
                        failed += 1
                        print(f"✗ {loc}: output mismatch\n--- want\n{expected}\n--- got\n{got}\n")
                        continue
    print(f"checked {checked} programs, {failed} failed, {skipped} skipped (nocheck or non-stdlib imports)")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
