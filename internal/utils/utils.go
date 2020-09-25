package utils

import (
	"context"
	"errors"
	"log"
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
			timer = time.Minute
			log.Printf("RPM limit: Minute sleep")
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
