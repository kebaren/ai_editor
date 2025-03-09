package textbuffer

// 红黑树的颜色
type Color bool

const (
	// 红色节点
	RED Color = true
	// 黑色节点
	BLACK Color = false
)

// TreeNode 表示红黑树的节点
type TreeNode struct {
	// 节点颜色
	color Color
	// 左子节点
	left *TreeNode
	// 右子节点
	right *TreeNode
	// 父节点
	parent *TreeNode
	// 节点大小（包括子树中的所有节点）
	size int
	// 节点高度（从该节点到叶子节点的最长路径）
	height int
	// 节点包含的行长度
	length int
	// 节点包含的行内容
	line string
	// 该节点及其子树中所有行的总长度
	totalLength int
	// 该节点及其子树中的换行符数量
	lineBreakCount int
	// 该节点是否包含换行符
	containsLineBreak bool
}

// RBTree 表示红黑树
type RBTree struct {
	// 根节点
	root *TreeNode
	// 树中的节点数量
	size int
}

// NewRBTree 创建一个新的红黑树
func NewRBTree() *RBTree {
	return &RBTree{
		root: nil,
		size: 0,
	}
}

// Size 返回树中的节点数量
func (t *RBTree) Size() int {
	return t.size
}

// IsEmpty 判断树是否为空
func (t *RBTree) IsEmpty() bool {
	return t.size == 0
}

// Clear 清空树
func (t *RBTree) Clear() {
	t.root = nil
	t.size = 0
}

// GetRoot 获取根节点
func (t *RBTree) GetRoot() *TreeNode {
	return t.root
}

// Insert 在指定位置插入文本
func (t *RBTree) Insert(offset int, text string) {
	if text == "" {
		return
	}

	if t.root == nil {
		// 如果树为空，创建根节点
		t.root = &TreeNode{
			color:             BLACK,
			left:              nil,
			right:             nil,
			parent:            nil,
			size:              1,
			height:            1,
			length:            len([]rune(text)),
			line:              text,
			totalLength:       len([]rune(text)),
			lineBreakCount:    countLineBreaks(text),
			containsLineBreak: containsLineBreak(text),
		}
		t.size = 1
		return
	}

	// 找到插入位置
	node, nodeOffset := t.findNodeAtOffset(offset)
	if node == nil {
		return
	}

	// 更新节点内容
	runes := []rune(node.line)
	newRunes := []rune(text)

	// 在节点内的指定位置插入文本
	newLine := string(append(append([]rune{}, runes[:nodeOffset]...), append(newRunes, runes[nodeOffset:]...)...))

	// 检查是否需要分割节点（如果插入的文本包含换行符或原节点包含换行符）
	if containsLineBreak(text) || containsLineBreak(newLine) {
		lines := splitLines(newLine)
		if len(lines) > 1 {
			// 删除原节点
			t.deleteNode(node)

			// 插入分割后的多个节点
			for _, line := range lines {
				t.insertLine(line)
			}
			return
		}
	}

	// 更新节点属性
	oldLength := node.length
	node.line = newLine
	node.length = len([]rune(newLine))
	node.totalLength = node.totalLength - oldLength + node.length
	node.lineBreakCount = countLineBreaks(newLine)
	node.containsLineBreak = containsLineBreak(newLine)

	// 更新祖先节点的属性
	t.updateAncestors(node)
}

