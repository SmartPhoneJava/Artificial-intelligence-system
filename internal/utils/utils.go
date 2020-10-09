package utils

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"
)

var Err429 = errors.New("429 Too Many Requests")

func MakeAction(
	ctx context.Context,
	action func() error,
) error {

	var count = 0
	for {
		var timer = time.Millisecond * 100
		if count > 5 {
			count = 0
			timer = time.Second * 10
			log.Printf("RPM limit: 10s sleep")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(timer):
			err := action()
			if err == Err429 {
				count++
				continue
			}
			return err
		}
	}
}

func String(n int32) string {
	buf := [11]byte{}
	pos := len(buf)
	i := int64(n)
	signed := i < 0
	if signed {
		i = -i
	}
	for {
		pos--
		buf[pos], i = '0'+byte(i%10), i/10
		if i == 0 {
			if signed {
				pos--
				buf[pos] = '-'
			}
			return string(buf[pos:])
		}
	}
}

func InvertArr(arr []string) []string {
	var inverted = make([]string, len(arr))
	for i := range arr {
		inverted[len(inverted)-i-1] = arr[i]
	}
	return inverted
}

func RequestInt(r *http.Request, key string) (int, error) {
	return strconv.Atoi(r.URL.Query().Get(key))
}
