package textbuffer

// Position 表示文本中的位置，由行号和列号组成
type Position struct {
	// Line 行号，从0开始
	Line int
	// Column 列号，从0开始
	Column int
}

// NewPosition 创建一个新的Position
func NewPosition(line, column int) Position {
	return Position{
		Line:   line,
		Column: column,
	}
}

// IsBeforeOrEqual 判断当前位置是否在另一个位置之前或相等
func (p Position) IsBeforeOrEqual(other Position) bool {
	if p.Line < other.Line {
		return true
	}
	if p.Line == other.Line && p.Column <= other.Column {
		return true
	}
	return false
}

// IsBefore 判断当前位置是否在另一个位置之前
func (p Position) IsBefore(other Position) bool {
	if p.Line < other.Line {
		return true
	}
	if p.Line == other.Line && p.Column < other.Column {
		return true
	}
	return false
}

// IsAfterOrEqual 判断当前位置是否在另一个位置之后或相等
func (p Position) IsAfterOrEqual(other Position) bool {
	return !p.IsBefore(other)
}

// IsAfter 判断当前位置是否在另一个位置之后
func (p Position) IsAfter(other Position) bool {
	return !p.IsBeforeOrEqual(other)
}

// Equals 判断两个位置是否相等
func (p Position) Equals(other Position) bool {
	return p.Line == other.Line && p.Column == other.Column
}

// Range 表示文本中的一个范围，由起始位置和结束位置组成
type Range struct {
	// Start 起始位置
	Start Position
	// End 结束位置
	End Position
}

// NewRange 创建一个新的Range
func NewRange(start, end Position) Range {
	return Range{
		Start: start,
		End:   end,
	}
}

// Contains 判断当前范围是否包含指定位置
func (r Range) Contains(position Position) bool {
	return position.IsAfterOrEqual(r.Start) && position.IsBeforeOrEqual(r.End)
}

// ContainsRange 判断当前范围是否包含另一个范围
func (r Range) ContainsRange(other Range) bool {
	return r.Contains(other.Start) && r.Contains(other.End)
}

// Overlaps 判断当前范围是否与另一个范围重叠
func (r Range) Overlaps(other Range) bool {
	return r.Contains(other.Start) || r.Contains(other.End) ||
		other.Contains(r.Start) || other.Contains(r.End)
}

// IsEmpty 判断当前范围是否为空
func (r Range) IsEmpty() bool {
	return r.Start.Equals(r.End)
}
