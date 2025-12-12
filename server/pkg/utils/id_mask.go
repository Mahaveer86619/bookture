package utils

import (
	"fmt"

	"github.com/Mahaveer86619/bookture/server/pkg/config"
	"github.com/speps/go-hashids/v2"
)

const (
	minLength = 8
	alphabet  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
)

var hd *hashids.HashIDData

func init() {
	hd = hashids.NewData()
	hd.Salt = config.AppConfig.ID_SALT
	hd.MinLength = minLength
	hd.Alphabet = alphabet
}

func MaskID(id uint) string {
	h, _ := hashids.NewWithData(hd)
	encoded, _ := h.Encode([]int{int(id)})
	return encoded
}

func UnmaskID(hash string) (uint, error) {
	h, _ := hashids.NewWithData(hd)
	decoding, err := h.DecodeWithError(hash)
	if err != nil {
		return 0, err
	}
	if len(decoding) == 0 {
		return 0, fmt.Errorf("invalid id")
	}
	return uint(decoding[0]), nil
}
