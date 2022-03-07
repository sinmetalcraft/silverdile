package silverdile

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const MinResizeSize = 0
const MaxResizeSize = 2560

var sizeRegexp = regexp.MustCompile(`=s[\d]+`)

type ResizeInfo struct {
	Bucket string
	Object string
	Size   int
}

// BuildResizeInfo is Request URLからResizeに必要な情報を生成する
// App Engine Image Serviceと同じ雰囲気のURLを利用する時に使う
//
// 期待する形式
// `/{bucket}/{object}`
// `/{bucket}/{object}/=sXXX`
func BuildResizeInfo(path string) (*ResizeInfo, error) {
	ret := &ResizeInfo{}

	blocks := strings.Split(path, "/")
	if len(blocks) < 3 {
		return nil, NewErrInvalidArgument("Fewer expected blocks separated by `/`", map[string]interface{}{}, nil)
	}
	ret.Bucket = blocks[1]
	blocks = blocks[2:len(blocks)]

	size, err := buildSize(path)
	if err != nil {
		return nil, err
	}
	if size > 0 {
		ret.Size = size
		blocks = blocks[:len(blocks)-1] // 末尾は Size で消費されたので残りの部分が Object Path
	}
	ret.Object = strings.Join(blocks, "/")

	if len(ret.Bucket) < 1 || len(ret.Object) < 1 {
		return nil, NewErrInvalidArgument("invalid path", map[string]interface{}{"path": path}, nil)
	}

	return ret, nil
}

func buildSize(path string) (int, error) {
	l := sizeRegexp.FindAllStringSubmatch(path, -1)
	if len(l) < 1 {
		return 0, nil
	}

	v := l[len(l)-1]
	vv := v[0]
	size, err := strconv.Atoi(vv[2:])
	if err != nil {
		return 0, NewErrInvalidArgument(
			fmt.Sprintf("invalid resize arugment. size range is %d ~ %d, but got %d", MinResizeSize, MaxResizeSize, size),
			map[string]interface{}{}, err)
	}
	if size < MinResizeSize || size > MaxResizeSize {
		return 0, NewErrInvalidArgument(
			fmt.Sprintf("invalid resize arugment. size range is %d ~ %d, but got %d", MinResizeSize, MaxResizeSize, size),
			map[string]interface{}{}, nil)
	}
	return size, nil
}
