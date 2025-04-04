package db

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	maxPos          byte          = 63
	stateExpiration time.Duration = 0
)

var (
	ErrPosOutOfRange = errors.New("position out of range 0-63")
	ErrInvalidState  = errors.New("state is not an 8 byte bitmap")
	ErrNoStateInDb   = errors.New("no state in db")
)

var ZeroState [8]byte

func SetBit(db *redis.Client, ctx context.Context, stateKey string, pos uint8, set bool) error {
	if pos > maxPos {
		return ErrPosOutOfRange
	}

	var val int
	if set {
		val = 1
	}

	setErr := db.SetBit(ctx, stateKey, int64(pos), val).Err()
	if setErr == redis.Nil {
		return ErrNoStateInDb
	}
	if setErr != nil {
		return setErr
	}

	return nil
}

func GetState(db *redis.Client, ctx context.Context, stateKey string) ([8]byte, error) {
	var bitmap [8]byte

	s, getErr := db.Get(ctx, stateKey).Bytes()
	if getErr == redis.Nil {
		return bitmap, ErrNoStateInDb
	}
	if getErr != nil {
		return bitmap, getErr
	}

	if len(s) != 8 {
		return bitmap, ErrInvalidState
	}

	copy(bitmap[:], s)
	return bitmap, nil
}

func SetState(db *redis.Client, ctx context.Context, stateKey string, state [8]byte) error {
	err := db.Set(ctx, stateKey, state, stateExpiration).Err()
	if err != nil {
		return err
	}

	return nil
}