// Delete 删除指定范围的文本
func (t *RBTree) Delete(startOffset, endOffset int) {
	if startOffset >= endOffset {
		return
	}

	// 找到起始位置
	startNode, startNodeOffset := t.findNodeAtOffset(startOffset)
	if startNode == nil {
		return
	}

	// 找到结束位置
	endNode, endNodeOffset := t.findNodeAtOffset(endOffset)
	if endNode == nil {
		return
	}

	// 如果起始和结束位置在同一个节点内
	if startNode == endNode {
		runes := []rune(startNode.line)
		newLine := string(append([]rune{}, append(runes[:startNodeOffset], runes[endNodeOffset:]...)...))

		// 更新节点属性
		oldLength := startNode.length
		startNode.line = newLine
		startNode.length = len([]rune(newLine))
		startNode.totalLength = startNode.totalLength - oldLength + startNode.length
		startNode.lineBreakCount = countLineBreaks(newLine)
		startNode.containsLineBreak = containsLineBreak(newLine)

		// 更新祖先节点的属性
		t.updateAncestors(startNode)
		return
	}

	// 如果起始和结束位置在不同的节点内
	// 1. 处理起始节点
	startRunes := []rune(startNode.line)
	startNewLine := string(startRunes[:startNodeOffset])

	// 2. 处理结束节点
	endRunes := []rune(endNode.line)
	endNewLine := string(endRunes[endNodeOffset:])

	// 3. 合并起始和结束节点
	mergedLine := startNewLine + endNewLine

	// 4. 删除中间的所有节点
	nodesToDelete := t.getNodesBetween(startNode, endNode)
	for _, node := range nodesToDelete {
		t.deleteNode(node)
	}

	// 5. 插入合并后的节点
	if mergedLine != "" {
		t.insertLine(mergedLine)
	}
}

// GetText 获取整个文本内容
func (t *RBTree) GetText() string {
	if t.root == nil {
		return ""
	}

	var result string
	t.inOrderTraversal(t.root, func(node *TreeNode) {
		result += node.line
	})

	return result
}

// GetLineCount 获取行数
func (t *RBTree) GetLineCount() int {
	if t.root == nil {
		return 0
	}

	return t.root.lineBreakCount + 1
}

// GetLineContent 获取指定行的内容
func (t *RBTree) GetLineContent(lineIndex int) string {
	if lineIndex < 0 || lineIndex >= t.GetLineCount() {
		return ""
	}

	node, offset := t.findNodeAtLine(lineIndex)
	if node == nil {
		return ""
	}

	lines := splitLines(node.line)
	if offset < len(lines) {
		return lines[offset]
	}

	return ""
}

// GetLines 获取所有行的内容
func (t *RBTree) GetLines() []string {
	text := t.GetText()
	return splitLines(text)
}

// GetPositionAt 获取指定偏移量对应的位置
func (t *RBTree) GetPositionAt(offset int) Position {
	if t.root == nil {
		return Position{Line: 0, Column: 0}
	}

	if offset <= 0 {
		return Position{Line: 0, Column: 0}
	}

	lineCount := 0
	columnCount := 0
	currentOffset := 0

	t.inOrderTraversal(t.root, func(node *TreeNode) {
		if currentOffset >= offset {
			return
		}

		runes := []rune(node.line)
		for i, r := range runes {
			if currentOffset+1 > offset {
				return
			}

			currentOffset++

			if r == '\n' {
				lineCount++
				columnCount = 0
			} else {
				columnCount++
			}
		}
	})

	return Position{Line: lineCount, Column: columnCount}
}

// GetOffsetAt 获取指定位置对应的偏移量
func (t *RBTree) GetOffsetAt(position Position) int {
	if t.root == nil {
		return 0
	}

	if position.Line < 0 {
		return 0
	}

	lineCount := 0
	offset := 0

	t.inOrderTraversal(t.root, func(node *TreeNode) {
		if lineCount > position.Line {
			return
		}

		runes := []rune(node.line)
		for _, r := range runes {
			if lineCount == position.Line && offset >= position.Column {
				return
			}

			offset++

			if r == '\n' {
				lineCount++
				if lineCount > position.Line {
					return
				}
			}
		}
	})

	return offset
}

// 辅助方法

// findNodeAtOffset 找到指定偏移量所在的节点和在节点内的偏移量
func (t *RBTree) findNodeAtOffset(offset int) (*TreeNode, int) {
	if t.root == nil {
		return nil, 0
	}

	currentOffset := 0
	var result *TreeNode
	var nodeOffset int

	t.inOrderTraversal(t.root, func(node *TreeNode) {
		if result != nil {
			return
		}

		if currentOffset <= offset && offset < currentOffset+node.length {
			result = node
			nodeOffset = offset - currentOffset
			return
		}

		currentOffset += node.length
	})

	if result == nil && t.size > 0 {
		// 如果偏移量超出文本长度，返回最后一个节点
		lastNode := t.getLastNode()
		return lastNode, lastNode.length
	}

	return result, nodeOffset
}

