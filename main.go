package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// Node はハフマン木の1つの節を表す
type Node struct {
	char rune // 文字(葉のときだけ意味を持つ)
	freq int  // 文字の出現頻度
	left *Node  // 左の子ノード
	right *Node // 右の子ノード
}

// isLeafは葉かどうかを判定
func (n *Node) isLeaf() bool {
	return n.left == nil && n.right == nil
}

// buildCodes木を再帰的にたどり，各文字の符号をcodeに書き込む
func buildCodes(node *Node, path string, codes map[rune]string){
	if node == nil {
		return
	}
	// 葉ならここまでの道pathがこの文字の符号
	if node.isLeaf() {
		codes[node.char] = path
		return
	}
	// 葉でなければ左に0,右に1を足して子へ潜っていく
	buildCodes(node.left, path+"0", codes)
	buildCodes(node.right, path+"1", codes)
}

// encode: 文字列を符号表でビット列に変換
func encode(text string, codes map[rune]string) string {
	var sb strings.Builder
	for _, ch := range text {
		sb.WriteString(codes[ch]) // 1文字ずつ符号に置き換えて連結
	}
	return sb.String()
}

// decode: ビット列で木をたどって文字列に戻す
func decode(bits string, root *Node) string {
	var sb strings.Builder
	node := root // 現在地を根に置く
	for _, bit := range bits {
		if bit == '0' {
			node = node.left // 0なら左へ降りる
		} else {
			node = node.right // 1なら右へ降りる
		}
		if node.isLeaf() {
			sb.WriteRune(node.char) // 葉に着いたら文字を出力
			node = root // 根に戻る
		}
	}
	return sb.String()
}

func main() {
	// コマンドライン引数で渡されたファイルを読み込む
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err) // 読み込み失敗時はエラーで止める
	}
	text := string(data) // バイト列を文字列として扱う

	// 文字ごとの出現回数を数える
	freq := make(map[rune]int)
	for _, ch := range text {
		freq[ch]++
	}

	// 下記文字を「葉ノード」にして札の山を作る
	var nodes []*Node
	for ch, count := range freq {
		nodes = append(nodes, &Node{char: ch, freq: count})
	}

	// 札が1枚に成るまで「1番小さい2つを合体」を繰り返す
	for len(nodes) > 1 {
		// 頻度の小さい順に並び替える
		sort.Slice(nodes, func(i,j int) bool {
			return nodes[i].freq < nodes[j].freq
		})

		// 先頭2つを取り出す
		left := nodes[0]
		right := nodes[1]

		// 2つを合体させて新しいノードを作る
		parent := &Node{
			freq: left.freq + right.freq,
			left: left,
			right: right,
		}

		// 使った2枚を山から除き，親を山に戻す
		nodes = append(nodes[2:], parent)

		// fmt.Printf("合体: freq %d + freq %d => freq %d （残り %d枚）\n",
		// 	left.freq, right.freq, parent.freq, len(nodes))
	}

	// 最後に残った1枚が根
	root := nodes[0]

	// 木をたどって符号表を作る
	codes := make(map[rune]string)
	buildCodes(root, "", codes)

	// 符号表を表示
	// fmt.Println("符号表:")
	// for ch, code :=  range codes {
	// 	fmt.Printf("%q → %s （%dビット）\n", ch, code, len(code))
	// }

	// 符号化
	encoded := encode(text, codes)

	// 復号
	decoded := decode(encoded, root)

	// 結果を表示
	fmt.Printf("元の文字列   : %s\n", text)
	fmt.Printf("符号化後     : %s\n", encoded)
	fmt.Printf("復号後       : %s\n", decoded)
	fmt.Printf("復号は正しい? : %v\n", text == decoded)

	// --- 圧縮率を計算 ---
	originalBits := len(text) * 8 // ASCII: 1文字8ビット
	compressedBits := len(encoded)
	fmt.Println("--- 圧縮効果 ---")
	fmt.Printf("ASCII        : %dビット\n", originalBits)
	fmt.Printf("ハフマン     : %dビット\n", compressedBits)
	fmt.Printf("削減率       : %.1f%%\n",
			100*(1-float64(compressedBits)/float64(originalBits)))



}