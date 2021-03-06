package silverdile

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/morikuni/failure"
)

const MinResizeSize = 0
const MaxResizeSize = 2560

var sizeRegexp = regexp.MustCompile(`=s[\d]+`)

type ImageOption struct {
	Bucket             string
	Object             string
	Size               int
	CacheControlMaxAge int // Pathから指定はできず、App側で指定する
}

// BuildImageOption is Request URLからImageOptionを生成する
// App Engine Image Serviceと同じ雰囲気のURLを利用する時に使う
//
// 期待する形式
// `/{bucket}/{object}`
// `/{bucket}/{object}/=sXXX`
func BuildImageOption(path string) (*ImageOption, error) {
	var ret ImageOption

	blocks := strings.Split(path, "/")
	if len(blocks) < 3 {
		return nil, failure.New(InvalidArgument, failure.Messagef("Fewer expected blocks separated by `/`"))
	}
	ret.Bucket = blocks[1]
	ret.Object = blocks[2]

	// resize 指定がない場合は、そこで終わり
	if len(blocks) < 4 {
		return &ret, nil
	}

	l := sizeRegexp.FindAllStringSubmatch(path, -1)
	if len(l) > 0 {
		v := l[len(l)-1]
		vv := v[0]
		size, err := strconv.Atoi(vv[2:])
		if err != nil {
			return nil, err
		}
		if size < MinResizeSize || size > MaxResizeSize {
			return nil, failure.New(InvalidArgument, failure.Messagef("invalid resize arugment. size range is %d ~ %d, but got %d", MinResizeSize, MaxResizeSize, size))
		}
		ret.Size = size
		return &ret, nil
	}

	return nil, failure.New(InvalidArgument)
}