// findNodeAtLine 找到指定行所在的节点和在节点内的行偏移量
func (t *RBTree) findNodeAtLine(lineIndex int) (*TreeNode, int) {
	if t.root == nil {
		return nil, 0
	}

	currentLine := 0
	var result *TreeNode
	var lineOffset int

	t.inOrderTraversal(t.root, func(node *TreeNode) {
		if result != nil {
			return
		}

		lineBreaks := countLineBreaks(node.line)
		if currentLine <= lineIndex && lineIndex < currentLine+lineBreaks+1 {
			result = node
			lineOffset = lineIndex - currentLine
			return
		}

		currentLine += lineBreaks + 1
	})

	return result, lineOffset
}

// getNodesBetween 获取两个节点之间的所有节点（包括这两个节点）
func (t *RBTree) getNodesBetween(startNode, endNode *TreeNode) []*TreeNode {
	var nodes []*TreeNode
	inRange := false

	t.inOrderTraversal(t.root, func(node *TreeNode) {
		if node == startNode {
			inRange = true
		}

		if inRange {
			nodes = append(nodes, node)
		}

		if node == endNode {
			inRange = false
		}
	})

	return nodes
}

// getLastNode 获取树中的最后一个节点
func (t *RBTree) getLastNode() *TreeNode {
	if t.root == nil {
		return nil
	}

	current := t.root
	for current.right != nil {
		current = current.right
	}

	return current
}

// inOrderTraversal 中序遍历树
func (t *RBTree) inOrderTraversal(node *TreeNode, visit func(*TreeNode)) {
	if node == nil {
		return
	}

	t.inOrderTraversal(node.left, visit)
	visit(node)
	t.inOrderTraversal(node.right, visit)
}

// insertLine 插入一行文本
func (t *RBTree) insertLine(line string) {
	newNode := &TreeNode{
		color:             RED,
		left:              nil,
		right:             nil,
		parent:            nil,
		size:              1,
		height:            1,
		length:            len([]rune(line)),
		line:              line,
		totalLength:       len([]rune(line)),
		lineBreakCount:    countLineBreaks(line),
		containsLineBreak: containsLineBreak(line),
	}

	if t.root == nil {
		newNode.color = BLACK
		t.root = newNode
		t.size = 1
		return
	}

	// 简化插入逻辑，直接添加到末尾
	current := t.root
	for current.right != nil {
		current = current.right
	}

	newNode.parent = current
	current.right = newNode

	// 修复红黑树性质
	t.fixAfterInsertion(newNode)

	// 更新树的大小
	t.size++

	// 更新祖先节点的属性
	t.updateAncestors(newNode)
}

// deleteNode 删除节点
func (t *RBTree) deleteNode(node *TreeNode) {
	if node == nil {
		return
	}

	// 如果节点有两个子节点
	if node.left != nil && node.right != nil {
		// 找到后继节点
		successor := t.getSuccessor(node)

		// 交换节点内容
		node.line = successor.line
		node.length = successor.length
		node.totalLength = successor.totalLength
		node.lineBreakCount = successor.lineBreakCount
		node.containsLineBreak = successor.containsLineBreak

		// 删除后继节点
		t.deleteNode(successor)
		return
	}

	// 如果节点有一个子节点或没有子节点
	var child *TreeNode
	if node.left != nil {
		child = node.left
	} else {
		child = node.right
	}

	// 如果节点是根节点
	if node.parent == nil {
		t.root = child
		if child != nil {
			child.parent = nil
			child.color = BLACK
		}
		t.size--
		return
	}

	// 如果节点不是根节点
	if node.parent.left == node {
		node.parent.left = child
	} else {
		node.parent.right = child
	}

	if child != nil {
		child.parent = node.parent
	}

	// 如果删除的是黑色节点，需要修复红黑树性质
	if node.color == BLACK {
		t.fixAfterDeletion(child, node.parent)
	}

	// 更新树的大小
	t.size--

	// 更新祖先节点的属性
	t.updateAncestors(node.parent)
}

