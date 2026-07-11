package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Node はハフマン木の1つの節を表す
type Node struct {
	char  rune
	freq  int
	left  *Node
	right *Node
}

func (n *Node) isLeaf() bool {
	return n.left == nil && n.right == nil
}

// buildTree は頻度表から「いつも同じ」ハフマン木を組み立てる。
// 圧縮側と展開側で必ず同じ木になるよう、文字順にそろえて安定ソートを使う。
func buildTree(freq map[rune]int) *Node {
	// 文字を昇順に並べてから葉を作る（順番を決定的にする）
	var chars []rune
	for ch := range freq {
		chars = append(chars, ch)
	}
	sort.Slice(chars, func(i, j int) bool { return chars[i] < chars[j] })

	var nodes []*Node
	for _, ch := range chars {
		nodes = append(nodes, &Node{char: ch, freq: freq[ch]})
	}

	// 1番小さい2つを合体、を繰り返す
	for len(nodes) > 1 {
		sort.SliceStable(nodes, func(i, j int) bool {
			return nodes[i].freq < nodes[j].freq
		})
		left := nodes[0]
		right := nodes[1]
		parent := &Node{freq: left.freq + right.freq, left: left, right: right}
		nodes = append(nodes[2:], parent)
	}
	return nodes[0]
}

func buildCodes(node *Node, path string, codes map[rune]string) {
	if node == nil {
		return
	}
	if node.isLeaf() {
		codes[node.char] = path
		return
	}
	buildCodes(node.left, path+"0", codes)
	buildCodes(node.right, path+"1", codes)
}

func encode(text string, codes map[rune]string) string {
	var sb strings.Builder
	for _, ch := range text {
		sb.WriteString(codes[ch])
	}
	return sb.String()
}

func decode(bits string, root *Node) string {
	var sb strings.Builder
	node := root
	for _, bit := range bits {
		if bit == '0' {
			node = node.left
		} else {
			node = node.right
		}
		if node.isLeaf() {
			sb.WriteRune(node.char)
			node = root
		}
	}
	return sb.String()
}

func packBits(bits string) ([]byte, int) {
	padding := (8 - len(bits)%8) % 8
	bits = bits + strings.Repeat("0", padding)
	out := make([]byte, len(bits)/8)
	for i := 0; i < len(bits); i++ {
		if bits[i] == '1' {
			out[i/8] |= 1 << (7 - i%8)
		}
	}
	return out, padding
}

func unpackBits(data []byte, padding int) string {
	var sb strings.Builder
	for _, b := range data {
		for i := 7; i >= 0; i-- {
			if b&(1<<i) != 0 {
				sb.WriteByte('1')
			} else {
				sb.WriteByte('0')
			}
		}
	}
	s := sb.String()
	return s[:len(s)-padding]
}

// Archive は圧縮ファイルに保存する中身。木を作り直すための頻度表も一緒に持つ。
type Archive struct {
	Freq    map[rune]int
	Padding int
	Data    []byte
}

// compress はテキストをハフマン圧縮して path に書き出す
func compress(text, path string) error {
	freq := make(map[rune]int)
	for _, ch := range text {
		freq[ch]++
	}
	root := buildTree(freq)

	codes := make(map[rune]string)
	buildCodes(root, "", codes)

	encoded := encode(text, codes)
	packed, padding := packBits(encoded)

	arc := Archive{Freq: freq, Padding: padding, Data: packed}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return gob.NewEncoder(f).Encode(arc) // 頻度表・パディング・データをまとめて保存
}

// decompress は圧縮ファイル path を読み、元の文字列に復元する
func decompress(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var arc Archive
	if err := gob.NewDecoder(f).Decode(&arc); err != nil {
		return "", err
	}

	root := buildTree(arc.Freq) // 保存された頻度表から木を作り直す
	bits := unpackBits(arc.Data, arc.Padding)
	return decode(bits, root), nil
}

func main() {
	inPath := os.Args[1]
	outPath := inPath + ".huff"

	data, err := os.ReadFile(inPath)
	if err != nil {
		panic(err)
	}
	text := string(data)

	// 圧縮
	if err := compress(text, outPath); err != nil {
		panic(err)
	}

	// 圧縮ファイル「単体」から復元
	decoded, err := decompress(outPath)
	if err != nil {
		panic(err)
	}

	fmt.Printf("入力ファイル  : %s\n", inPath)
	fmt.Printf("圧縮ファイル  : %s\n", outPath)
	fmt.Printf("復号は正しい? : %v\n", text == decoded)

	info, _ := os.Stat(outPath)
	fmt.Println("--- 圧縮効果（実バイト・木込み）---")
	fmt.Printf("元ファイル    : %d バイト\n", len(data))
	fmt.Printf("圧縮ファイル  : %d バイト（木の情報を含む）\n", info.Size())
	fmt.Printf("削減率        : %.1f%%\n",
		100*(1-float64(info.Size())/float64(len(data))))
}
