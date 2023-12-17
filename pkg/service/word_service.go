package service

import (
	"fmt"

	"github.com/kakurineuin/learn-english-word/pkg/model"
)

type WordService struct{}

func (wordService *WordService) FindWordByDictionary(
	word string,
) (*[]model.WordMeaning, error) {
	fmt.Println("TODO............. test ok!")
	// TODO: 待實做
	return &[]model.WordMeaning{}, nil
}