// getSuccessor 获取节点的后继节点
func (t *RBTree) getSuccessor(node *TreeNode) *TreeNode {
	if node == nil {
		return nil
	}

	if node.right != nil {
		// 如果有右子树，后继节点是右子树中的最小节点
		current := node.right
		for current.left != nil {
			current = current.left
		}
		return current
	}

	// 如果没有右子树，后继节点是祖先节点中第一个将当前节点作为左子树的节点
	current := node
	parent := node.parent
	for parent != nil && current == parent.right {
		current = parent
		parent = parent.parent
	}

	return parent
}

// fixAfterInsertion 插入节点后修复红黑树性质
func (t *RBTree) fixAfterInsertion(node *TreeNode) {
	if node == nil {
		return
	}

	// 如果是根节点，直接设为黑色
	if node.parent == nil {
		node.color = BLACK
		return
	}

	// 如果父节点是黑色，不需要修复
	if node.parent.color == BLACK {
		return
	}

	// 获取叔叔节点
	uncle := t.getUncle(node)

	// 如果叔叔节点是红色
	if uncle != nil && uncle.color == RED {
		// 将父节点和叔叔节点设为黑色
		node.parent.color = BLACK
		uncle.color = BLACK

		// 将祖父节点设为红色
		grandparent := t.getGrandparent(node)
		if grandparent != nil {
			grandparent.color = RED

			// 递归修复祖父节点
			t.fixAfterInsertion(grandparent)
		}

		return
	}

	// 如果叔叔节点是黑色或不存在
	grandparent := t.getGrandparent(node)
	if grandparent == nil {
		return
	}

	// 如果是"折线"形状，先旋转成"直线"形状
	if node.parent == grandparent.left && node == node.parent.right {
		t.rotateLeft(node.parent)
		node = node.left
	} else if node.parent == grandparent.right && node == node.parent.left {
		t.rotateRight(node.parent)
		node = node.right
	}

	// 如果是"直线"形状
	node.parent.color = BLACK
	grandparent.color = RED

	if node == node.parent.left {
		t.rotateRight(grandparent)
	} else {
		t.rotateLeft(grandparent)
	}
}

// fixAfterDeletion 删除节点后修复红黑树性质
func (t *RBTree) fixAfterDeletion(node *TreeNode, parent *TreeNode) {
	if node != nil && node.color == RED {
		// 如果替代节点是红色，直接设为黑色
		node.color = BLACK
		return
	}

	if parent == nil {
		// 如果是根节点，不需要修复
		return
	}

	// 获取兄弟节点
	var sibling *TreeNode
	if node == parent.left {
		sibling = parent.right
	} else {
		sibling = parent.left
	}

	// 如果兄弟节点是红色
	if sibling != nil && sibling.color == RED {
		sibling.color = BLACK
		parent.color = RED

		if node == parent.left {
			t.rotateLeft(parent)
		} else {
			t.rotateRight(parent)
		}

		// 更新兄弟节点
		if node == parent.left {
			sibling = parent.right
		} else {
			sibling = parent.left
		}
	}

	// 如果兄弟节点的两个子节点都是黑色
	if (sibling.left == nil || sibling.left.color == BLACK) &&
		(sibling.right == nil || sibling.right.color == BLACK) {
		sibling.color = RED

		if parent.color == RED {
			parent.color = BLACK
		} else {
			t.fixAfterDeletion(parent, parent.parent)
		}

		return
	}

	// 如果兄弟节点的子节点有红色
	if node == parent.left {
		if sibling.right == nil || sibling.right.color == BLACK {
			if sibling.left != nil {
				sibling.left.color = BLACK
			}
			sibling.color = RED
			t.rotateRight(sibling)
			sibling = parent.right
		}

		sibling.color = parent.color
		parent.color = BLACK
		if sibling.right != nil {
			sibling.right.color = BLACK
		}
		t.rotateLeft(parent)
	} else {
		if sibling.left == nil || sibling.left.color == BLACK {
			if sibling.right != nil {
				sibling.right.color = BLACK
			}
			sibling.color = RED
			t.rotateLeft(sibling)
			sibling = parent.left
		}

		sibling.color = parent.color
		parent.color = BLACK
		if sibling.left != nil {
			sibling.left.color = BLACK
		}
		t.rotateRight(parent)
	}
}

