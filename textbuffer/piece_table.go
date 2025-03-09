package textbuffer

// PieceTable 是一种高效的文本编辑数据结构
// 它将文本分成多个片段（Piece），每个片段指向原始文本或添加的文本
type PieceTable struct {
	// 原始文本缓冲区
	original []rune
	// 添加的文本缓冲区
	added []rune
	// 片段数组
	pieces []Piece
}

// BufferType 表示片段指向的缓冲区类型
type BufferType int

const (
	// Original 表示原始缓冲区
	Original BufferType = iota
	// Added 表示添加缓冲区
	Added
)

// Piece 表示文本的一个片段
type Piece struct {
	// 缓冲区类型
	bufferType BufferType
	// 在缓冲区中的起始位置
	start int
	// 片段长度
	length int
}

// NewPieceTable 创建一个新的PieceTable
func NewPieceTable(text string) *PieceTable {
	runes := []rune(text)
	pt := &PieceTable{
		original: runes,
		added:    []rune{},
	}

	if len(runes) > 0 {
		pt.pieces = []Piece{
			{
				bufferType: Original,
				start:      0,
				length:     len(runes),
			},
		}
	} else {
		pt.pieces = []Piece{}
	}

	return pt
}

// Insert 在指定位置插入文本
func (pt *PieceTable) Insert(offset int, text string) {
	if len(text) == 0 {
		return
	}

	// 如果没有片段，直接添加一个新片段
	if len(pt.pieces) == 0 {
		addedIndex := len(pt.added)
		pt.added = append(pt.added, []rune(text)...)
		addedLength := len([]rune(text))

		pt.pieces = []Piece{
			{
				bufferType: Added,
				start:      addedIndex,
				length:     addedLength,
			},
		}
		return
	}

	// 找到插入位置所在的片段
	pieceIndex, pieceOffset := pt.findPiece(offset)

	// 添加新文本到added缓冲区
	addedIndex := len(pt.added)
	pt.added = append(pt.added, []rune(text)...)
	addedLength := len([]rune(text))

	// 创建新的片段数组
	newPieces := make([]Piece, 0, len(pt.pieces)+2)

	// 添加插入位置之前的片段
	newPieces = append(newPieces, pt.pieces[:pieceIndex]...)

	// 如果插入位置不在片段的开始，需要分割当前片段
	if pieceOffset > 0 {
		// 添加当前片段的前半部分
		newPieces = append(newPieces, Piece{
			bufferType: pt.pieces[pieceIndex].bufferType,
			start:      pt.pieces[pieceIndex].start,
			length:     pieceOffset,
		})
	}

	// 添加新插入的文本片段
	newPieces = append(newPieces, Piece{
		bufferType: Added,
		start:      addedIndex,
		length:     addedLength,
	})

	// 如果插入位置不在片段的结尾，需要添加当前片段的后半部分
	remainingLength := pt.pieces[pieceIndex].length - pieceOffset
	if remainingLength > 0 {
		newPieces = append(newPieces, Piece{
			bufferType: pt.pieces[pieceIndex].bufferType,
			start:      pt.pieces[pieceIndex].start + pieceOffset,
			length:     remainingLength,
		})
	}

	// 添加插入位置之后的片段
	if pieceIndex < len(pt.pieces)-1 {
		newPieces = append(newPieces, pt.pieces[pieceIndex+1:]...)
	}

	pt.pieces = newPieces
}

// Delete 删除指定范围的文本
func (pt *PieceTable) Delete(offset, length int) {
	if length <= 0 {
		return
	}

	// 如果没有片段，直接返回
	if len(pt.pieces) == 0 {
		return
	}

	// 找到删除范围的起始和结束位置
	startPieceIndex, startPieceOffset := pt.findPiece(offset)
	endPieceIndex, endPieceOffset := pt.findPiece(offset + length)

	// 创建新的片段数组
	newPieces := make([]Piece, 0, len(pt.pieces))

	// 添加删除范围之前的片段
	newPieces = append(newPieces, pt.pieces[:startPieceIndex]...)

	// 如果删除范围的起始位置不在片段的开始，需要保留当前片段的前半部分
	if startPieceOffset > 0 {
		newPieces = append(newPieces, Piece{
			bufferType: pt.pieces[startPieceIndex].bufferType,
			start:      pt.pieces[startPieceIndex].start,
			length:     startPieceOffset,
		})
	}

	// 如果删除范围的结束位置不在片段的结尾，需要保留当前片段的后半部分
	if endPieceOffset < pt.pieces[endPieceIndex].length {
		newPieces = append(newPieces, Piece{
			bufferType: pt.pieces[endPieceIndex].bufferType,
			start:      pt.pieces[endPieceIndex].start + endPieceOffset,
			length:     pt.pieces[endPieceIndex].length - endPieceOffset,
		})
	}

	// 添加删除范围之后的片段
	if endPieceIndex < len(pt.pieces)-1 {
		newPieces = append(newPieces, pt.pieces[endPieceIndex+1:]...)
	}

	pt.pieces = newPieces
}

// GetText 获取整个文本内容
func (pt *PieceTable) GetText() string {
	var result []rune
	for _, piece := range pt.pieces {
		var buffer []rune
		if piece.bufferType == Original {
			buffer = pt.original
		} else {
			buffer = pt.added
		}
		result = append(result, buffer[piece.start:piece.start+piece.length]...)
	}
	return string(result)
}

// GetLength 获取文本总长度
func (pt *PieceTable) GetLength() int {
	length := 0
	for _, piece := range pt.pieces {
		length += piece.length
	}
	return length
}

// findPiece 找到指定偏移量所在的片段和在片段内的偏移量
func (pt *PieceTable) findPiece(offset int) (pieceIndex, pieceOffset int) {
	currentOffset := 0
	for i, piece := range pt.pieces {
		if currentOffset <= offset && offset < currentOffset+piece.length {
			return i, offset - currentOffset
		}
		currentOffset += piece.length
	}
	// 如果偏移量超出文本长度，返回最后一个片段的结尾
	if len(pt.pieces) > 0 {
		return len(pt.pieces) - 1, pt.pieces[len(pt.pieces)-1].length
	}
	// 如果没有片段，返回0, 0
	return 0, 0
}
