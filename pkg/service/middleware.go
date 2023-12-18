package service

import (
	"github.com/go-kit/log"
	"github.com/kakurineuin/learn-english-word/pkg/model"
)

type loggingMiddleware struct {
	logger log.Logger
	next   WordService
}

func (mw loggingMiddleware) FindWordByDictionary(
	word string,
) (wordMeanings []model.WordMeaning, err error) {
	defer func() {
		mw.logger.Log("method", "FindWordByDictionary", "word", word, "err", err)
	}()
	return mw.next.FindWordByDictionary(word)
}