// rotateLeft 左旋转
func (t *RBTree) rotateLeft(node *TreeNode) {
	if node == nil || node.right == nil {
		return
	}

	right := node.right

	// 更新右子节点
	node.right = right.left
	if right.left != nil {
		right.left.parent = node
	}

	// 更新父节点
	right.parent = node.parent
	if node.parent == nil {
		t.root = right
	} else if node == node.parent.left {
		node.parent.left = right
	} else {
		node.parent.right = right
	}

	// 更新左子节点
	right.left = node
	node.parent = right

	// 更新节点属性
	t.updateNodeProperties(node)
	t.updateNodeProperties(right)
}

// rotateRight 右旋转
func (t *RBTree) rotateRight(node *TreeNode) {
	if node == nil || node.left == nil {
		return
	}

	left := node.left

	// 更新左子节点
	node.left = left.right
	if left.right != nil {
		left.right.parent = node
	}

	// 更新父节点
	left.parent = node.parent
	if node.parent == nil {
		t.root = left
	} else if node == node.parent.left {
		node.parent.left = left
	} else {
		node.parent.right = left
	}

	// 更新右子节点
	left.right = node
	node.parent = left

	// 更新节点属性
	t.updateNodeProperties(node)
	t.updateNodeProperties(left)
}

// getGrandparent 获取祖父节点
func (t *RBTree) getGrandparent(node *TreeNode) *TreeNode {
	if node == nil || node.parent == nil {
		return nil
	}

	return node.parent.parent
}

// getUncle 获取叔叔节点
func (t *RBTree) getUncle(node *TreeNode) *TreeNode {
	grandparent := t.getGrandparent(node)
	if grandparent == nil {
		return nil
	}

	if node.parent == grandparent.left {
		return grandparent.right
	} else {
		return grandparent.left
	}
}

// updateNodeProperties 更新节点属性
func (t *RBTree) updateNodeProperties(node *TreeNode) {
	if node == nil {
		return
	}

	// 更新大小
	node.size = 1
	if node.left != nil {
		node.size += node.left.size
	}
	if node.right != nil {
		node.size += node.right.size
	}

	// 更新高度
	leftHeight := 0
	if node.left != nil {
		leftHeight = node.left.height
	}

	rightHeight := 0
	if node.right != nil {
		rightHeight = node.right.height
	}

	node.height = 1 + max(leftHeight, rightHeight)

	// 更新总长度
	node.totalLength = node.length
	if node.left != nil {
		node.totalLength += node.left.totalLength
	}
	if node.right != nil {
		node.totalLength += node.right.totalLength
	}

	// 更新换行符数量
	node.lineBreakCount = countLineBreaks(node.line)
	if node.left != nil {
		node.lineBreakCount += node.left.lineBreakCount
	}
	if node.right != nil {
		node.lineBreakCount += node.right.lineBreakCount
	}
}

// updateAncestors 更新祖先节点的属性
func (t *RBTree) updateAncestors(node *TreeNode) {
	current := node
	for current != nil {
		t.updateNodeProperties(current)
		current = current.parent
	}
}

// 工具函数

// countLineBreaks 计算字符串中的换行符数量
func countLineBreaks(s string) int {
	count := 0
	for _, r := range s {
		if r == '\n' {
			count++
		}
	}
	return count
}

// containsLineBreak 判断字符串是否包含换行符
func containsLineBreak(s string) bool {
	for _, r := range s {
		if r == '\n' {
			return true
		}
	}
	return false
}

// splitLines 将字符串分割成行
func splitLines(s string) []string {
	var lines []string
	var currentLine string

	for _, r := range s {
		if r == '\n' {
			lines = append(lines, currentLine+string(r))
			currentLine = ""
		} else {
			currentLine += string(r)
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
