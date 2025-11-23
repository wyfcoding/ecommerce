package algorithm

import (
	"sync"
)

// ============================================================================
// 6. 分治算法 - 大规模数据处理
// ============================================================================

// DivideAndConquerProcessor 分治处理器
type DivideAndConquerProcessor struct {
	data []int64
	mu   sync.RWMutex
}

// NewDivideAndConquerProcessor 创建分治处理器
func NewDivideAndConquerProcessor(data []int64) *DivideAndConquerProcessor {
	return &DivideAndConquerProcessor{
		data: data,
	}
}

// MergeSort 归并排序（分治）
// 应用: 大规模订单排序、库存排序
func (dcp *DivideAndConquerProcessor) MergeSort() []int64 {
	dcp.mu.Lock()
	defer dcp.mu.Unlock()

	if len(dcp.data) <= 1 {
		return dcp.data
	}

	result := make([]int64, len(dcp.data))
	copy(result, dcp.data)
	dcp.mergeSort(result, 0, len(result)-1)
	return result
}

// mergeSort 递归排序
func (dcp *DivideAndConquerProcessor) mergeSort(arr []int64, left, right int) {
	if left < right {
		mid := (left + right) / 2
		dcp.mergeSort(arr, left, mid)
		dcp.mergeSort(arr, mid+1, right)
		dcp.merge(arr, left, mid, right)
	}
}

// merge 合并两个有序数组
func (dcp *DivideAndConquerProcessor) merge(arr []int64, left, mid, right int) {
	leftArr := make([]int64, mid-left+1)
	rightArr := make([]int64, right-mid)

	copy(leftArr, arr[left:mid+1])
	copy(rightArr, arr[mid+1:right+1])

	i, j, k := 0, 0, left

	for i < len(leftArr) && j < len(rightArr) {
		if leftArr[i] <= rightArr[j] {
			arr[k] = leftArr[i]
			i++
		} else {
			arr[k] = rightArr[j]
			j++
		}
		k++
	}

	for i < len(leftArr) {
		arr[k] = leftArr[i]
		i++
		k++
	}

	for j < len(rightArr) {
		arr[k] = rightArr[j]
		j++
		k++
	}
}

// CountInversions 计算逆序对（分治）
// 应用: 分析订单异常度
func (dcp *DivideAndConquerProcessor) CountInversions() int64 {
	dcp.mu.Lock()
	defer dcp.mu.Unlock()

	if len(dcp.data) <= 1 {
		return 0
	}

	arr := make([]int64, len(dcp.data))
	copy(arr, dcp.data)
	_, count := dcp.mergeSortCount(arr, 0, len(arr)-1)
	return count
}

// mergeSortCount 计算逆序对
func (dcp *DivideAndConquerProcessor) mergeSortCount(arr []int64, left, right int) ([]int64, int64) {
	if left >= right {
		return arr, 0
	}

	mid := (left + right) / 2
	_, leftCount := dcp.mergeSortCount(arr, left, mid)
	_, rightCount := dcp.mergeSortCount(arr, mid+1, right)

	mergeCount := dcp.mergeCount(arr, left, mid, right)

	return arr, leftCount + rightCount + mergeCount
}

// mergeCount 合并时计算逆序对
func (dcp *DivideAndConquerProcessor) mergeCount(arr []int64, left, mid, right int) int64 {
	leftArr := make([]int64, mid-left+1)
	rightArr := make([]int64, right-mid)

	copy(leftArr, arr[left:mid+1])
	copy(rightArr, arr[mid+1:right+1])

	i, j, k := 0, 0, left
	var count int64 = 0

	for i < len(leftArr) && j < len(rightArr) {
		if leftArr[i] <= rightArr[j] {
			arr[k] = leftArr[i]
			i++
		} else {
			arr[k] = rightArr[j]
			count += int64(len(leftArr) - i)
			j++
		}
		k++
	}

	for i < len(leftArr) {
		arr[k] = leftArr[i]
		i++
		k++
	}

	for j < len(rightArr) {
		arr[k] = rightArr[j]
		j++
		k++
	}

	return count
}
