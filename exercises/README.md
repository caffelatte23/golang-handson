# 演習問題

本文第3〜8章の演習です。A Tour of Go の演習と同じテーマを、**自動採点のテスト付き** で収録しています。

## 進め方

1. 各ディレクトリの `exercise.go` を開き、`TODO` の部分を実装します。
2. テストを実行し、すべて通れば正解です。

```bash
cd exercises
go test ./sqrt/          # 1 つだけ実行する
go test ./...            # すべて実行する (最初は未実装なので失敗する)
go test -race ./crawler/ # 並行処理の演習には -race を付ける
```

3. 行き詰まったら、`solution.go` の解答例を見るか、解答例のコードでテストを実行してみます。

```bash
go test -tags solution ./sqrt/
```

`exercise.go` と `solution.go` はビルドタグ (`//go:build !solution` と `//go:build solution`) で切り替わります。`-tags solution` を付けると、解答例のほうがコンパイルされます。

## 一覧

| ディレクトリ | 章 | 内容 | 学ぶこと |
| --- | --- | --- | --- |
| [`sqrt`](sqrt/) | 第3章 | ニュートン法で平方根を求める | for、浮動小数点数 |
| [`pic`](pic/) | 第4章 | 2 次元スライスで画像データを作る | make、2 次元スライス |
| [`wordcount`](wordcount/) | 第4章 | 単語の出現回数を数える | map |
| [`fibonacci`](fibonacci/) | 第4章 | フィボナッチ数列を返すクロージャ | 関数値、クロージャ |
| [`stringer`](stringer/) | 第5章 | IP アドレスを文字列で表示する | fmt.Stringer |
| [`reader`](reader/) | 第5章 | 'A' を無限に返す Reader | io.Reader |
| [`rot13`](rot13/) | 第5章 | ROT13 で変換する Reader | Reader のラップ |
| [`image`](image/) | 第5章 | image.Image の実装 | インターフェースの実装 |
| [`generics`](generics/) | 第6章 | Map / Filter / Reduce と連結リスト | 型パラメータ、iter.Seq |
| [`errsqrt`](errsqrt/) | 第7章 | 負の数でエラーを返す Sqrt | error インターフェース |
| [`equivtree`](equivtree/) | 第8章 | 2 つの二分木が同じ値を持つか判定する | goroutine、channel |
| [`crawler`](crawler/) | 第8章 | 並行 Web クローラー | Mutex、WaitGroup |
