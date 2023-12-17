package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	"github.com/kakurineuin/learn-english-word/pkg/model"
	"github.com/kakurineuin/learn-english-word/pkg/service"
)

type Endpoints struct {
	FindWordByDictionary endpoint.Endpoint
}

type FindWordByDictionaryRequest struct {
	Word string
}

type FindWordByDictionaryResponse struct {
	WordMeanings *[]model.WordMeaning
	Err          error
}

// MakeAddEndpoint struct holds the endpoint response definition
func MakeEndpoints(wordService service.WordService) Endpoints {
	return Endpoints{
		FindWordByDictionary: makeFindWordByDictionaryEndpoint(wordService),
	}
}

func makeFindWordByDictionaryEndpoint(wordService service.WordService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(FindWordByDictionaryRequest)
		wordMeangins, err := wordService.FindWordByDictionary(req.Word)
		if err != nil {
			return FindWordByDictionaryResponse{wordMeangins, err}, nil
		}
		return FindWordByDictionaryResponse{wordMeangins, nil}, nil
	}
}
