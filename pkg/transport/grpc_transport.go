package transport

import (
	"context"
	"errors"

	gt "github.com/go-kit/kit/transport/grpc"
	"github.com/go-kit/log"

	"github.com/kakurineuin/learn-english-word/pb"
	"github.com/kakurineuin/learn-english-word/pkg/endpoint"
)

type GRPCServer struct {
	findWordByDictionary gt.Handler
	pb.UnimplementedWordServiceServer
}

// NewGRPCServer initializes a new gRPC server
func NewGRPCServer(endpointds endpoint.Endpoints, logger log.Logger) pb.WordServiceServer {
	return &GRPCServer{
		findWordByDictionary: gt.NewServer(
			endpointds.FindWordByDictionary,
			decodeFindWordByDictionaryRequest,
			encodeFindWordByDictionaryResponse,
		),
	}
}

func (s GRPCServer) FindWordByDictionary(
	ctx context.Context,
	req *pb.FindWordByDictionaryRequest,
) (*pb.FindWordByDictionaryResponse, error) {
	_, resp, err := s.findWordByDictionary.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.(*pb.FindWordByDictionaryResponse), nil
}

func decodeFindWordByDictionaryRequest(_ context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(*pb.FindWordByDictionaryRequest)
	if !ok {
		return nil, errors.New("invalid request body")
	}

	return endpoint.FindWordByDictionaryRequest{Word: req.Word}, nil
}

func encodeFindWordByDictionaryResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp, ok := response.(endpoint.FindWordByDictionaryResponse)
	if !ok {
		return nil, errors.New("invalid response body")
	}

	pbWordMeanings := []*pb.WordMeaning{}

	for _, wm := range resp.WordMeanings {
		pbExamples := []*pb.Example{}

		for _, example := range wm.Examples {
			pbSentences := []*pb.Sentence{}

			for _, sentence := range example.Examples {
				pbSentences = append(pbSentences, &pb.Sentence{
					AudioUrl: sentence.AudioUrl,
					Text:     sentence.Text,
				})
			}

			pbExamples = append(pbExamples, &pb.Example{
				Pattern:  example.Pattern,
				Examples: pbSentences,
			})
		}

		pbWordMeaning := pb.WordMeaning{
			Id:           wm.Id.Hex(),
			Word:         wm.Word,
			PartOfSpeech: wm.PartOfSpeech,
			Gram:         wm.Gram,
			Pronunciation: &pb.Pronunciation{
				Text:       wm.Pronunciation.Text,
				UkAudioUrl: wm.Pronunciation.UkAudioUrl,
				UsAudioUrl: wm.Pronunciation.UsAudioUrl,
			},
			DefGram:               wm.DefGram,
			Definition:            wm.Definition,
			Examples:              pbExamples,
			OrderByNo:             wm.OrderByNo,
			QueryByWords:          wm.QueryByWords,
			FavoriteWordMeaningId: wm.FavoriteWordMeaningId.Hex(),
		}
		pbWordMeanings = append(pbWordMeanings, &pbWordMeaning)
	}

	return &pb.FindWordByDictionaryResponse{
		WordMeanings: pbWordMeanings,
	}, nil
}
