package service

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/kakurineuin/learn-english-word/pkg/model"
)

type WordService interface {
	FindWordByDictionary(word string) (*[]model.WordMeaning, error)
}

type wordService struct{}

func New(logger log.Logger) WordService {
	var wordWervice WordService = &wordService{}
	wordWervice = loggingMiddleware{logger, wordWervice}
	return wordWervice
}

func (wordService *wordService) FindWordByDictionary(
	word string,
) (*[]model.WordMeaning, error) {
	fmt.Println("TODO............. test ok!")
	// TODO: 待實做
	return &[]model.WordMeaning{}, nil
}
