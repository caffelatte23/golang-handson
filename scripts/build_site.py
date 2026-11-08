#!/usr/bin/env python3
"""book/ の Markdown から mdBook 用のソース (site/src) を生成する。

- SUMMARY.md (目次) を章のファイル名と見出しから作る
- book/ の外を指す相対リンク (../exercises/ など) は GitHub 上の URL に書き換える
  (Web サイトにはリポジトリのファイルが含まれないため)
- 各ページの先頭に「非公式・個人用」の注意書きを入れる

使い方: python3 scripts/build_site.py && mdbook build site
"""
import pathlib
import re
import shutil

ROOT = pathlib.Path(__file__).resolve().parent.parent
BOOK = ROOT / "book"
SRC = ROOT / "site" / "src"
REPO = "https://github.com/caffelatte23/golang-handson"

PARTS = [
    ("第1部 基礎編", range(1, 10)),
    ("第2部 実践編 (OSS の調査)", range(10, 14)),
    ("第3部 ハンズオン", range(14, 19)),
]

BANNER = (
    "> ⚠️ **非公式・個人用のドキュメントです。** Go プロジェクトとは関係がなく、"
    "AI (Claude) の支援を受けて執筆しています。正確な情報は "
    "[go.dev](https://go.dev/doc/) を参照してください。\n\n"
)

LINK = re.compile(r"\]\((\.\./[^)#\s]+)(#[^)\s]*)?\)")


def rewrite_links(text: str) -> str:
    def rep(m: re.Match) -> str:
        rel, anchor = m.group(1), m.group(2) or ""
        target = (BOOK / rel).resolve()
        try:
            path = target.relative_to(ROOT).as_posix()
        except ValueError:
            return m.group(0)
        kind = "tree" if target.is_dir() else "blob"
        return f"]({REPO}/{kind}/main/{path}{anchor})"

    return LINK.sub(rep, text)


def title(md: pathlib.Path) -> str:
    for line in md.read_text().splitlines():
        if line.startswith("# "):
            return line[2:].strip()
    return md.stem


def main() -> None:
    if SRC.exists():
        shutil.rmtree(SRC)
    SRC.mkdir(parents=True)

    chapters = {int(p.name[:2]): p for p in BOOK.glob("[0-9][0-9]-*.md")}
    for p in chapters.values():
        text = rewrite_links(p.read_text())
        # 最初の見出しの直後に注意書きを入れる (第0章は本文に同じ注意書きがあるので除く)
        if not p.name.startswith("00-"):
            lines = text.split("\n", 1)
            text = lines[0] + "\n\n" + BANNER + (lines[1] if len(lines) > 1 else "")
        (SRC / p.name).write_text(text)

    summary = ["# Summary", "", f"[{title(chapters[0])}]({chapters[0].name})", ""]
    for part, nums in PARTS:
        summary.append(f"# {part}")
        summary.append("")
        for n in nums:
            if n in chapters:
                summary.append(f"- [{title(chapters[n])}]({chapters[n].name})")
        summary.append("")
    summary.append("---")
    summary.append("")
    summary.append(f"[{title(chapters[99])}]({chapters[99].name})")
    (SRC / "SUMMARY.md").write_text("\n".join(summary) + "\n")
    print(f"generated {len(chapters)} pages into {SRC.relative_to(ROOT)}")


if __name__ == "__main__":
    main()
